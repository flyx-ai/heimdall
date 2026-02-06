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

var openAIBaseURL = "https://api.openai.com/v1"

// GetOpenAIBaseURL returns the current base URL for OpenAI API calls
func GetOpenAIBaseURL() string {
	return openAIBaseURL
}

// SetOpenAIBaseURL allows setting a custom base URL for OpenAI API calls (useful for testing)
func SetOpenAIBaseURL(url string) {
	openAIBaseURL = url
}

type requestMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type file struct {
	Filename string `json:"filename"`
	FileData string `json:"file_data"`
}

type fileInput struct {
	Type string `json:"type"`
	File file   `json:"file"`
}

type fileInputMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type requestMessageWithFile struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type imageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail"`
}
type imageInput struct {
	Type     string   `json:"type"`
	ImageURL imageURL `json:"image_url"`
}

type requestMessageWithImage struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type openAIChunk struct {
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

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAIRequest struct {
	Model          string         `json:"model"`
	Messages       any            `json:"messages"`
	Stream         bool           `json:"stream"`
	StreamOptions  streamOptions  `json:"stream_options"`
	Temperature    float32        `json:"temperature,omitempty"`
	TopP           float32        `json:"top_p,omitempty"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`
}

type Openai struct {
	apiKeys []string
}

func NewOpenAI(apiKeys []string) Openai {
	return Openai{
		apiKeys: apiKeys,
	}
}

func (oa Openai) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	model := req.Model.GetName()

	openaiRequest := openAIRequest{
		Model:         model,
		Stream:        true,
		StreamOptions: streamOptions{IncludeUsage: true},
		Temperature:   1.0,
	}

	request, err := prepareModelRequest(
		openaiRequest,
		req.Model,
		req.SystemMessage,
		req.UserMessage,
		req.History,
	)
	if err != nil {
		return response.Completion{}, 0, err
	}

	body, err := json.Marshal(request)
	if err != nil {
		return response.Completion{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/chat/completions", openAIBaseURL),
		bytes.NewReader(body))
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf(
			"create request: %w",
			err,
		)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Completion{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response.Completion{}, resp.StatusCode, errors.New(
			"received non-200 status code",
		)
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder
	var usage response.Usage
	var rawEvents []json.RawMessage
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
			return response.Completion{}, 0, fmt.Errorf(
				"read line: %w",
				err,
			)
		}

		line = strings.TrimPrefix(line, "data: ")
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		var chunk openAIChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return response.Completion{}, 0, fmt.Errorf(
				"unmarshal chunk: %w",
				err,
			)
		}

		rawEvents = append(rawEvents, json.RawMessage(line))

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

	rawResp, err := json.Marshal(rawEvents)
	if err != nil {
		rawResp = nil
	}

	return response.Completion{
		Content:     fullContent.String(),
		Model:       req.Model.GetName(),
		Usage:       usage,
		RawRequest:  body,
		RawResponse: rawResp,
	}, 0, nil
}

func (oa Openai) Name() string {
	return models.OpenaiProvider
}

// Check if an error is retryable based on the status code
// func isRetryableError(statusCode int) bool {
// 	switch statusCode {
// 	case http.StatusTooManyRequests, // Rate limit error
// 		http.StatusInternalServerError,
// 		http.StatusBadGateway,
// 		http.StatusServiceUnavailable,
// 		http.StatusGatewayTimeout:
// 		return true
// 	default:
// 		return false
// 	}
// }

