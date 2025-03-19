package heimdall

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

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
)

const googleBaseUrl = "https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s"

type GoogleOptions struct {
	vertexAIClient *genai.Client
}

type GoogleOption func(*GoogleOptions)

type Google struct {
	apiKeys        []string
	vertexAIClient *genai.Client
}

func WithVertexAI(
	ctx context.Context,
	projectID,
	location,
	credentialsJSON string,
) GoogleOption {
	return func(opts *GoogleOptions) {
		client, err := genai.NewClient(
			ctx,
			projectID,
			location,
			option.WithCredentialsJSON([]byte(credentialsJSON)),
		)
		if err != nil {
			return
		}
		opts.vertexAIClient = client
	}
}

// NewGoogle register google as a provider on the router. If you want to use vertex ai you have to add it using the GoogleOptions functions.
func NewGoogle(apiKeys []string, opts ...GoogleOption) Google {
	options := &GoogleOptions{}

	for _, opt := range opts {
		opt(options)
	}

	return Google{
		apiKeys:        apiKeys,
		vertexAIClient: options.vertexAIClient,
	}
}

type geminiRequest struct {
	SystemInstruction systemInstruction `json:"system_instruction,omitzero"`
	Contents          []content         `json:"contents"`
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

func (g Google) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	requestLog *Logging,
) (CompletionResponse, error) {
	switch req.Model {
	case ModelVertexGemini20FlashLite,
		ModelVertexGemini20Flash,
		ModelVertexGemini10Pro,
		ModelVertexGemini10ProVision,
		ModelVertexGemini15Pro,
		ModelVertexGemini15FlashThinking:
		if g.vertexAIClient == nil {
			return CompletionResponse{}, errors.New(
				"vertex ai model requested without having configured the client",
			)
		}
		return g.completeResponseVertex(ctx, req, requestLog)
	default:
		for i, key := range g.apiKeys {
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"attempting to complete request with key_number: %v",
					i,
				),
			})
			response, _, err := g.doRequest(ctx, req, client, nil, key)
			if err == nil {
				return response, nil
			}
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"request could not be completed, err: %v",
					err,
				),
			})
		}
	}

	return g.tryWithBackup(ctx, req, client, nil, requestLog)
}

func (g Google) completeResponseVertex(
	ctx context.Context,
	req CompletionRequest,
	requestLog *Logging,
) (CompletionResponse, error) {
	var parts []genai.Part
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "user" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "file" {
			parts = append(parts, genai.FileData{
				MIMEType: string(msg.FileType),
				FileURI:  msg.Content,
			})
		}
	}

	maxRetries := 5
	initialBackoff := 100 * time.Millisecond
	maxBackoff := 10 * time.Second

	var lastErr error
	for attempt := range maxRetries {
		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with expoential backoff. attempt: %v",
				attempt,
			),
		})

		select {
		case <-ctx.Done():
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"context was called with error: %v",
					ctx.Err(),
				),
			})
			return CompletionResponse{}, ctx.Err()
		default:

			model := g.vertexAIClient.GenerativeModel(req.Model.Name)

			res, err := model.GenerateContent(ctx, parts...)
			if err == nil {
				rb, err := json.MarshalIndent(res, "", "  ")
				if err != nil {
					return CompletionResponse{}, err
				}

				return CompletionResponse{
					Content: string(rb),
					Model:   req.Model,
					Usage: Usage{
						PromptTokens: int(
							res.UsageMetadata.PromptTokenCount,
						),
						CompletionTokens: int(
							res.UsageMetadata.CandidatesTokenCount,
						),
						TotalTokens: int(
							res.UsageMetadata.TotalTokenCount,
						),
					},
				}, nil
			}

			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"request could not be completed, err: %v",
					err,
				),
			})

			lastErr = err

			backoff := min(initialBackoff*time.Duration(
				1<<attempt,
			), maxBackoff)

			// Generate cryptographically secure random number
			var randomBytes [8]byte
			var jitter time.Duration
			if _, err := rand.Read(randomBytes[:]); err != nil {
				// Fallback to 1.0 multiplier if random generation fails
				jitter = backoff
			} else {
				// Convert to float64 between 0 and 1
				randFloat := float64(binary.LittleEndian.Uint64(randomBytes[:])) / (1 << 64)
				jitter = time.Duration(float64(backoff) * (0.8 + 0.4*randFloat))
			}

			timer := time.NewTimer(jitter)
			select {
			case <-ctx.Done():
				timer.Stop()
				return CompletionResponse{}, ctx.Err()
			case <-timer.C:
				continue
			}
		}
	}

	return CompletionResponse{}, fmt.Errorf(
		"max retries exceeded: %w",
		lastErr,
	)
}

