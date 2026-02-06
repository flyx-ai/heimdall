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
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

const googleBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s"

type Google struct {
	apiKeys []string
}

type cacheContentRequest struct {
	Model             string        `json:"model"`
	Contents          []content     `json:"contents"`
	SystemInstruction systemContent `json:"system_instruction"`
	TTL               string        `json:"ttl"`
}

type systemContent struct {
	Parts []part `json:"parts"`
	Role  string `json:"role"`
}

type cacheContentResponse struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
	ExpireTime time.Time `json:"expireTime"`
}

// NewGoogle register google as a provider on the router.
func NewGoogle(apiKeys []string) Google {
	return Google{
		apiKeys: apiKeys,
	}
}

type geminiRequest struct {
	SystemInstruction systemInstruction `json:"system_instruction"`
	Contents          []content         `json:"contents"`
	Tools             models.GoogleTool `json:"tools"`
	Config            map[string]any    `json:"generationConfig"`
}

type content struct {
	Role  string `json:"role"`
	Parts []any  `json:"parts"`
}

type systemInstruction struct {
	Parts any `json:"parts"`
}

type fileData struct {
	MimeType string `json:"mime_type,omitzero"`
	FileURI  string `json:"file_uri,omitzero"`
}
type imageData struct {
	MimeType string `json:"mime_type,omitzero"`
	Data     string `json:"data,omitzero"`
}

type fileURI struct {
	FileData fileData `json:"file_data"`
}

type filePart struct {
	InlineData any `json:"inline_data,omitzero"`
}