// tryWithBackup implements LLMProvider.
func (oa Openai) tryWithBackup(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	key := oa.apiKeys[0]

	maxRetries := 5
	initialBackoff := 100 * time.Millisecond
	maxBackoff := 10 * time.Second

	var lastErr error
	for attempt := range maxRetries {
		requestLog.Events = append(requestLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with expoential backoff. attempt: %v",
				attempt,
			),
		})

		select {
		case <-ctx.Done():
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"context was called with error: %v",
					ctx.Err(),
				),
			})
			return response.Completion{}, ctx.Err()
		default:
			res, resCode, err := oa.doRequest(
				ctx,
				req,
				client,
				chunkHandler,
				key,
			)
			if err == nil {
				return res, nil
			}
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"request could not be completed, err: %v",
					err,
				),
			})

			if !isRetryableError(resCode) {
				requestLog.Events = append(requestLog.Events, response.Event{
					Timestamp: time.Now(),
					Description: fmt.Sprintf(
						"request was not retryable due to err: %v",
						err,
					),
				})
				return response.Completion{}, err
			}

			lastErr = err

			backoff := min(initialBackoff*time.Duration(
				1<<attempt,
			), maxBackoff)

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

	return response.Completion{}, fmt.Errorf(
		"max retries exceeded: %w",
		lastErr,
	)
}

func (oa Openai) CompleteResponse(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	requestLog *response.Logging,
) (response.Completion, error) {
	if _, ok := req.Model.(*models.GPTImage); ok {
		reqLog := requestLog
		if reqLog == nil {
			req.Tags["request_type"] = "image_generation"
			reqLog = &response.Logging{
				Events: []response.Event{
					{
						Timestamp:   time.Now(),
						Description: "start of call to CompleteResponse (DALL-E 3)",
					},
				},
				SystemMsg: req.SystemMessage, // Might not be applicable
				UserMsg:   req.UserMessage,
				Start:     time.Now(),
			}
		}
		if reqLog != nil {

			if reqLog.Start.IsZero() {
				reqLog.Start = time.Now()
			}
			reqLog.Events = append(
				reqLog.Events,
				response.Event{
					Timestamp:   time.Now(),
					Description: "Handling DALL-E 3 request in CompleteResponse",
				},
			)
		}

		var lastErr error
		var lastStatusCode int
		for i, key := range oa.apiKeys {
			reqLog.Events = append(reqLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"Attempting DALL-E 3 request with key_number: %d",
					i,
				),
			})

			// Retry logic for transient errors (5xx)
			maxRetries := 3
			for attempt := range maxRetries {
				res, statusCode, err := oa.callImageGenerationAPI(
					ctx,
					req,
					client,
					key,
				)

				lastStatusCode = statusCode

				if err == nil {
					reqLog.Events = append(reqLog.Events, response.Event{
						Timestamp: time.Now(),
						Description: fmt.Sprintf(
							"DALL-E 3 request succeeded with key_number: %d, status: %d",
							i,
							statusCode,
						),
					})

					return res, nil
				}

				lastErr = err

				// Retry on 5xx errors
				if statusCode >= 500 && statusCode < 600 && attempt < maxRetries-1 {
					reqLog.Events = append(reqLog.Events, response.Event{
						Timestamp: time.Now(),
						Description: fmt.Sprintf(
							"DALL-E 3 request got %d error, retrying (attempt %d/%d)",
							statusCode,
							attempt+1,
							maxRetries,
						),
					})
					backoff := time.Duration(1<<attempt) * time.Second
					select {
					case <-ctx.Done():
						return response.Completion{}, ctx.Err()
					case <-time.After(backoff):
						continue
					}
				}

				reqLog.Events = append(reqLog.Events, response.Event{
					Timestamp: time.Now(),
					Description: fmt.Sprintf(
						"DALL-E 3 request failed with key_number: %d, status: %d, err: %v",
						i,
						statusCode,
						err,
					),
				})
				break
			}

			if lastStatusCode == http.StatusUnauthorized ||
				lastStatusCode == http.StatusForbidden ||
				lastStatusCode == http.StatusTooManyRequests {
				continue
			}
		}

		if lastErr == nil {
			lastErr = errors.New(
				"image generation failed after trying all keys with unknown error",
			)
		}
		return response.Completion{}, fmt.Errorf(
			"image generation failed after trying all keys (last status %d): %w",
			lastStatusCode,
			lastErr,
		)
	}

	reqLog := &response.Logging{}
	if requestLog == nil {
		req.Tags["request_type"] = "completion"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to StreamResponse",
				},
			},
			SystemMsg: req.SystemMessage,
			UserMsg:   req.UserMessage,
			Start:     time.Now(),
		}
	}
	if requestLog != nil {
		reqLog = requestLog
	}

	for i, key := range oa.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		res, _, err := oa.doRequest(ctx, req, client, nil, key)
		if err == nil {
			return res, nil
		}

		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"request could not be completed, err: %v",
				err,
			),
		})
	}

	return oa.tryWithBackup(ctx, req, client, nil, reqLog)
}

