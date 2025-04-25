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
	return resCode > 400
}

func (g Google) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	if req.SystemMessage == "" || req.UserMessage == "" {
		return response.Completion{}, 0, errors.New(
			"gemini models require both system message and user message",
		)
	}

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
	case models.Gemini15FlashModel:
		preparedReq, err := prepareGemini15FlashRequest(
			geminiReq,
			model,
			req.SystemMessage,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini15ProModel:
		preparedReq, err := prepareGemini15ProRequest(
			geminiReq,
			model,
			req.SystemMessage,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini20FlashModel:
		preparedReq, err := prepareGemini20FlashRequest(
			geminiReq,
			model,
			req.SystemMessage,
			req.UserMessage,
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
			req.SystemMessage,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini25ProPreviewModel:
		preparedReq, err := prepareGemini25ProPreviewRequest(
			geminiReq,
			model,
			req.SystemMessage,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
		if err != nil {
			return response.Completion{}, 0, err
		}

		requestBody = body
	case models.Gemini25FlashPreviewModel:
		preparedReq, err := prepareGemini25FlashPreviewRequest(
			geminiReq,
			model,
			req.SystemMessage,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		body, err := json.Marshal(preparedReq)
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
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	// TODO: implement file, image etc on model
	model, ok := requestedModel.(models.Gemini15Flash)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini15Flash failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	request.Contents[lastIndex].Parts = append(
		request.Contents[lastIndex].Parts,
		part{
			Text: userMsg,
		},
	)

	if model.Thinking != "" {
		request = handleThinkingBudget(request, model.Thinking)
	}

	return request, nil
}

func prepareGemini15ProRequest(
	request geminiRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
) (geminiRequest, error) {
	model, ok := requestedModel.(models.Gemini15Pro)
	if !ok {
		return request, errors.New(
			"internal error; model type assertion to models.Gemini15Pro failed",
		)
	}

	request.SystemInstruction.Parts = part{
		Text: systemInst,
	}

	lastIndex := 0
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	request.Contents[lastIndex].Parts = append(
		request.Contents[lastIndex].Parts,
		part{
			Text: userMsg,
		},
	)

	// Only one of image or PDF inputs may be provided
	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return geminiRequest{}, errors.New("only pdf file or image file can be provided, not both")
	}
	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}
	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
	}

	if len(model.StructuredOutput) == 1 {
		request.Config = map[string]any{
			"response_mime_type": "application/json",
			"response_schema":    model.StructuredOutput,
		}
	}

	if model.Thinking != "" {
		request = handleThinkingBudget(request, model.Thinking)
	}

	return request, nil
}

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
	if len(request.Contents) > 1 {
		lastIndex = len(request.Contents) - 1
	}

	request.Contents[lastIndex].Role = "user"
	request.Contents[lastIndex].Parts = append(
		request.Contents[lastIndex].Parts,
		part{Text: userMsg},
	)
	// Only one of image or PDF inputs may be provided
	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New("only pdf file or image file can be provided, not both")
	}
	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}
	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
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

	request.Contents[lastIndex].Parts = append(
		request.Contents[lastIndex].Parts,
		part{Text: userMsg},
	)
	// Only one of image or PDF inputs may be provided
	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New("only pdf file or image file can be provided, not both")
	}
	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}
	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
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

	request.Contents[lastIndex].Parts = append(
		request.Contents[lastIndex].Parts,
		part{Text: userMsg},
	)
	// Only one of image or PDF inputs may be provided
	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New("only pdf file or image file can be provided, not both")
	}
	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}
	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
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

	request.Contents[lastIndex].Parts = append(
		request.Contents[lastIndex].Parts,
		part{Text: userMsg},
	)
	// Only one of image or PDF inputs may be provided
	if len(model.PdfFiles) > 0 && len(model.ImageFile) > 0 {
		return request, errors.New("only pdf file or image file can be provided, not both")
	}
	if len(model.ImageFile) > 0 {
		request = handleVisionData(request, model.ImageFile)
	}
	if len(model.PdfFiles) > 0 {
		request = handlePdfData(request, model.PdfFiles, lastIndex)
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
				if len(base64Part) > 0 {
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

// handlePdfData appends PDF inputs (URIs or base64) to request contents at given index
func handlePdfData(request geminiRequest, pdfs []models.GooglePdf, contentIdx int) geminiRequest {
	const pdfMimeType = "application/pdf"
	
	for _, pdf := range pdfs {
		pdfStr := string(pdf)
		
		if strings.HasPrefix(pdfStr, "https://") {
			// external URI
			request.Contents[contentIdx].Parts = append(
				request.Contents[contentIdx].Parts,
				filePart{InlineData: fileData{MimeType: pdfMimeType, FileURI: pdfStr}},
			)
		} else {
			// inline base64
			data := pdfStr
			prefix := fmt.Sprintf("data:%s;base64,", pdfMimeType)
			if parts := strings.SplitN(pdfStr, prefix, 2); len(parts) == 2 {
				data = parts[1]
			}
			request.Contents[contentIdx].Parts = append(
				request.Contents[contentIdx].Parts,
				filePart{InlineData: imageData{MimeType: pdfMimeType, Data: data}},
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
			"thinkingConfig": map[string]int64{
				"thinkingBudget": 24576,
			},
		}
	case models.MediumThinkBudget:
		request.Config = map[string]any{
			"thinkingConfig": map[string]int64{
				"thinkingBudget": 12288,
			},
		}
	case models.LowThinkBudget:
		request.Config = map[string]any{
			"thinkingConfig": map[string]int64{
				"thinkingBudget": 0,
			},
		}
	}

	return request
}