type part struct {
	Text     string `json:"text,omitzero"`
	FileData any    `json:"file_data,omitzero"`
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
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

type usageMetadata struct {
	PromptTokenCount        int             `json:"promptTokenCount"`
	CandidatesTokenCount    int             `json:"candidatesTokenCount"`
	TotalTokenCount         int             `json:"totalTokenCount"`
	ThoughtsTokenCount      int             `json:"thoughtsTokenCount,omitempty"`
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
	if len(g.apiKeys) == 0 {
		return response.Completion{}, errors.New("no API keys available")
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
	if len(g.apiKeys) == 0 {
		return response.Completion{}, errors.New("no API keys available")
	}
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

// CacheContentPayload represents the data to be cached. Must be either text or fileData but not both.
type CacheContentPayload struct {
	Text     string
	FileData map[string]string
}

// CacheContent caches the provided content with the specified TTL and returns a content ID
// that can be used to reference this content in subsequent requests.
func (g Google) CacheContent(
	ctx context.Context,
	model string,
	payload CacheContentPayload,
	systemInstruction string,
	ttl time.Duration,
) (string, error) {
	if len(g.apiKeys) == 0 {
		return "", errors.New("no API keys available")
	}

	key := g.apiKeys[0]
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/cachedContents?key=%s",
		key,
	)

	if payload.Text != "" && payload.FileData != nil {
		return "", errors.New("only one of text or fileData can be provided")
	}

	if len(payload.FileData) > 1 {
		return "", errors.New("you can only provide one file")
	}

	reqBody := cacheContentRequest{
		Model: "models/" + model,
		Contents: []content{{
			Role: "user",
		}},
		SystemInstruction: systemContent{
			Role: "system",
			Parts: []part{{
				Text: systemInstruction,
			}},
		},
		TTL: fmt.Sprintf("%ds", int(ttl.Seconds())),
	}

	if payload.Text != "" {
		reqBody.Contents[0].Parts = append(
			reqBody.Contents[0].Parts,
			part{
				Text: payload.Text,
			},
		)
	}
	if payload.FileData != nil {
		var mimeType string
		var fileURI string

		for k, v := range payload.FileData {
			mimeType = k
			fileURI = v
		}

		reqBody.Contents[0].Parts = append(
			reqBody.Contents[0].Parts,
			part{
				FileData: fileData{
					MimeType: mimeType,
					FileURI:  fileURI,
				},
			},
		)
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"unexpected status code %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var cacheResp cacheContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&cacheResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return cacheResp.Name, nil
}

func (g Google) UpdateCachedContentTTL(
	ctx context.Context,
	cacheName string,
	ttl time.Duration,
) error {
	if len(g.apiKeys) == 0 {
		return errors.New("no API keys available")
	}

	key := g.apiKeys[0]
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/%s?key=%s",
		cacheName,
		key,
	)

	reqBody := struct {
		TTL string `json:"ttl"`
	}{
		TTL: fmt.Sprintf("%ds", int(ttl.Seconds())),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		url,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"unexpected status code %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var cacheResp cacheContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&cacheResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// CachedContentsList represents the response from listing cached contents
type CachedContentsList struct {
	CachedContents []CachedContent `json:"cachedContents"`
	NextPageToken  string          `json:"nextPageToken,omitempty"`
}

// CachedContent represents a single cached content item
type CachedContent struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
	ExpireTime time.Time `json:"expireTime"`
}

// ListCachedContents retrieves a list of all cached contents
func (g Google) ListCachedContents(
	ctx context.Context,
) (*CachedContentsList, error) {
	if len(g.apiKeys) == 0 {
		return nil, errors.New("no API keys available")
	}

	key := g.apiKeys[0]
	baseURL := "https://generativelanguage.googleapis.com/v1beta/cachedContents?key=" + key

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"unexpected status code %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var result CachedContentsList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteCachedContent retrieves a list of all cached contents
func (g Google) DeleteCachedContent(
	ctx context.Context,
	cacheName string,
) error {
	if len(g.apiKeys) == 0 {
		return errors.New("no API keys available")
	}

	key := g.apiKeys[0]
	baseURL := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/%s?key=%s",
		cacheName,
		key,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		baseURL,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"unexpected status code %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}

func (g Google) StreamResponse(
	ctx context.Context,
	client http.Client,
	req request.Completion,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	if len(g.apiKeys) == 0 {
		return response.Completion{}, errors.New("no API keys available")
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
	return resCode == 429 || resCode >= 500
}

func (g Google) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	// Handle image generation models separately
	if _, ok := req.Model.(*models.Gemini25FlashImage); ok {
		return g.doGemini25FlashImageRequest(ctx, req, client, key)
	}
	if _, ok := req.Model.(*models.Gemini3ProImagePreview); ok {
		return g.doGemini3ProImageRequest(ctx, req, client, key)
	}

	if req.SystemMessage == "" || req.UserMessage == "" {
		return response.Completion{}, 400, errors.New(
			"gemini models require both system message and user message",
		)
	}

	systemMessage := req.SystemMessage
	userMessage := req.UserMessage

	model := req.Model
	geminiReq := geminiRequest{
		Contents: make([]content, len(req.History)+1),
	}

	for i, his := range req.History {
		role := his.Role
		if role == "assistant" {
			role = "model"
		}
		geminiReq.Contents[i] = content{
			Role: role,
			Parts: []any{
				part{Text: his.Content},
			},
		}
	}

	var requestBody []byte

	switch model.GetName() {
	case models.Gemini20FlashModel:
		preparedReq, err := prepareGemini20FlashRequest(
			geminiReq,
			model,
			systemMessage,
			userMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini20FlashLiteModel:
		preparedReq, err := prepareGemini20FlashLiteRequest(
			geminiReq,
			model,
			systemMessage,
			userMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini25ProModel:
		preparedReq, err := prepareGemini25ProPreviewRequest(
			geminiReq,
			model,
			systemMessage,
			userMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini25FlashModel:
		preparedReq, err := prepareGemini25FlashPreviewRequest(
			geminiReq,
			model,
			systemMessage,
			userMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini3ProModel:
		preparedReq, err := prepareGemini3ProPreviewRequest(
			geminiReq,
			model,
			systemMessage,
			userMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini3FlashModel:
		preparedReq, err := prepareGemini3FlashPreviewRequest(
			geminiReq,
			model,
			systemMessage,
			userMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini25FlashLiteModel:
		preparedReq, err := prepareGemini25FlashLiteRequest(
			geminiReq,
			model,
			systemMessage,
			userMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	default:
		return response.Completion{}, 0, fmt.Errorf(
			"unsupported Gemini model: %s",
			model.GetName(),
		)
	}

	if requestBody == nil {
		return response.Completion{}, 0, errors.New(
			"request body is nil - model may not be supported",
		)
	}

	apiURL := fmt.Sprintf(googleBaseURL, req.Model.GetName(), key)
	log.Printf("[Heimdall] Making request to Google API: model=%s url=%s", req.Model.GetName(), strings.Split(apiURL, "?")[0])
	log.Printf("[Heimdall] Request body size: %d bytes", len(requestBody))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL,
		bytes.NewReader(requestBody))
	if err != nil {
		return response.Completion{}, 0, err
	}

	log.Printf("[Heimdall] Sending HTTP request...")
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[Heimdall] HTTP request failed: %v", err)
		return response.Completion{}, 0, err
	}
	defer resp.Body.Close()
	log.Printf("[Heimdall] Got response: status=%d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Printf("[Heimdall] Error response (status %d), failed to read body: %v", resp.StatusCode, readErr)
			return response.Completion{}, resp.StatusCode, fmt.Errorf(
				"received non-200 status code (%d), failed to read error body: %w",
				resp.StatusCode, readErr,
			)
		}
		log.Printf("[Heimdall] Error response (status %d): %s", resp.StatusCode, string(bodyBytes))
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"received non-200 status code (%d): %s",
			resp.StatusCode, string(bodyBytes),
		)
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder
	var thoughts strings.Builder
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

		rawEvents = append(rawEvents, json.RawMessage(line))

		if len(responseChunk.Candidates) > 0 {
			if len(responseChunk.Candidates[0].Content.Parts) > 0 {
				part := responseChunk.Candidates[0].Content.Parts[0]
				if part.Thought {
					thoughts.WriteString(part.Text)
				} else {
					fullContent.WriteString(part.Text)
				}

				if chunkHandler != nil {
					if !part.Thought {
						if err := chunkHandler(part.Text); err != nil {
							return response.Completion{}, 0, err
						}
					}
				}
			}
		}

		chunks++

		if len(responseChunk.Candidates) > 0 &&
			responseChunk.Candidates[0].FinishReason == "STOP" {
			usage = response.Usage{
				PromptTokens:     responseChunk.UsageMetadata.PromptTokenCount,
				CompletionTokens: responseChunk.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      responseChunk.UsageMetadata.TotalTokenCount,
			}
		}
	}

	rawResp, err := json.Marshal(rawEvents)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("marshal raw response events: %w", err)
	}

	return response.Completion{
		Content:     fullContent.String(),
		Thoughts:    thoughts.String(),
		Model:       req.Model.GetName(),
		Usage:       usage,
		RawRequest:  requestBody,
		RawResponse: rawResp,
	}, 0, nil
}

var _ LLMProvider = new(Google)

func prepareGemini20FlashRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini20Flash)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini20Flash failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) >= 1 {
		lastIndex = len(request.Contents) - 1
	}

	if len(request.Contents) > 0 {
		request.Contents[lastIndex].Parts = append(
			request.Contents[lastIndex].Parts,
			part{Text: userMsg},
		)
		request.Contents[lastIndex].Role = "user"
	}

	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}

	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.Files) > 0 {
		request = handleGenericFiles(request, model.Files, lastIndex)
	}

	if len(model.StructuredOutput) > 0 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if len(model.Tools) > 1 {
		request.Tools = model.Tools
	}

	if model.Thinking != "" {
		request = handleThinkingBudget(request, model.Thinking)
	}

	return request, nil
}

func prepareGemini20FlashLiteRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini20FlashLite)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini20FlashLite failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	if len(request.Contents) > 0 {
		request.Contents[lastIndex].Parts = append(
			request.Contents[lastIndex].Parts,
			part{Text: userMsg},
		)
		request.Contents[lastIndex].Role = "user"
	}

	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}

	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.Files) > 0 {
		request = handleGenericFiles(request, model.Files, lastIndex)
	}

	if len(model.StructuredOutput) > 0 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if len(model.Tools) > 1 {
		request.Tools = model.Tools
	}

	if model.Thinking != "" {
		request = handleThinkingBudget(request, model.Thinking)
	}

	return request, nil
}

