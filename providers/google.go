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

const googleBaseUrl = "https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s"

type Google struct {
	apiKeys []string
}

// NewGoogle register google as a provider on the router.
func NewGoogle(apiKeys []string) Google {
	return Google{
		apiKeys: apiKeys,
	}
}

type geminiRequest struct {
	SystemInstruction systemInstruction `json:"system_instruction,omitzero"`
	Contents          []content         `json:"contents"`
	Tools             models.GoogleTool `json:"tools"`
	Config            map[string]any    `json:"generationConfig"`
}

type content struct {
	Parts []part `json:"parts"`
}

type systemInstruction struct {
	Parts part `json:"parts"`
}

type fileData struct {
	MimeType string `json:"mime_type,omitzero"`
	FileURI  string `json:"file_uri,omitzero"`
}

type part struct {
	Text     string   `json:"text,omitzero"`
	FileData fileData `json:"file_data,omitzero"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata usageMetadata     `json:"usageMetadata"`
	ModelVersion  string            `json:"modelVersion"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiContent struct {
	Parts []geminiResponsePart `json:"parts"`
	Role  string               `json:"role"`
}

type geminiResponsePart struct {
	Text string `json:"text"`
}

type usageMetadata struct {
	PromptTokenCount        int             `json:"promptTokenCount"`
	CandidatesTokenCount    int             `json:"candidatesTokenCount"`
	TotalTokenCount         int             `json:"totalTokenCount"`
	PromptTokensDetails     []tokensDetails `json:"promptTokensDetails"`
	CandidatesTokensDetails []tokensDetails `json:"candidatesTokensDetails"`
}

type tokensDetails struct {
	Modality   string `json:"modality"`
	TokenCount int    `json:"tokenCount"`
}

func (g Google) CompleteResponse(
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

	for i, key := range g.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		res, _, err := g.doRequest(ctx, req, client, nil, key)
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

	return g.tryWithBackup(ctx, req, client, nil, reqLog)
}

// TODO figure out how to do tools with vertex sdk similar to the api
func (g Google) tryWithBackup(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	key := g.apiKeys[0]

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
			res, resCode, err := g.doRequest(
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

func (g Google) Name() string {
	return models.GoogleProvider
}

func (g Google) StreamResponse(
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

	for i, key := range g.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		res, _, err := g.doRequest(ctx, req, client, chunkHandler, key)
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

	return g.tryWithBackup(ctx, req, client, chunkHandler, reqLog)
}

func isRetryableError(resCode int) bool {
	return resCode > 400
}

func (g Google) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	model := req.Model
	geminiReq := geminiRequest{
		Contents: make([]content, 1),
	}

	var requestBody []byte

	switch model.GetName() {
	case models.Gemini15FlashModel:
		request, err := prepareGemini15FlashRequest(
			geminiReq,
			model,
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
	case models.Gemini15ProModel:
		request, err := prepareGemini15ProRequest(
			geminiReq,
			model,
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
	case models.Gemini20FlashModel:
		request, err := prepareGemini20FlashRequest(
			geminiReq,
			model,
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
	case models.Gemini20FlashLiteModel:
		request, err := prepareGemini20FlashLiteRequest(
			geminiReq,
			model,
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
	case models.Gemini25ProPreviewModel:
		request, err := prepareGemini25ProPreviewRequest(
			geminiReq,
			model,
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
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf(googleBaseUrl, req.Model.GetName(), key),
		bytes.NewReader(requestBody))
	if err != nil {
		return response.Completion{}, 0, err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Completion{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response.Completion{}, 0, errors.New(
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
			return response.Completion{}, 0, err
		}
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return response.Completion{}, 0, err
		}

		line = strings.TrimPrefix(line, "data: ")
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		var responseChunk geminiResponse
		if err := json.Unmarshal([]byte(line), &responseChunk); err != nil {
			return response.Completion{}, 0, err
		}

		if len(responseChunk.Candidates) > 0 {
			fullContent.WriteString(
				responseChunk.Candidates[0].Content.Parts[0].Text,
			)

			if chunkHandler != nil {
				if err := chunkHandler(responseChunk.Candidates[0].Content.Parts[0].Text); err != nil {
					return response.Completion{}, 0, err
				}
			}
		}

		chunks++

		if responseChunk.Candidates[0].FinishReason == "STOP" {
			usage = response.Usage{
				PromptTokens:     responseChunk.UsageMetadata.PromptTokenCount,
				CompletionTokens: responseChunk.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      responseChunk.UsageMetadata.TotalTokenCount,
			}
		}
	}

	return response.Completion{
		Content: fullContent.String(),
		Model:   req.Model.GetName(),
		Usage:   usage,
	}, 0, nil
}

var _ LLMProvider = new(Google)

func prepareGemini15FlashRequest(
	request geminiRequest,
	requestedModel models.Model,
	messages []request.Message,
) (geminiRequest, error) {
	// TODO: implement file, image etc on model
	_, ok := requestedModel.(models.Gemini15Flash)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini15Flash failed",
		)
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			request.SystemInstruction.Parts = part{
				Text: msg.Content,
			}
		}
		if msg.Role == "user" {
			request.Contents[0].Parts = append(
				request.Contents[0].Parts,
				part{
					Text: msg.Content,
				},
			)
		}
	}

	return request, nil
}

func prepareGemini15ProRequest(
	request geminiRequest,
	requestedModel models.Model,
	messages []request.Message,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini15Pro)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini15Pro failed",
		)
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			request.SystemInstruction.Parts = part{
				Text: msg.Content,
			}
		}
		if msg.Role == "user" {
			request.Contents[0].Parts = append(
				request.Contents[0].Parts,
				part{
					Text: msg.Content,
				},
			)
		}
	}

	if len(model.PdfFile) == 1 {
		var mimeType string
		var fileURI string

		for name, data := range model.PdfFile {
			mimeType = name
			fileURI = data
		}

		request.Contents[0].Parts = append(
			request.Contents[0].Parts,
			part{
				FileData: fileData{
					MimeType: mimeType,
					FileURI:  fileURI,
				},
			},
		)
	}

	if len(model.StructuredOutput) == 1 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	return request, nil
}