func (g Google) tryWithBackup(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *Logging,
) (CompletionResponse, error) {
	key := g.apiKeys[0]

	maxRetries := 5
	initialBackoff := 100 * time.Millisecond
	maxBackoff := 10 * time.Second

	var lastErr error
	for attempt := range maxRetries {
		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with expoential backoff. attempt: %v",
				attempt,
			),
		})

		select {
		case <-ctx.Done():
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"context was called with error: %v",
					ctx.Err(),
				),
			})
			return CompletionResponse{}, ctx.Err()
		default:
			response, resCode, err := g.doRequest(
				ctx,
				req,
				client,
				chunkHandler,
				key,
			)
			if err == nil {
				return response, nil
			}
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"request could not be completed, err: %v",
					err,
				),
			})

			if !isRetryableError(resCode) {
				requestLog.Events = append(requestLog.Events, Event{
					Timestamp: time.Now(),
					Description: fmt.Sprintf(
						"request was not retryable due to err: %v",
						err,
					),
				})
				return CompletionResponse{}, err
			}

			lastErr = err

			backoff := min(initialBackoff*time.Duration(
				1<<attempt,
			), maxBackoff)

			// Generate cryptographically secure random number
			var randomBytes [8]byte
			var jitter time.Duration
			if _, err := rand.Read(randomBytes[:]); err != nil {
				// Fallback to 1.0 multiplier if random generation fails
				jitter = backoff
			} else {
				// Convert to float64 between 0 and 1
				randFloat := float64(binary.LittleEndian.Uint64(randomBytes[:])) / (1 << 64)
				jitter = time.Duration(float64(backoff) * (0.8 + 0.4*randFloat))
			}

			timer := time.NewTimer(jitter)
			select {
			case <-ctx.Done():
				timer.Stop()
				return CompletionResponse{}, ctx.Err()
			case <-timer.C:
				continue
			}
		}
	}

	return CompletionResponse{}, fmt.Errorf(
		"max retries exceeded: %w",
		lastErr,
	)
}

func (g Google) getApiKeys() []string {
	return g.apiKeys
}

func (g Google) name() string {
	return "google"
}

func (g Google) streamResponse(
	ctx context.Context,
	client http.Client,
	req CompletionRequest,
	chunkHandler func(chunk string) error,
	requestLog *Logging,
) (CompletionResponse, error) {
	switch req.Model {
	case ModelVertexGemini20FlashLite,
		ModelVertexGemini20Flash,
		ModelVertexGemini10Pro,
		ModelVertexGemini10ProVision,
		ModelVertexGemini15Pro,
		ModelVertexGemini15FlashThinking:
		if g.vertexAIClient == nil {
			return CompletionResponse{}, errors.New(
				"vertex ai model requested without having configured the client",
			)
		}

		// streaming response with the google sdk seems to not be working so we just do a regular completion request for now
		return g.completeResponseVertex(ctx, req, requestLog)
	default:
		for i, key := range g.apiKeys {
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"attempting to complete request with key_number: %v",
					i,
				),
			})
			response, _, err := g.doRequest(ctx, req, client, chunkHandler, key)
			if err == nil {
				return response, nil
			}
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"request could not be completed, err: %v",
					err,
				),
			})
		}
	}

	return g.tryWithBackup(ctx, req, client, chunkHandler, requestLog)
}

func isRetryableError(resCode int) bool {
	return resCode > 400
}

func (g Google) doRequest(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (CompletionResponse, int, error) {
	geminiReq := geminiRequest{
		Contents: make([]content, 1),
	}
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			geminiReq.SystemInstruction.Parts = part{
				Text: msg.Content,
			}
		}
		if msg.Role == "user" {
			geminiReq.Contents[0].Parts = append(
				geminiReq.Contents[0].Parts,
				part{
					Text: msg.Content,
				},
			)
		}
		if msg.Role == "file" {
			geminiReq.Contents[0].Parts = append(
				geminiReq.Contents[0].Parts,
				part{
					FileData: fileData{
						MimeType: string(msg.FileType),
						FileURI:  msg.Content,
					},
				},
			)
		}
	}

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return CompletionResponse{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf(googleBaseUrl, req.Model.Name, key),
		bytes.NewReader(body))
	if err != nil {
		return CompletionResponse{}, 0, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CompletionResponse{}, resp.StatusCode, err
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder
	var usage Usage
	chunks := 0
	now := time.Now()

	for {
		if chunks == 0 && time.Since(now).Seconds() > 3.0 {
			return CompletionResponse{}, 0, context.Canceled
		}
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return CompletionResponse{}, 0, err
		}

		line = strings.TrimPrefix(line, "data: ")
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		var responseChunk geminiResponse
		if err := json.Unmarshal([]byte(line), &responseChunk); err != nil {
			return CompletionResponse{}, 0, err
		}

		if len(responseChunk.Candidates) > 0 {
			fullContent.WriteString(
				responseChunk.Candidates[0].Content.Parts[0].Text,
			)

			if chunkHandler != nil {
				if err := chunkHandler(responseChunk.Candidates[0].Content.Parts[0].Text); err != nil {
					return CompletionResponse{}, 0, err
				}
			}
		}

		chunks++

		if responseChunk.Candidates[0].FinishReason == "STOP" {
			usage = Usage{
				PromptTokens:     responseChunk.UsageMetadata.PromptTokenCount,
				CompletionTokens: responseChunk.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      responseChunk.UsageMetadata.TotalTokenCount,
			}
		}
	}

	return CompletionResponse{
		Content: fullContent.String(),
		Model:   req.Model,
		Usage:   usage,
	}, 0, nil
}

var _ LLMProvider = new(Google)