func prepareGemini25FlashLiteRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini25FlashLite)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini25FlashLite failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	if len(request.Contents) > 0 {
		request.Contents[lastIndex].Parts = append(
			request.Contents[lastIndex].Parts,
			part{Text: userMsg},
		)
		request.Contents[lastIndex].Role = "user"
	}

	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}

	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.Files) > 0 {
		request = handleGenericFiles(request, model.Files, lastIndex)
	}

	if len(model.StructuredOutput) > 0 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if len(model.Tools) > 1 {
		request.Tools = model.Tools
	}

	if model.Thinking != "" {
		request = handleThinkingBudget(request, model.Thinking)
	}

	return request, nil
}

func prepareGemini25FlashPreviewRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini25FlashPreview)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini25FlashPreview failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	if len(request.Contents) > 0 {
		request.Contents[lastIndex].Parts = append(
			request.Contents[lastIndex].Parts,
			part{Text: userMsg},
		)
		request.Contents[lastIndex].Role = "user"
	}

	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}

	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.Files) > 0 {
		request = handleGenericFiles(request, model.Files, lastIndex)
	}

	if len(model.StructuredOutput) > 0 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if len(model.Tools) > 1 {
		request.Tools = model.Tools
	}

	if model.Thinking != "" {
		request = handleThinkingBudget(request, model.Thinking)
	}

	return request, nil
}

