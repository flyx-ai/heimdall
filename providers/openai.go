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

// callImageGenerationAPI handles the request to the OpenAI image generation endpoint.
func (oa Openai) callImageGenerationAPI(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	key string,
) (response.Completion, int, error) {
	dalleModel, ok := req.Model.(*models.Dalle3)
	if !ok {
		return response.Completion{}, 0, errors.New("internal error: model is not Dalle3")
	}

	// Construct the image generation request payload
	imageReqPayload := map[string]any{
		"model":           dalleModel.GetName(),
		"prompt":          req.UserMessage,            // Use the main user message as the prompt
		"n":               1,                          // DALL-E 3 only supports n=1
		"size":            models.Dalle3Size1024x1024, // Default size
		"response_format": "url",                      // Heimdall handles URLs
	}

	// Apply optional parameters if provided
	if dalleModel.Size != "" {
		// TODO: Add validation for allowed sizes (1024x1024, 1792x1024, 1024x1792)
		imageReqPayload["size"] = dalleModel.Size
	}
	if dalleModel.Quality != "" {
		// TODO: Add validation for allowed quality (standard, hd)
		imageReqPayload["quality"] = dalleModel.Quality
	}
	if dalleModel.Style != "" {
		// TODO: Add validation for allowed styles (vivid, natural)
		imageReqPayload["style"] = dalleModel.Style
	}
	if dalleModel.User != "" {
		imageReqPayload["user"] = dalleModel.User
	}

	bodyBytes, err := json.Marshal(imageReqPayload)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("marshal image request: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/images/generations", openAIBaseURL),
		bytes.NewReader(bodyBytes))
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("create image request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

	// Execute the request
	resp, err := client.Do(httpReq)
	if err != nil {
		// Return 0 for status code on client-side errors
		return response.Completion{}, 0, fmt.Errorf("image request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for error details
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"received non-200 status code (%d) from image generation API: %s",
			resp.StatusCode, string(bodyBytes),
		)
	}

	// Parse the response
	// Expected response structure:
	// { "created": ..., "data": [ { "url": "...", "revised_prompt": "..." } ] }
	var imageResp struct {
		Created int64 `json:"created"`
		Data    []struct {
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
			// Base64JSON string `json:"b64_json"` // We are requesting URL format
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		// Return original status code on successful request but bad body
		return response.Completion{}, resp.StatusCode, fmt.Errorf("decode image response: %w", err)
	}

	// Format the response content (URLs separated by newline)
	var contentBuilder strings.Builder
	for i, imgData := range imageResp.Data {
		if i > 0 {
			contentBuilder.WriteString("\n")
		}
		contentBuilder.WriteString(imgData.URL)
		// Optionally include revised prompt if needed in the future
		// contentBuilder.WriteString("\nRevised Prompt: " + imgData.RevisedPrompt)
	}

	// TODO: Image generation doesn't provide token usage in the same way.
	// We might need to estimate or return zero usage. Heimdall expects Usage.
	usage := response.Usage{
		PromptTokens:     0, // Not directly available for images
		CompletionTokens: 0, // Not directly available for images
		TotalTokens:      0, // Not directly available for images
	}

	return response.Completion{
		Content: contentBuilder.String(),
		Model:   req.Model.GetName(),
		Usage:   usage,
		// FinishReason could be set to "stop" or similar if applicable
	}, resp.StatusCode, nil
}

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

	// --- START DALLE3 check ---
	// Check if the model is Dalle3 and handle non-streaming image generation.
	if _, ok := req.Model.(*models.Dalle3); ok {
		reqLog := requestLog // Use provided log or create if nil
		if reqLog == nil {
			req.Tags["request_type"] = "image_generation" // Mark as non-streaming
			reqLog = &response.Logging{
				Events:    []response.Event{{Timestamp: time.Now(), Description: "start of call to CompleteResponse (DALL-E 3)"}},
				SystemMsg: req.SystemMessage, // Might not be applicable
				UserMsg:   req.UserMessage,
				Start:     time.Now(),
			}
		} else {
			// Ensure Start time is set if using existing log
			if reqLog.Start.IsZero() {
				reqLog.Start = time.Now()
			}
			reqLog.Events = append(reqLog.Events, response.Event{Timestamp: time.Now(), Description: "Handling DALL-E 3 request in CompleteResponse"})
		}

		// Try keys sequentially. Add exponential backoff if needed.
		var lastErr error
		var lastStatusCode int
		for i, key := range oa.apiKeys {
			reqLog.Events = append(reqLog.Events, response.Event{
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("Attempting DALL-E 3 request with key_number: %d", i),
			})

			res, statusCode, err := oa.callImageGenerationAPI(ctx, req, client, key)

			lastStatusCode = statusCode // Store the status code from the attempt

			if err == nil {
				reqLog.Events = append(reqLog.Events, response.Event{
					Timestamp:   time.Now(),
					Description: fmt.Sprintf("DALL-E 3 request succeeded with key_number: %d, status: %d", i, statusCode),
				})
				// TODO: Add duration logging if reqLog is used
				// if reqLog != nil {
				// 	reqLog.Latency = time.Since(reqLog.Start)
				// 	// reqLog.FinalResponse = res // Might be too verbose to log full response
				// }
				return res, nil
			}

			lastErr = err // Store the error from the attempt
			reqLog.Events = append(reqLog.Events, response.Event{
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("DALL-E 3 request failed with key_number: %d, status: %d, err: %v", i, statusCode, err),
			})

			// TODO: Implement more sophisticated retry logic here if needed.
			// Should we retry on specific status codes (e.g., 429, 5xx)?
			// For now, just try the next key. Check if the error indicates a potentially retryable issue with the key itself.
			if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden || statusCode == http.StatusTooManyRequests {
				// Potentially key-related or rate limit, good reason to try next key.
				continue
			}
			// If it's another error (e.g., bad request 400, server error 5xx), maybe don't try other keys.
			// For simplicity now, we try all keys regardless of error type, but this could be refined.
		}

		// If all keys failed
		// Make sure lastErr is not nil before returning
		if lastErr == nil {
			lastErr = errors.New("image generation failed after trying all keys with unknown error")
		}
		return response.Completion{}, fmt.Errorf("image generation failed after trying all keys (last status %d): %w", lastStatusCode, lastErr)
	}
	// --- END DALLE3 check ---

	// Existing logic for non-streaming text completions...
	reqLog := &response.Logging{}
	if requestLog == nil {
		req.Tags["request_type"] = "completion" // Corrected tag to 'completion'

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

	// --- START DALLE3 check ---
	// If the model is Dalle3, delegate to CompleteResponse as image generation is not streaming.
	if _, ok := req.Model.(*models.Dalle3); ok {
		// Ensure logging context is passed correctly
		logCtx := requestLog
		if logCtx == nil {
			logCtx = &response.Logging{Start: time.Now()} // Minimal log if none provided
			logCtx.Events = append(logCtx.Events, response.Event{Timestamp: time.Now(), Description: "Initiating non-streaming call for DALL-E 3 from StreamResponse"})
		} else {
			// Ensure Start time is set if using existing log
			if logCtx.Start.IsZero() {
				logCtx.Start = time.Now()
			}
			// Add event indicating delegation
			logCtx.Events = append(logCtx.Events, response.Event{Timestamp: time.Now(), Description: "Delegating DALL-E 3 request from StreamResponse to CompleteResponse"})
		}
		// Pass the potentially initialized/updated logCtx
		return oa.CompleteResponse(ctx, req, client, logCtx)
	}
	// --- END DALLE3 check ---

	// Existing logic for streaming text completions...
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
