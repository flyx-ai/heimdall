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
	"net/http"
	"strings"
	"time"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

const anthropicBaseUrl = "https://api.anthropic.com/v1"

type Anthropic struct {
	apiKeys []string
}

// NewAnthropic creates a new Anthropic LLM provider with the given API keys.
func NewAnthropic(apiKeys []string) Anthropic {
	return Anthropic{
		apiKeys: apiKeys,
	}
}

type (
	mediaSource struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	}
	anthropicMediaPayload struct {
		Type   string      `json:"type"`
		Source mediaSource `json:"source"`
	}
	anthropicTextPayload struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
)

type anthropicMsg struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicRequest struct {
	System      string         `json:"system"`
	Model       string         `json:"model"`
	Messages    []anthropicMsg `json:"messages"`
	Stream      bool           `json:"stream"`
	MaxTokens   int            `json:"max_tokens"`
	Temperature float32        `json:"temperature,omitempty"`
	TopP        float32        `json:"top_p,omitempty"`
}

// CompleteResponse implements LLMProvider.
func (a Anthropic) CompleteResponse(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	requestLog *response.Logging,
) (response.Completion, error) {
	reqLog := &response.Logging{}
	if requestLog == nil {
		req.Tags["request_type"] = "completion"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to CompleteResponse",
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

	for i, key := range a.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})

		res, _, err := a.doRequest(ctx, req, client, nil, key)
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

	return a.tryWithBackup(ctx, req, client, nil, reqLog)
}