func (oa Openai) StreamResponse(
	ctx context.Context,
	client http.Client,
	req request.Completion,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	if _, ok := req.Model.(*models.GPTImage); ok {
		logCtx := requestLog
		if logCtx == nil {
			logCtx = &response.Logging{
				Start: time.Now(),
			}
			logCtx.Events = append(
				logCtx.Events,
				response.Event{
					Timestamp:   time.Now(),
					Description: "Initiating non-streaming call for GPTImage from StreamResponse",
				},
			)
		}

		if logCtx != nil {
			if logCtx.Start.IsZero() {
				logCtx.Start = time.Now()
			}

			logCtx.Events = append(
				logCtx.Events,
				response.Event{
					Timestamp:   time.Now(),
					Description: "Delegating GPTImage request from StreamResponse to CompleteResponse",
				},
			)
		}

		return oa.CompleteResponse(ctx, req, client, logCtx)
	}

	reqLog := &response.Logging{}
	if requestLog == nil {
		req.Tags["request_type"] = "streaming"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to StreamResponse",
				},
			},
			SystemMsg: req.SystemMessage,
			UserMsg:   req.UserMessage,
			Start:     time.Now(),
		}
	}
	if requestLog != nil {
		reqLog = requestLog
	}

	for i, key := range oa.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		res, _, err := oa.doRequest(ctx, req, client, chunkHandler, key)
		if err == nil {
			return res, nil
		}

		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"request could not be completed, err: %v",
				err,
			),
		})
	}

	return oa.tryWithBackup(ctx, req, client, chunkHandler, reqLog)
}