func prepareGemini20FlashRequest(
	request geminiRequest,
	requestedModel models.Model,
	messages []request.Message,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini20Flash)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini20Flash failed",
		)
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			request.SystemInstruction.Parts = part{
				Text: msg.Content,
			}
		}
		if msg.Role == "user" {
			request.Contents[0].Parts = append(
				request.Contents[0].Parts,
				part{
					Text: msg.Content,
				},
			)
		}
	}

	if len(model.PdfFile) == 1 {
		var mimeType string
		var fileURI string

		for name, data := range model.PdfFile {
			mimeType = name
			fileURI = data
		}

		request.Contents[0].Parts = append(
			request.Contents[0].Parts,
			part{
				FileData: fileData{
					MimeType: mimeType,
					FileURI:  fileURI,
				},
			},
		)
	}

	if len(model.StructuredOutput) == 1 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if len(model.Tools) > 1 {
		request.Tools = model.Tools
	}

	return request, nil
}

func prepareGemini20FlashLiteRequest(
	request geminiRequest,
	requestedModel models.Model,
	messages []request.Message,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini20FlashLite)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini20FlashLite failed",
		)
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			request.SystemInstruction.Parts = part{
				Text: msg.Content,
			}
		}
		if msg.Role == "user" {
			request.Contents[0].Parts = append(
				request.Contents[0].Parts,
				part{
					Text: msg.Content,
				},
			)
		}
	}

	if len(model.PdfFile) == 1 {
		var mimeType string
		var fileURI string

		for name, data := range model.PdfFile {
			mimeType = name
			fileURI = data
		}

		request.Contents[0].Parts = append(
			request.Contents[0].Parts,
			part{
				FileData: fileData{
					MimeType: mimeType,
					FileURI:  fileURI,
				},
			},
		)
	}

	if len(model.StructuredOutput) == 1 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if len(model.Tools) > 1 {
		request.Tools = model.Tools
	}

	return request, nil
}

func prepareGemini25ProPreviewRequest(
	request geminiRequest,
	requestedModel models.Model,
	messages []request.Message,
) (geminiRequest, error) {
	_, ok := requestedModel.(models.Gemini25ProPreview)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini25ProPreview  failed",
		)
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			request.SystemInstruction.Parts = part{
				Text: msg.Content,
			}
		}
		if msg.Role == "user" {
			request.Contents[0].Parts = append(
				request.Contents[0].Parts,
				part{
					Text: msg.Content,
				},
			)
		}
	}

	return request, nil
}