func prepareGemini25ProPreviewRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini25ProPreview)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini25ProPreview failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	if len(request.Contents) > 0 {
		request.Contents[lastIndex].Parts = append(
			request.Contents[lastIndex].Parts,
			part{Text: userMsg},
		)
		request.Contents[lastIndex].Role = "user"
	}

	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}

	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.Files) > 0 {
		request = handleGenericFiles(request, model.Files, lastIndex)
	}

	if len(model.StructuredOutput) > 0 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if len(model.Tools) > 1 {
		request.Tools = model.Tools
	}

	if model.Thinking != "" {
		request = handleThinkingBudget(request, model.Thinking)
	}

	return request, nil
}

func handleVisionData(
	request geminiRequest,
	imageFiles []models.GoogleImagePayload,
) geminiRequest {
	for _, imgFile := range imageFiles {
		if strings.HasPrefix(imgFile.Data, "https://") {
			request.Contents[0].Parts = append(
				request.Contents[0].Parts,
				filePart{
					InlineData: fileData{
						MimeType: imgFile.MimeType,
						FileURI:  imgFile.Data,
					},
				},
			)
		}
		if !strings.HasPrefix(imgFile.Data, "https://") {
			base64 := imgFile.Data

			fullBase64 := fmt.Sprintf("data:%s;base64,", imgFile.MimeType)
			if strings.Contains(imgFile.Data, fullBase64) {
				base64Part := strings.Split(
					imgFile.Data,
					fullBase64,
				)
				if len(base64Part) > 1 {
					base64 = base64Part[1]
				}
			}

			request.Contents[0].Parts = append(
				request.Contents[0].Parts,
				filePart{
					InlineData: imageData{
						MimeType: imgFile.MimeType,
						Data:     base64,
					},
				},
			)
		}
	}

	return request
}

func handlePdfData(
	request geminiRequest,
	pdfs []models.GooglePdf,
	contentIdx int,
) geminiRequest {
	const pdfMimeType = "application/pdf"

	for _, pdf := range pdfs {
		pdfStr := string(pdf)

		if strings.HasPrefix(pdfStr, "https://") {
			request.Contents[contentIdx].Parts = append(
				request.Contents[contentIdx].Parts,
				fileURI{
					FileData: fileData{
						MimeType: pdfMimeType,
						FileURI:  pdfStr,
					},
				},
			)
		}
		if !strings.HasPrefix(pdfStr, "https://") {
			data := pdfStr
			prefix := fmt.Sprintf("data:%s;base64,", pdfMimeType)
			if parts := strings.SplitN(pdfStr, prefix, 2); len(parts) == 2 {
				data = parts[1]
			}
			request.Contents[contentIdx].Parts = append(
				request.Contents[contentIdx].Parts,
				filePart{
					InlineData: imageData{MimeType: pdfMimeType, Data: data},
				},
			)
		}
	}
	return request
}