// doRequest implements LLMProvider.
func (a Anthropic) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	modelName := req.Model.GetName()

	var messages []anthropicMsg

	if len(req.History) > 0 {
		for _, his := range req.History {
			messages = append(messages, anthropicMsg{
				Role:    his.Role,
				Content: his.Content,
			})
		}
	}

	switch modelName {
	case models.AnthropicClaude3OpusAlias:
		msgs, err := prepareClaude3Opus(
			req.Model,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}
		messages = append(messages, msgs...)
	case models.AnthropicClaude35HaikuAlias:
		msgs, err := prepareClaude35Haiku(
			req.Model,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		messages = append(messages, msgs...)
	case models.AnthropicClaude35SonnetAlias:
		msgs, err := prepareClaude35Sonnet(
			req.Model,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}

		messages = append(messages, msgs...)
	case models.AnthropicClaude37SonnetAlias:
		msgs, err := prepareClaude37Sonnet(
			req.Model,
			req.UserMessage,
		)
		if err != nil {
			return response.Completion{}, 0, err
		}
		messages = append(messages, msgs...)
	}

	apiReq := anthropicRequest{
		System:      req.SystemMessage,
		Model:       modelName,
		Messages:    messages,
		Stream:      true,
		MaxTokens:   4096,
		Temperature: 1.0,
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return response.Completion{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/messages", anthropicBaseUrl),
		bytes.NewReader(body))
	if err != nil {
		return response.Completion{}, 0, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", key)
	httpReq.Header.Set("Anthropic-Version", "2023-06-01")

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Completion{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response.Completion{}, resp.StatusCode, err
	}

	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder

	chunks := 0
	now := time.Now()
	isRunning := true

	type DeltaEvent struct {
		Type  string `json:"type"`
		Index int    `json:"index"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	}

	for isRunning {
		if chunks == 0 && time.Since(now).Seconds() > 3.0 {
			return response.Completion{}, 0, context.Canceled
		}

		var completeText strings.Builder

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				var event DeltaEvent
				err := json.Unmarshal([]byte(dataStr), &event)
				if err != nil {
					return response.Completion{}, 0, err
				}

				if event.Type == "content_block_delta" &&
					event.Delta.Type == "text_delta" {
					completeText.WriteString(event.Delta.Text)

					if chunkHandler != nil {
						if err := chunkHandler(event.Delta.Text); err != nil {
							return response.Completion{}, 0, err
						}
					}
				}

				chunks++
			}
		}

		err := scanner.Err()
		switch err {
		case nil:
			fullContent = completeText
			isRunning = false
		default:
			fmt.Println("Error reading input:", err)
			return response.Completion{}, 0, context.Canceled
		}
	}

	return response.Completion{
		Content: fullContent.String(),
		Model:   req.Model.GetName(),
		// TODO: try to standardize this across providers
		Usage: response.Usage{
			// CompletionTokens: lastResponse.Usage.OutputTokens,
			// PromptTokens:     lastResponse.Usage.InputTokens,
		},
	}, 0, nil
}

func (a Anthropic) Name() string {
	return models.AnthropicProvider
}

// StreamResponse implements LLMProvider.
func (a Anthropic) StreamResponse(
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

	for i, key := range a.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		res, _, err := a.doRequest(ctx, req, client, chunkHandler, key)
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

	return a.tryWithBackup(ctx, req, client, chunkHandler, requestLog)
}

// tryWithBackup implements LLMProvider.
func (a Anthropic) tryWithBackup(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	key := a.apiKeys[0]

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
			res, resCode, err := a.doRequest(
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

var _ LLMProvider = new(Anthropic)

func prepareClaude3Opus(
	requestedModel models.Model,
	userMsg string,
) ([]anthropicMsg, error) {
	model, ok := requestedModel.(models.Claude3Opus)
	if !ok {
		return nil, errors.New(
			"internal error; model type assertion to models.Claude3Opus failed",
		)
	}

	if len(model.ImageFile) > 0 && len(model.PdfFiles) > 0 {
		return nil, errors.New(
			"only image file or pdf files can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		return handleMedia(userMsg, model.ImageFile, nil), nil
	}

	if len(model.PdfFiles) > 0 {
		return handleMedia(userMsg, nil, model.PdfFiles), nil
	}

	return []anthropicMsg{
		{
			Role:    "user",
			Content: userMsg,
		},
	}, nil
}

func prepareClaude35Sonnet(
	requestedModel models.Model,
	userMsg string,
) ([]anthropicMsg, error) {
	model, ok := requestedModel.(models.Claude35Sonnet)
	if !ok {
		return nil, errors.New(
			"internal error; model type assertion to models.Claude35Sonnet failed",
		)
	}

	if len(model.ImageFile) > 0 && len(model.PdfFiles) > 0 {
		return nil, errors.New(
			"only image file or pdf files can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		return handleMedia(userMsg, model.ImageFile, nil), nil
	}

	if len(model.PdfFiles) > 0 {
		return handleMedia(userMsg, nil, model.PdfFiles), nil
	}

	return []anthropicMsg{
		{
			Role:    "user",
			Content: userMsg,
		},
	}, nil
}

func prepareClaude35Haiku(
	requestedModel models.Model,
	userMsg string,
) ([]anthropicMsg, error) {
	model, ok := requestedModel.(models.Claude35Haiku)
	if !ok {
		return nil, errors.New(
			"internal error; model type assertion to models.Claude35Haiku failed",
		)
	}

	if len(model.ImageFile) > 0 && len(model.PdfFiles) > 0 {
		return nil, errors.New(
			"only image file or pdf files can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		return handleMedia(userMsg, model.ImageFile, nil), nil
	}

	if len(model.PdfFiles) > 0 {
		return handleMedia(userMsg, nil, model.PdfFiles), nil
	}

	return []anthropicMsg{
		{
			Role:    "user",
			Content: userMsg,
		},
	}, nil
}

func prepareClaude37Sonnet(
	requestedModel models.Model,
	userMsg string,
) ([]anthropicMsg, error) {
	model, ok := requestedModel.(models.Claude37Sonnet)
	if !ok {
		return nil, errors.New(
			"internal error; model type assertion to models.Claude37Sonnet failed",
		)
	}

	if len(model.ImageFile) > 0 && len(model.PdfFiles) > 0 {
		return nil, errors.New(
			"only image file or pdf files can be provided, not both",
		)
	}

	if len(model.ImageFile) > 0 {
		return handleMedia(userMsg, model.ImageFile, nil), nil
	}

	if len(model.PdfFiles) > 0 {
		return handleMedia(userMsg, nil, model.PdfFiles), nil
	}

	return []anthropicMsg{
		{
			Role:    "user",
			Content: userMsg,
		},
	}, nil
}

func handleMedia(
	userMsg string,
	imageFile map[models.AnthropicImageType]string,
	pdfFiles []models.AnthropicPdf,
) []anthropicMsg {
	content := []any{}

	if len(imageFile) > 0 {
		mediaType := ""
		data := ""
		for t, val := range imageFile {
			mediaType = string(t)
			data = val
		}

		content = append(content, anthropicMediaPayload{
			Type: "image",
			Source: mediaSource{
				Type:      "base64",
				MediaType: mediaType,
				Data:      data,
			},
		})
	}

	if len(pdfFiles) > 0 {
		for _, pdfFile := range pdfFiles {
			content = append(content, anthropicMediaPayload{
				Type: "document",
				Source: mediaSource{
					Type:      "base64",
					MediaType: "application/pdf",
					Data:      string(pdfFile),
				},
			})
		}
	}

	content = append(content, anthropicTextPayload{
		Type: "text",
		Text: userMsg,
	})

	return []anthropicMsg{
		{
			Role:    "user",
			Content: content,
		},
	}
}
