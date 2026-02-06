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

const grokBaseURL = "https://api.x.ai/v1"

type Grok struct {
	apiKeys []string
}

func NewGrok(apiKeys []string) Grok {
	return Grok{
		apiKeys: apiKeys,
	}
}

func (g Grok) Name() string {
	return models.GrokProvider
}

func (g Grok) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	model := req.Model.GetName()

	grokRequest := openAIRequest{
		Model:         model,
		Stream:        true,
		StreamOptions: streamOptions{IncludeUsage: true},
		Temperature:   1.0,
	}

	var structuredOutput map[string]any
	switch m := req.Model.(type) {
	case models.Grok2Vision:
		structuredOutput = m.StructuredOutput
	case models.Grok3:
		structuredOutput = m.StructuredOutput
	case models.Grok3Mini:
		structuredOutput = m.StructuredOutput
	case models.Grok3Fast:
		structuredOutput = m.StructuredOutput
	case models.Grok3MiniFast:
		structuredOutput = m.StructuredOutput
	case models.Grok4:
		structuredOutput = m.StructuredOutput
	case models.Grok4Fast:
		structuredOutput = m.StructuredOutput
	}

	if len(structuredOutput) > 0 {
		grokRequest.ResponseFormat = map[string]any{
			"type":        "json_schema",
			"json_schema": structuredOutput,
		}
	}

	request, err := prepareGrokRequest(
		grokRequest,
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
		fmt.Sprintf("%s/chat/completions", grokBaseURL),
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
		return response.Completion{}, 0, fmt.Errorf("marshal raw response events: %w", err)
	}

	return response.Completion{
		Content:     fullContent.String(),
		Model:       req.Model.GetName(),
		Usage:       usage,
		RawRequest:  body,
		RawResponse: rawResp,
	}, 0, nil
}

func (g Grok) tryWithBackup(
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
				"attempting to complete request with exponential backoff. attempt: %v",
				attempt,
			),
		})

		select {
		case <-ctx.Done():
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"context was cancelled with error: %v",
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

func (g Grok) CompleteResponse(
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

func (g Grok) StreamResponse(
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

func prepareGrokRequest(
	request openAIRequest,
	requestedModel models.Model,
	systemInst string,
	userMsg string,
	history []request.Message,
) (openAIRequest, error) {
	switch m := requestedModel.(type) {
	case *models.Grok2Vision:
		return prepareGrokVisionRequest(request, m.ImageFile, systemInst, userMsg, history)
	case *models.Grok3:
		return prepareGrokVisionRequest(request, m.ImageFile, systemInst, userMsg, history)
	case *models.Grok3Fast:
		return prepareGrokVisionRequest(request, m.ImageFile, systemInst, userMsg, history)
	case *models.Grok4:
		return prepareGrokVisionRequest(request, m.ImageFile, systemInst, userMsg, history)
	case *models.Grok4Fast:
		return prepareGrokVisionRequest(request, m.ImageFile, systemInst, userMsg, history)
	default:
		return prepareBasicMessages(request, systemInst, userMsg, history)
	}
}

func prepareGrokVisionRequest(
	request openAIRequest,
	imageFiles []models.GrokImagePayload,
	systemInst string,
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

	if systemInst != "" {
		reqMsgWithImage = append(reqMsgWithImage, requestMessageWithImage{
			Role: "system",
			Content: []any{
				fileInputMessage{
					Type: "text",
					Text: systemInst,
				},
			},
		})
	}

	lastIndex := len(reqMsgWithImage)
	reqMsgWithImage = append(reqMsgWithImage, requestMessageWithImage{
		Role:    "user",
		Content: []any{},
	})

	for _, img := range imageFiles {
		detail := "auto"
		if img.Detail != "" {
			detail = img.Detail
		}

		ii := imageInput{
			Type: "image_url",
			ImageURL: imageURL{
				URL:    img.URL,
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

var _ LLMProvider = new(Grok)