func (oa Openai) callImageGenerationAPI(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	key string,
) (response.Completion, int, error) {
	gptImageModel, ok := req.Model.(*models.GPTImage)
	if !ok {
		return response.Completion{}, 0, errors.New(
			"internal error: model is not GPTImage",
		)
	}

	imageReqPayload := map[string]any{
		"model":  gptImageModel.GetName(),
		"prompt": req.UserMessage,
		"n":      1,
		"size":   models.GPTImageSize1024x1024,
	}

	if gptImageModel.Background != "" {
		imageReqPayload["background"] = gptImageModel.Background
	}
	if gptImageModel.Size != "" {
		imageReqPayload["size"] = gptImageModel.Size
	}
	if gptImageModel.Quality != "" {
		imageReqPayload["quality"] = gptImageModel.Quality
	}
	if gptImageModel.User != "" {
		imageReqPayload["user"] = gptImageModel.User
	}
	if gptImageModel.OutputFormat != "" {
		imageReqPayload["output_format"] = gptImageModel.OutputFormat
	}
	if gptImageModel.OutputFormat == "jpeg" ||
		gptImageModel.OutputFormat == "webp" &&
			gptImageModel.OutputCompression != "" {
		imageReqPayload["output_compression"] = gptImageModel.OutputCompression
	}
	if gptImageModel.Moderation != "" {
		imageReqPayload["moderation"] = gptImageModel.Moderation
	}

	bodyBytes, err := json.Marshal(imageReqPayload)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf(
			"marshal image request: %w",
			err,
		)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/images/generations", openAIBaseURL),
		bytes.NewReader(bodyBytes))
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf(
			"create image request: %w",
			err,
		)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf(
			"image request failed: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"received non-200 status code (%d) from image generation API: %s",
			resp.StatusCode, string(bodyBytes),
		)
	}

	var imageResp struct {
		Created int64 `json:"created"`
		Data    []struct {
			Base64JSON    string `json:"b64_json"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}

	var rawResponse bytes.Buffer
	teeBody := io.TeeReader(resp.Body, &rawResponse)
	if err := json.NewDecoder(teeBody).Decode(&imageResp); err != nil {
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"decode image response: %w",
			err,
		)
	}

	var contentBuilder strings.Builder
	for i, imgData := range imageResp.Data {
		if i > 0 {
			contentBuilder.WriteString("\n")
		}
		contentBuilder.WriteString(imgData.Base64JSON)
	}

	// TODO
	usage := response.Usage{
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
	}

	return response.Completion{
		Content:     contentBuilder.String(),
		Model:       req.Model.GetName(),
		Usage:       usage,
		RawRequest:  bodyBytes,
		RawResponse: rawResponse.Bytes(),
	}, resp.StatusCode, nil
}

var _ LLMProvider = new(Openai)

func prepareModelRequest(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	switch m := requestedModel.(type) {
	case models.GPT41:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT41Mini:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT41Nano:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT4O:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT4OMini:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT5:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT5Mini:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT5Nano:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT5Chat:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT51:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT51Chat:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT51Codex:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.GPT51CodexMini:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.O1:
		return prepareRequest(request, m.StructuredOutput, m.PdfFile, m.ImageFile, systemInst, userMsg, history)
	case models.O3Mini:
		if m.StructuredOutput != nil {
			request.ResponseFormat = map[string]any{
				"type":        "json_schema",
				"json_schema": m.StructuredOutput,
			}
		}
		return prepareBasicMessages(request, systemInst, userMsg, history)
	default:
		return prepareBasicMessages(request, systemInst, userMsg, history)
	}
}

func prepareRequest(
	request openAIRequest,
	structuredOutput map[string]any,
	pdfFile map[string]string,
	imageFile []models.OpenaiImagePayload,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	if structuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": structuredOutput,
		}
	}

	if len(pdfFile) > 0 && len(imageFile) > 0 {
		return openAIRequest{}, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(imageFile) > 0 {
		return prepareRequestWithImage(request, imageFile, userMsg, history)
	}

	if len(pdfFile) > 0 {
		return prepareRequestWithPdf(request, pdfFile, userMsg, history)
	}

	return prepareBasicMessages(request, systemInst, userMsg, history)
}

func prepareRequestWithImage(
	request openAIRequest,
	imageFiles []models.OpenaiImagePayload,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	reqMsgWithImage := []requestMessageWithImage{}

	for _, his := range history {
		reqMsgWithImage = append(reqMsgWithImage, requestMessageWithImage{
			Role: his.Role,
			Content: []any{
				fileInputMessage{
					Type: "text",
					Text: his.Content,
				},
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
		reqMsgWithImage[lastIndex].Content = append(
			reqMsgWithImage[lastIndex].Content,
			ii,
		)
	}

	reqMsgWithImage[lastIndex].Content = append(
		reqMsgWithImage[lastIndex].Content,
		fileInputMessage{
			Type: "text",
			Text: userMsg,
		},
	)

	request.Messages = reqMsgWithImage
	return request, nil
}

func prepareRequestWithPdf(
	request openAIRequest,
	pdfFiles map[string]string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	reqMsgWithFile := []requestMessageWithFile{}

	for _, his := range history {
		reqMsgWithFile = append(reqMsgWithFile, requestMessageWithFile{
			Role: his.Role,
			Content: []any{
				fileInputMessage{
					Type: "text",
					Text: his.Content,
				},
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
	reqMsgWithFile[lastIndex].Content = append(
		reqMsgWithFile[lastIndex].Content,
		fi,
	)

	reqMsgWithFile[lastIndex].Content = append(
		reqMsgWithFile[lastIndex].Content,
		fileInputMessage{
			Type: "text",
			Text: userMsg,
		},
	)

	request.Messages = reqMsgWithFile
	return request, nil
}

func prepareBasicMessages(
	request openAIRequest,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)

	for i := range history {
		requestMessages[i] = requestMessage{
			Role:    history[i].Role,
			Content: history[i].Content,
		}
	}

	requestMessages[hisLen] = requestMessage{
		Role:    "system",
		Content: systemInst,
	}

	requestMessages[hisLen+1] = requestMessage{
		Role:    "user",
		Content: userMsg,
	}

	request.Messages = requestMessages
	return request, nil
}
