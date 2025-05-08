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

const openAIBaseURL = "https://api.openai.com/v1"

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

type imageUrl struct {
	Url    string `json:"url"`
	Detail string `json:"detail"`
}
type imageInput struct {
	Type     string   `json:"type"`
	ImageUrl imageUrl `json:"image_url"`
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

	var requestBody []byte

	switch model {
	case models.GPT41MiniAlias:
		request, err := prepareGPT4MiniRequest(
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

		requestBody = body
	case models.GPT41NanoAlias:
		request, err := prepareGPT41NanoRequest(
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

		requestBody = body
	case models.GPT41Alias:
		request, err := prepareGPT41Request(
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

		requestBody = body
	case models.GPT4OAlias:
		request, err := prepareGPT4ORequest(
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

		requestBody = body
	case models.GPT4OMiniAlias:
		request, err := prepareGPT4OMiniRequest(
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

		requestBody = body
	case models.O1Alias:
		request, err := prepareO1Request(
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

		requestBody = body
	case models.O3MiniAlias:
		request, err := prepareO3MiniRequest(
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

		requestBody = body

	default:
		requestMessages := make([]requestMessage, 1)
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "user",
			Content: req.UserMessage,
		})

		openaiRequest.Messages = requestMessages
		body, err := json.Marshal(openaiRequest)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/chat/completions", openAIBaseURL),
		bytes.NewReader(requestBody))
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
		Model:   req.Model.GetName(),
		Usage:   usage,
	}, 0, nil
}

func (oa Openai) Name() string {
	return models.OpenaiProvider
}

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
			reqLog.Events = append(reqLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"DALL-E 3 request failed with key_number: %d, status: %d, err: %v",
					i,
					statusCode,
					err,
				),
			})

			if statusCode == http.StatusUnauthorized ||
				statusCode == http.StatusForbidden ||
				statusCode == http.StatusTooManyRequests {
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

	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
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
		Content: contentBuilder.String(),
		Model:   req.Model.GetName(),
		Usage:   usage,
	}, resp.StatusCode, nil
}

var _ LLMProvider = new(Openai)

func prepareGPT4ORequest(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	gpt4O, ok := requestedModel.(models.GPT4O)
	if !ok {
		return request, errors.New(
			"internal error; model was o3-mini but type assertion to models.O3Mini failed",
		)
	}

	if gpt4O.StructuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": gpt4O.StructuredOutput,
		}
	}

	if len(gpt4O.PdfFile) == 1 && len(gpt4O.ImageFile) == 1 {
		return openAIRequest{}, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(gpt4O.ImageFile) == 1 {
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

		lastIndex := len(reqMsgWithImage)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithImage[lastIndex].Role = "user"

		for _, img := range gpt4O.ImageFile {
			detail := "auto"
			if img.Detail != "" {
				detail = img.Detail
			}

			ii := imageInput{
				Type: "image_url",
				ImageUrl: imageUrl{
					Url:    img.Url,
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

	if len(gpt4O.PdfFile) == 1 {
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

		lastIndex := len(reqMsgWithFile)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithFile[lastIndex].Role = "user"
		var filename string
		var fileData string

		for name, data := range gpt4O.PdfFile {
			filename = name
			fileData = data
		}

		fi := fileInput{
			Type: "file",
			File: file{
				Filename: filename,
				FileData: string(fileData),
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

	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range history {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareGPT4OMiniRequest(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	gpt4OMini, ok := requestedModel.(models.GPT4OMini)
	if !ok {
		return request, errors.New(
			"internal error; model was o3-mini but type assertion to models.O3Mini failed",
		)
	}

	if gpt4OMini.StructuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": gpt4OMini.StructuredOutput,
		}
	}

	if len(gpt4OMini.PdfFile) == 1 && len(gpt4OMini.ImageFile) == 1 {
		return openAIRequest{}, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(gpt4OMini.ImageFile) == 1 {
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

		lastIndex := len(reqMsgWithImage)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithImage[lastIndex].Role = "user"

		for _, img := range gpt4OMini.ImageFile {
			detail := "auto"
			if img.Detail != "" {
				detail = img.Detail
			}

			ii := imageInput{
				Type: "image_url",
				ImageUrl: imageUrl{
					Url:    img.Url,
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

	if len(gpt4OMini.PdfFile) == 1 {
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

		lastIndex := len(reqMsgWithFile)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithFile[lastIndex].Role = "user"
		var filename string
		var fileData string

		for name, data := range gpt4OMini.PdfFile {
			filename = name
			fileData = data
		}

		fi := fileInput{
			Type: "file",
			File: file{
				Filename: filename,
				FileData: string(fileData),
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

	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range history {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareO1Request(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	o1, ok := requestedModel.(models.O1)
	if !ok {
		return request, errors.New(
			"internal error; model was o3-mini but type assertion to models.O3Mini failed",
		)
	}

	if o1.StructuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": o1.StructuredOutput,
		}
	}
	if len(o1.PdfFile) == 1 && len(o1.ImageFile) == 1 {
		return openAIRequest{}, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(o1.ImageFile) == 1 {
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

		lastIndex := len(reqMsgWithImage)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithImage[lastIndex].Role = "user"

		for _, img := range o1.ImageFile {
			detail := "auto"
			if img.Detail != "" {
				detail = img.Detail
			}

			ii := imageInput{
				Type: "image_url",
				ImageUrl: imageUrl{
					Url:    img.Url,
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

	if len(o1.PdfFile) == 1 {
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

		lastIndex := len(reqMsgWithFile)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithFile[lastIndex].Role = "user"
		var filename string
		var fileData string

		for name, data := range o1.PdfFile {
			filename = name
			fileData = data
		}

		fi := fileInput{
			Type: "file",
			File: file{
				Filename: filename,
				FileData: string(fileData),
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

	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range history {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareGPT4MiniRequest(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	gpt41Mini, ok := requestedModel.(models.GPT41Mini)
	if !ok {
		return request, errors.New(
			"internal error; model was gpt 4.1 mini but type assertion to models.GPT41Mini failed",
		)
	}

	if gpt41Mini.StructuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": gpt41Mini.StructuredOutput,
		}
	}
	if len(gpt41Mini.PdfFile) == 1 && len(gpt41Mini.ImageFile) == 1 {
		return openAIRequest{}, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(gpt41Mini.ImageFile) == 1 {
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

		lastIndex := len(reqMsgWithImage)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithImage[lastIndex].Role = "user"

		for _, img := range gpt41Mini.ImageFile {
			detail := "auto"
			if img.Detail != "" {
				detail = img.Detail
			}

			ii := imageInput{
				Type: "image_url",
				ImageUrl: imageUrl{
					Url:    img.Url,
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

	if len(gpt41Mini.PdfFile) == 1 {
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

		lastIndex := len(reqMsgWithFile)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithFile[lastIndex].Role = "user"
		var filename string
		var fileData string

		for name, data := range gpt41Mini.PdfFile {
			filename = name
			fileData = data
		}

		fi := fileInput{
			Type: "file",
			File: file{
				Filename: filename,
				FileData: string(fileData),
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

	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range history {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareGPT41NanoRequest(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	gpt41Nano, ok := requestedModel.(models.GPT41Nano)
	if !ok {
		return request, errors.New(
			"internal error; model was gpt 4.1 nano but type assertion to models.GPT41Nano failed",
		)
	}

	if gpt41Nano.StructuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": gpt41Nano.StructuredOutput,
		}
	}
	if len(gpt41Nano.PdfFile) == 1 && len(gpt41Nano.ImageFile) == 1 {
		return openAIRequest{}, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(gpt41Nano.ImageFile) == 1 {
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

		lastIndex := len(reqMsgWithImage)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithImage[lastIndex].Role = "user"

		for _, img := range gpt41Nano.ImageFile {
			detail := "auto"
			if img.Detail != "" {
				detail = img.Detail
			}

			ii := imageInput{
				Type: "image_url",
				ImageUrl: imageUrl{
					Url:    img.Url,
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

	if len(gpt41Nano.PdfFile) == 1 {
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

		lastIndex := len(reqMsgWithFile)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithFile[lastIndex].Role = "user"
		var filename string
		var fileData string

		for name, data := range gpt41Nano.PdfFile {
			filename = name
			fileData = data
		}

		fi := fileInput{
			Type: "file",
			File: file{
				Filename: filename,
				FileData: string(fileData),
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

	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range history {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareGPT41Request(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	gpt41, ok := requestedModel.(models.GPT41)
	if !ok {
		return request, errors.New(
			"internal error; model was gpt 4.1 but type assertion to models.GPT41 failed",
		)
	}

	if gpt41.StructuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": gpt41.StructuredOutput,
		}
	}
	if len(gpt41.PdfFile) == 1 && len(gpt41.ImageFile) == 1 {
		return openAIRequest{}, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(gpt41.ImageFile) == 1 {
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

		lastIndex := len(reqMsgWithImage)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithImage[lastIndex].Role = "user"

		for _, img := range gpt41.ImageFile {
			detail := "auto"
			if img.Detail != "" {
				detail = img.Detail
			}

			ii := imageInput{
				Type: "image_url",
				ImageUrl: imageUrl{
					Url:    img.Url,
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

	if len(gpt41.PdfFile) == 1 {
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

		lastIndex := len(reqMsgWithFile)
		if lastIndex == 1 {
			lastIndex = 0
		}

		reqMsgWithFile[lastIndex].Role = "user"
		var filename string
		var fileData string

		for name, data := range gpt41.PdfFile {
			filename = name
			fileData = data
		}

		fi := fileInput{
			Type: "file",
			File: file{
				Filename: filename,
				FileData: string(fileData),
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

	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range history {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareO3MiniRequest(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	o3Mini, ok := requestedModel.(models.O3Mini)
	if !ok {
		return request, errors.New(
			"internal error; model was o3-mini but type assertion to models.O3Mini failed",
		)
	}

	if o3Mini.StructuredOutput != nil {
		request.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": o3Mini.StructuredOutput,
		}
	}

	hisLen := len(history)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range history {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: systemInst,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: userMsg,
		})
	}

	request.Messages = requestMessages

	return request, nil
}