func handleGenericFiles(
	request geminiRequest,
	files []models.GoogleFilePayload,
	contentIdx int,
) geminiRequest {
	for _, file := range files {
		if strings.HasPrefix(file.Data, "https://") {
			request.Contents[contentIdx].Parts = append(
				request.Contents[contentIdx].Parts,
				fileURI{
					FileData: fileData{
						MimeType: file.MimeType,
						FileURI:  file.Data,
					},
				},
			)
		}

		if !strings.HasPrefix(file.Data, "https://") {
			data := file.Data
			prefix := fmt.Sprintf("data:%s;base64,", file.MimeType)
			if parts := strings.SplitN(file.Data, prefix, 2); len(parts) == 2 {
				data = parts[1]
			}
			request.Contents[contentIdx].Parts = append(
				request.Contents[contentIdx].Parts,
				filePart{
					InlineData: imageData{
						MimeType: file.MimeType,
						Data:     data,
					},
				},
			)
		}
	}
	return request
}

func handleThinkingBudget(
	request geminiRequest,
	budget models.ThinkBudget,
) geminiRequest {
	switch budget {
	case models.HighThinkBudget:
		request.Config = map[string]any{
			"thinkingConfig": map[string]any{
				"thinkingBudget":  int64(24576),
				"includeThoughts": true,
			},
		}
	case models.MediumThinkBudget:
		request.Config = map[string]any{
			"thinkingConfig": map[string]any{
				"thinkingBudget":  int64(12288),
				"includeThoughts": true,
			},
		}
	case models.LowThinkBudget:
		request.Config = map[string]any{
			"thinkingConfig": map[string]any{
				"thinkingBudget":  int64(0),
				"includeThoughts": false,
			},
		}
	}

	return request
}

func handleThinkingLevel(
	request geminiRequest,
	level models.ThinkingLevel,
) geminiRequest {
	if level == "" {
		return request
	}

	if request.Config == nil {
		request.Config = map[string]any{}
	}

	thinkingConfig := map[string]any{
		"thinkingLevel": string(level),
	}

	if existingConfig, ok := request.Config["thinkingConfig"].(map[string]any); ok {
		for k, v := range thinkingConfig {
			existingConfig[k] = v
		}
	} else {
		request.Config["thinkingConfig"] = thinkingConfig
	}

	return request
}

func handleMediaResolution(
	request geminiRequest,
	resolution models.MediaResolution,
) geminiRequest {
	if resolution == "" {
		return request
	}

	if request.Config == nil {
		request.Config = map[string]any{}
	}

	request.Config["mediaResolution"] = string(resolution)

	return request
}

func prepareGemini3ProPreviewRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini3ProPreview)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini3ProPreview failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	if len(request.Contents) > 0 {
		request.Contents[lastIndex].Parts = append(
			request.Contents[lastIndex].Parts,
			part{Text: userMsg},
		)
		request.Contents[lastIndex].Role = "user"
	}

	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}

	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.Files) > 0 {
		request = handleGenericFiles(request, model.Files, lastIndex)
	}

	if len(model.StructuredOutput) > 0 {
		if request.Config == nil {
			request.Config = map[string]any{}
		}
		request.Config["response_mime_type"] = "application/json"
		request.Config["response_schema"] = model.StructuredOutput
	}

	if len(model.Tools) > 0 {
		request.Tools = model.Tools
	}

	if model.ThinkingLevel != "" {
		request = handleThinkingLevel(request, model.ThinkingLevel)
	}

	if model.MediaResolution != "" {
		request = handleMediaResolution(request, model.MediaResolution)
	}

	return request, nil
}

func prepareGemini3FlashPreviewRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini3FlashPreview)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini3FlashPreview failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	if len(request.Contents) > 0 {
		request.Contents[lastIndex].Parts = append(
			request.Contents[lastIndex].Parts,
			part{Text: userMsg},
		)
		request.Contents[lastIndex].Role = "user"
	}

	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New(
			"only pdf file or image file can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}

	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.Files) > 0 {
		request = handleGenericFiles(request, model.Files, lastIndex)
	}

	if len(model.StructuredOutput) > 0 {
		if request.Config == nil {
			request.Config = map[string]any{}
		}
		request.Config["response_mime_type"] = "application/json"
		request.Config["response_schema"] = model.StructuredOutput
	}

	if len(model.Tools) > 0 {
		request.Tools = model.Tools
	}

	if model.ThinkingLevel != "" {
		request = handleThinkingLevel(request, model.ThinkingLevel)
	}

	if model.MediaResolution != "" {
		request = handleMediaResolution(request, model.MediaResolution)
	}

	return request, nil
}

