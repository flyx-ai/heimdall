package providers

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

var openRouterBaseURL = "https://openrouter.ai/api/v1"

type openRouterRequest struct {
	Model          string         `json:"model"`
	Messages       any            `json:"messages"`
	Stream         bool           `json:"stream"`
	StreamOptions  streamOptions  `json:"stream_options"`
	Temperature    float32        `json:"temperature,omitempty"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`
}

type openRouterChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenRouter struct {
	apiKeys []string
}

func NewOpenRouter(apiKeys []string) OpenRouter {
	return OpenRouter{
		apiKeys: apiKeys,
	}
}

func (or OpenRouter) Name() string {
	return models.OpenRouterProvider
}

func (or OpenRouter) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	model, ok := req.Model.(models.OpenRouterModel)
	if !ok {
		return response.Completion{}, 0, errors.New("model must be OpenRouterModel")
	}

	openRouterReq := openRouterRequest{
		Model:         model.ModelName,
		Stream:        true,
		StreamOptions: streamOptions{IncludeUsage: true},
		Temperature:   1.0,
	}

	if len(model.StructuredOutput) > 0 {
		openRouterReq.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": model.StructuredOutput,
		}
	}

	preparedReq, err := prepareOpenRouterRequest(
		openRouterReq,
		model,
		req.SystemMessage,
		req.UserMessage,
		req.History,
	)
	if err != nil {
		return response.Completion{}, 0, err
	}

	body, err := json.Marshal(preparedReq)
	if err != nil {
		return response.Completion{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/chat/completions", openRouterBaseURL),
		bytes.NewReader(body))
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Completion{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"received status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder
	var usage response.Usage
	chunks := 0
	now := time.Now()

	for {
		if chunks == 0 && time.Since(now).Seconds() > 3.0 {
			return response.Completion{}, 0, context.Canceled
		}
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return response.Completion{}, 0, fmt.Errorf("read line: %w", err)
		}

		line = strings.TrimPrefix(line, "data: ")
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		var chunk openRouterChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			fullContent.WriteString(chunk.Choices[0].Delta.Content)

			if chunkHandler != nil {
				if err := chunkHandler(chunk.Choices[0].Delta.Content); err != nil {
					return response.Completion{}, 0, err
				}
			}
		}

		chunks++
		if chunk.Usage.TotalTokens != 0 {
			usage = response.Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}
	}

	return response.Completion{
		Content: fullContent.String(),
		Model:   model.ModelName,
		Usage:   usage,
	}, 0, nil
}

func (or OpenRouter) tryWithBackup(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	key := or.apiKeys[0]

	maxRetries := 5
	initialBackoff := 100 * time.Millisecond
	maxBackoff := 10 * time.Second

	var lastErr error
	for attempt := range maxRetries {
		requestLog.Events = append(requestLog.Events, response.Event{
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("attempting request with exponential backoff. attempt: %v", attempt),
		})

		select {
		case <-ctx.Done():
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("context cancelled: %v", ctx.Err()),
			})
			return response.Completion{}, ctx.Err()
		default:
			res, resCode, err := or.doRequest(ctx, req, client, chunkHandler, key)
			if err == nil {
				return res, nil
			}

			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("request failed: %v", err),
			})

			if !isRetryableError(resCode) {
				return response.Completion{}, err
			}

			lastErr = err

			backoff := min(initialBackoff*time.Duration(1<<attempt), maxBackoff)

			var randomBytes [8]byte
			var jitter time.Duration
			if _, err := rand.Read(randomBytes[:]); err != nil {
				jitter = backoff
			} else {
				randFloat := float64(binary.LittleEndian.Uint64(randomBytes[:])) / (1 << 64)
				jitter = time.Duration(float64(backoff) * (0.8 + 0.4*randFloat))
			}

			timer := time.NewTimer(jitter)
			select {
			case <-ctx.Done():
				timer.Stop()
				return response.Completion{}, ctx.Err()
			case <-timer.C:
				continue
			}
		}
	}

	return response.Completion{}, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (or OpenRouter) CompleteResponse(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	requestLog *response.Logging,
) (response.Completion, error) {
	reqLog := &response.Logging{}
	if requestLog == nil {
		req.Tags["request_type"] = "completion"
		reqLog = &response.Logging{
			Events: []response.Event{{
				Timestamp:   time.Now(),
				Description: "start of call to CompleteResponse",
			}},
			SystemMsg: req.SystemMessage,
			UserMsg:   req.UserMessage,
			Start:     time.Now(),
		}
	} else {
		reqLog = requestLog
	}

	for i, key := range or.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("attempting request with key_number: %v", i),
		})
		res, _, err := or.doRequest(ctx, req, client, nil, key)
		if err == nil {
			return res, nil
		}

		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("request failed: %v", err),
		})
	}

	return or.tryWithBackup(ctx, req, client, nil, reqLog)
}

func (or OpenRouter) StreamResponse(
	ctx context.Context,
	client http.Client,
	req request.Completion,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	reqLog := &response.Logging{}
	if requestLog == nil {
		req.Tags["request_type"] = "streaming"
		reqLog = &response.Logging{
			Events: []response.Event{{
				Timestamp:   time.Now(),
				Description: "start of call to StreamResponse",
			}},
			SystemMsg: req.SystemMessage,
			UserMsg:   req.UserMessage,
			Start:     time.Now(),
		}
	} else {
		reqLog = requestLog
	}

	for i, key := range or.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("attempting request with key_number: %v", i),
		})
		res, _, err := or.doRequest(ctx, req, client, chunkHandler, key)
		if err == nil {
			return res, nil
		}

		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("request failed: %v", err),
		})
	}

	return or.tryWithBackup(ctx, req, client, chunkHandler, reqLog)
}

