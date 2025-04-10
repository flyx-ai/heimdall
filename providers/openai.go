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
	case models.GPT4OAlias:
		request, err := prepareGPT4ORequest(
			openaiRequest,
			req.Model,
			req.Messages,
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
			req.Messages,
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
			req.Messages,
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
			req.Messages,
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
		requestMessages := make([]requestMessage, len(req.Messages))
		for i, msg := range req.Messages {
			requestMessages[i] = requestMessage(requestMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

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
	reqLog := &response.Logging{}
	if requestLog == nil {
		var systemMsg string
		var userMsg string
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				systemMsg = msg.Content
			}
			if msg.Role == "user" {
				userMsg = msg.Content
			}
		}

		req.Tags["request_type"] = "streaming"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to StreamResponse",
				},
			},
			SystemMsg: systemMsg,
			UserMsg:   userMsg,
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
	reqLog := &response.Logging{}
	if requestLog == nil {
		var systemMsg string
		var userMsg string
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				systemMsg = msg.Content
			}
			if msg.Role == "user" {
				userMsg = msg.Content
			}
		}

		req.Tags["request_type"] = "streaming"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to StreamResponse",
				},
			},
			SystemMsg: systemMsg,
			UserMsg:   userMsg,
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
	messages []request.Message,
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
		reqMsgWithImage := []requestMessageWithImage{
			{
				Role:    "user",
				Content: []any{},
			},
		}

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
			reqMsgWithImage[0].Content = append(reqMsgWithImage[0].Content, ii)
		}

		for _, msg := range messages {
			if msg.Role == "user" {
				reqMsgWithImage[0].Content = append(
					reqMsgWithImage[0].Content,
					fileInputMessage{
						Type: "text",
						Text: msg.Content,
					},
				)
			}
		}

		request.Messages = reqMsgWithImage

		return request, nil
	}

	if len(gpt4O.PdfFile) == 1 {
		reqMsgWithFile := []requestMessageWithFile{
			{
				Role:    "user",
				Content: []any{},
			},
		}

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

		reqMsgWithFile[0].Content = append(reqMsgWithFile[0].Content, fi)
		for _, msg := range messages {
			if msg.Role == "user" {
				reqMsgWithFile[0].Content = append(
					reqMsgWithFile[0].Content,
					fileInputMessage{
						Type: "text",
						Text: msg.Content,
					},
				)
			}
		}

		request.Messages = reqMsgWithFile

		return request, nil
	}

	requestMessages := make([]requestMessage, len(messages))
	for i, msg := range messages {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareGPT4OMiniRequest(
	request openAIRequest,
	requestedModel models.Model,
	messages []request.Message,
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
		reqMsgWithImage := []requestMessageWithImage{
			{
				Role:    "user",
				Content: []any{},
			},
		}

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
			reqMsgWithImage[0].Content = append(reqMsgWithImage[0].Content, ii)
		}

		for _, msg := range messages {
			if msg.Role == "user" {
				reqMsgWithImage[0].Content = append(
					reqMsgWithImage[0].Content,
					fileInputMessage{
						Type: "text",
						Text: msg.Content,
					},
				)
			}
		}

		request.Messages = reqMsgWithImage

		return request, nil
	}

	if len(gpt4OMini.PdfFile) == 1 {
		reqMsgWithFile := []requestMessageWithFile{
			{
				Role:    "user",
				Content: []any{},
			},
		}

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

		reqMsgWithFile[0].Content = append(reqMsgWithFile[0].Content, fi)
		for _, msg := range messages {
			if msg.Role == "user" {
				reqMsgWithFile[0].Content = append(
					reqMsgWithFile[0].Content,
					fileInputMessage{
						Type: "text",
						Text: msg.Content,
					},
				)
			}
		}

		request.Messages = reqMsgWithFile

		return request, nil
	}

	requestMessages := make([]requestMessage, len(messages))
	for i, msg := range messages {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareO1Request(
	request openAIRequest,
	requestedModel models.Model,
	messages []request.Message,
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
		reqMsgWithImage := []requestMessageWithImage{
			{
				Role:    "user",
				Content: []any{},
			},
		}

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
			reqMsgWithImage[0].Content = append(reqMsgWithImage[0].Content, ii)
		}

		for _, msg := range messages {
			if msg.Role == "user" {
				reqMsgWithImage[0].Content = append(
					reqMsgWithImage[0].Content,
					fileInputMessage{
						Type: "text",
						Text: msg.Content,
					},
				)
			}
		}

		request.Messages = reqMsgWithImage

		return request, nil
	}

	if len(o1.PdfFile) == 1 {
		reqMsgWithFile := []requestMessageWithFile{
			{
				Role:    "user",
				Content: []any{},
			},
		}

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

		reqMsgWithFile[0].Content = append(reqMsgWithFile[0].Content, fi)
		for _, msg := range messages {
			if msg.Role == "user" {
				reqMsgWithFile[0].Content = append(
					reqMsgWithFile[0].Content,
					fileInputMessage{
						Type: "text",
						Text: msg.Content,
					},
				)
			}
		}

		request.Messages = reqMsgWithFile

		return request, nil
	}

	requestMessages := make([]requestMessage, len(messages))
	for i, msg := range messages {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	request.Messages = requestMessages

	return request, nil
}

func prepareO3MiniRequest(
	request openAIRequest,
	requestedModel models.Model,
	messages []request.Message,
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

	requestMessages := make([]requestMessage, len(messages))
	for i, msg := range messages {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	request.Messages = requestMessages

	return request, nil
}