func (g Google) doGemini3ProImageRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	key string,
) (response.Completion, int, error) {
	imageModel, ok := req.Model.(*models.Gemini3ProImagePreview)
	if !ok {
		return response.Completion{}, 0, errors.New(
			"internal error: model is not Gemini3ProImagePreview",
		)
	}

	parts := []any{}

	if len(imageModel.ImageFile) > 0 {
		for _, img := range imageModel.ImageFile {
			base64 := img.Data
			fullBase64 := fmt.Sprintf("data:%s;base64,", img.MimeType)
			if strings.Contains(img.Data, fullBase64) {
				base64Part := strings.Split(img.Data, fullBase64)
				if len(base64Part) > 1 {
					base64 = base64Part[1]
				}
			}
			parts = append(parts, filePart{
				InlineData: imageData{
					MimeType: img.MimeType,
					Data:     base64,
				},
			})
		}
	}

	if len(imageModel.PdfFiles) > 0 {
		for _, pdf := range imageModel.PdfFiles {
			pdfStr := string(pdf)
			if strings.HasPrefix(pdfStr, "https://") {
				parts = append(parts, fileURI{
					FileData: fileData{
						MimeType: "application/pdf",
						FileURI:  pdfStr,
					},
				})
			} else {
				data := pdfStr
				prefix := "data:application/pdf;base64,"
				if pdfParts := strings.SplitN(pdfStr, prefix, 2); len(pdfParts) == 2 {
					data = pdfParts[1]
				}
				parts = append(parts, filePart{
					InlineData: imageData{
						MimeType: "application/pdf",
						Data:     data,
					},
				})
			}
		}
	}

	parts = append(parts, part{Text: req.UserMessage})

	requestPayload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": parts,
				"role":  "user",
			},
		},
	}

	if req.SystemMessage != "" {
		requestPayload["systemInstruction"] = map[string]any{
			"parts": []any{
				part{Text: req.SystemMessage},
			},
		}
	}

	generationConfig := map[string]any{
		"responseModalities": []string{"Text", "Image"},
	}

	if imageModel.AspectRatio != "" {
		generationConfig["imageConfig"] = map[string]any{
			"aspectRatio": string(imageModel.AspectRatio),
		}
	}

	if imageModel.ThinkingLevel != "" {
		generationConfig["thinkingConfig"] = map[string]any{
			"thinkingLevel": string(imageModel.ThinkingLevel),
		}
	}

	if imageModel.MediaResolution != "" {
		generationConfig["mediaResolution"] = string(imageModel.MediaResolution)
	}

	requestPayload["generationConfig"] = generationConfig

	bodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		models.Gemini3ProImageModel,
		key,
	)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"received non-200 status code (%d): %s",
			resp.StatusCode, string(bodyBytes),
		)
	}

	var rawResponse bytes.Buffer
	teeBody := io.TeeReader(resp.Body, &rawResponse)
	var imageResp gemini25FlashImageResponse
	if err := json.NewDecoder(teeBody).Decode(&imageResp); err != nil {
		return response.Completion{}, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(imageResp.Candidates) == 0 || len(imageResp.Candidates[0].Content.Parts) == 0 {
		return response.Completion{}, 0, errors.New("no image data in response")
	}

	var imgData string
	for _, p := range imageResp.Candidates[0].Content.Parts {
		if p.InlineData != nil && p.InlineData.Data != "" {
			imgData = p.InlineData.Data
			break
		}
	}

	if imgData == "" {
		return response.Completion{}, 0, errors.New("no image data found in response parts")
	}

	return response.Completion{
		Content: imgData,
		Model:   models.Gemini3ProImageModel,
		Usage: response.Usage{
			PromptTokens:     imageResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: imageResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      imageResp.UsageMetadata.TotalTokenCount,
		},
		RawRequest:  bodyBytes,
		RawResponse: rawResponse.Bytes(),
	}, http.StatusOK, nil
}