var _ LLMProvider = new(OpenRouter)

func prepareOpenRouterRequest(
	req openRouterRequest,
	model models.OpenRouterModel,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openRouterRequest, error) {
	if len(model.PdfFile) > 0 && len(model.ImageFile) > 0 {
		return openRouterRequest{}, errors.New("only pdf file or image file, not both")
	}

	if len(model.ImageFile) > 0 {
		return prepareOpenRouterRequestWithImage(req, model.ImageFile, systemInst, userMsg, history)
	}

	if len(model.PdfFile) > 0 {
		return prepareOpenRouterRequestWithPdf(req, model.PdfFile, systemInst, userMsg, history)
	}

	return prepareOpenRouterBasicMessages(req, systemInst, userMsg, history)
}

func prepareOpenRouterRequestWithImage(
	req openRouterRequest,
	imageFiles []models.OpenRouterImagePayload,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openRouterRequest, error) {
	reqMsgWithImage := []requestMessageWithImage{}

	if systemInst != "" {
		reqMsgWithImage = append(reqMsgWithImage, requestMessageWithImage{
			Role: "system",
			Content: []any{
				fileInputMessage{Type: "text", Text: systemInst},
			},
		})
	}

	for _, his := range history {
		reqMsgWithImage = append(reqMsgWithImage, requestMessageWithImage{
			Role: his.Role,
			Content: []any{
				fileInputMessage{Type: "text", Text: his.Content},
			},
		})
	}

	lastIndex := 0
	if len(reqMsgWithImage) > 0 {
		lastIndex = len(reqMsgWithImage) - 1
	}

	if len(reqMsgWithImage) == 0 || reqMsgWithImage[lastIndex].Role != "user" {
		reqMsgWithImage = append(reqMsgWithImage, requestMessageWithImage{
			Role:    "user",
			Content: []any{},
		})
		lastIndex = len(reqMsgWithImage) - 1
	}

	for _, img := range imageFiles {
		detail := "auto"
		if img.Detail != "" {
			detail = img.Detail
		}

		ii := imageInput{
			Type: "image_url",
			ImageURL: imageURL{
				URL:    img.Url,
				Detail: detail,
			},
		}
		reqMsgWithImage[lastIndex].Content = append(reqMsgWithImage[lastIndex].Content, ii)
	}

	reqMsgWithImage[lastIndex].Content = append(reqMsgWithImage[lastIndex].Content,
		fileInputMessage{Type: "text", Text: userMsg})

	req.Messages = reqMsgWithImage
	return req, nil
}

func prepareOpenRouterRequestWithPdf(
	req openRouterRequest,
	pdfFiles map[string]string,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openRouterRequest, error) {
	reqMsgWithFile := []requestMessageWithFile{}

	if systemInst != "" {
		reqMsgWithFile = append(reqMsgWithFile, requestMessageWithFile{
			Role: "system",
			Content: []any{
				fileInputMessage{Type: "text", Text: systemInst},
			},
		})
	}

	for _, his := range history {
		reqMsgWithFile = append(reqMsgWithFile, requestMessageWithFile{
			Role: his.Role,
			Content: []any{
				fileInputMessage{Type: "text", Text: his.Content},
			},
		})
	}

	lastIndex := 0
	if len(reqMsgWithFile) > 0 {
		lastIndex = len(reqMsgWithFile) - 1
	}

	if len(reqMsgWithFile) == 0 || reqMsgWithFile[lastIndex].Role != "user" {
		reqMsgWithFile = append(reqMsgWithFile, requestMessageWithFile{
			Role:    "user",
			Content: []any{},
		})
		lastIndex = len(reqMsgWithFile) - 1
	}

	var filename string
	var fileData string
	for name, data := range pdfFiles {
		filename = name
		fileData = data
	}

	fi := fileInput{
		Type: "file",
		File: file{
			Filename: filename,
			FileData: fileData,
		},
	}
	reqMsgWithFile[lastIndex].Content = append(reqMsgWithFile[lastIndex].Content, fi)

	reqMsgWithFile[lastIndex].Content = append(reqMsgWithFile[lastIndex].Content,
		fileInputMessage{Type: "text", Text: userMsg})

	req.Messages = reqMsgWithFile
	return req, nil
}

func prepareOpenRouterBasicMessages(
	req openRouterRequest,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openRouterRequest, error) {
	hisLen := len(history)
	capacity := hisLen + 2
	if systemInst == "" {
		capacity = hisLen + 1
	}
	requestMessages := make([]requestMessage, 0, capacity)

	if systemInst != "" {
		requestMessages = append(requestMessages, requestMessage{
			Role:    "system",
			Content: systemInst,
		})
	}

	for i := range history {
		requestMessages = append(requestMessages, requestMessage{
			Role:    history[i].Role,
			Content: history[i].Content,
		})
	}

	requestMessages = append(requestMessages, requestMessage{
		Role:    "user",
		Content: userMsg,
	})

	req.Messages = requestMessages
	return req, nil
}