// Image generation response structures
type gemini25FlashImageResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text,omitempty"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// doGemini25FlashImageRequest handles image generation via Gemini 2.5 Flash Image model
func (g Google) doGemini25FlashImageRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	key string,
) (response.Completion, int, error) {
	imageModel, ok := req.Model.(*models.Gemini25FlashImage)
	if !ok {
		return response.Completion{}, 0, errors.New(
			"internal error: model is not Gemini25FlashImage",
		)
	}

	// Build the request payload
	parts := []any{}

	// Add image attachments if present
	if len(imageModel.ImageFile) > 0 {
		for _, img := range imageModel.ImageFile {
			base64 := img.Data

			fullBase64 := fmt.Sprintf("data:%s;base64,", img.MimeType)
			if strings.Contains(img.Data, fullBase64) {
				base64Part := strings.Split(
					img.Data,
					fullBase64,
				)
				if len(base64Part) > 1 {
					base64 = base64Part[1]
				}
			}

			parts = append(parts, filePart{
				InlineData: imageData{
					MimeType: img.MimeType,
					Data:     base64,
				},
			})
		}
	}

	// Add PDF attachments if present
	if len(imageModel.PdfFiles) > 0 {
		for _, pdf := range imageModel.PdfFiles {
			pdfStr := string(pdf)
			if strings.HasPrefix(pdfStr, "https://") {
				parts = append(parts, fileURI{
					FileData: fileData{
						MimeType: "application/pdf",
						FileURI:  pdfStr,
					},
				})
			} else {
				data := pdfStr
				prefix := fmt.Sprintf("data:%s;base64,", "application/pdf")
				if pdfParts := strings.SplitN(pdfStr, prefix, 2); len(pdfParts) == 2 {
					data = pdfParts[1]
				}
				parts = append(parts, filePart{
					InlineData: imageData{
						MimeType: "application/pdf",
						Data:     data,
					},
				})
			}
		}
	}

	// Add user message
	parts = append(parts, part{Text: req.UserMessage})

	requestPayload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": parts,
				"role":  "user",
			},
		},
	}

	if req.SystemMessage != "" {
		requestPayload["systemInstruction"] = map[string]any{
			"parts": []any{
				part{Text: req.SystemMessage},
			},
		}
	}

	// Add generation config
	// https://ai.google.dev/gemini-api/docs/image-generation
	generationConfig := map[string]any{
		"responseModalities": []string{"Text", "Image"},
	}

	if imageModel.AspectRatio != "" {
		imageConfig := map[string]any{
			"aspectRatio": string(imageModel.AspectRatio),
		}
		generationConfig["imageConfig"] = imageConfig
	}

	requestPayload["generationConfig"] = generationConfig

	bodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use generateContent endpoint for Gemini 2.5 Flash Image
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		models.Gemini25FlashImageModel,
		key,
	)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Completion{}, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"received non-200 status code (%d) from Gemini image generation API: %s",
			resp.StatusCode, string(bodyBytes),
		)
	}

	var rawResponse bytes.Buffer
	teeBody := io.TeeReader(resp.Body, &rawResponse)
	var imageResp gemini25FlashImageResponse
	if err := json.NewDecoder(teeBody).Decode(&imageResp); err != nil {
		return response.Completion{}, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(imageResp.Candidates) == 0 || len(imageResp.Candidates[0].Content.Parts) == 0 {
		return response.Completion{}, 0, errors.New("no image data in response")
	}

	// Extract base64 image data from parts
	var imageData string
	for _, part := range imageResp.Candidates[0].Content.Parts {
		if part.InlineData != nil && part.InlineData.Data != "" {
			imageData = part.InlineData.Data
			break
		}
	}

	if imageData == "" {
		return response.Completion{}, 0, errors.New("no image data found in response parts")
	}

	return response.Completion{
		Content: imageData,
		Model:   models.Gemini25FlashImageModel,
		Usage: response.Usage{
			PromptTokens:     imageResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: imageResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      imageResp.UsageMetadata.TotalTokenCount,
		},
		RawRequest:  bodyBytes,
		RawResponse: rawResponse.Bytes(),
	}, http.StatusOK, nil
}
