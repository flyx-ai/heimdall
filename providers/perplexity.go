//go:build perplexity

package providers

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

const perplexityBaseUrl = "https://api.perplexity.ai/chat/completions"

type Perplexity struct {
	apiKeys []string
}

func NewPerplexity(apiKeys []string) Perplexity {
	return Perplexity{
		apiKeys,
	}
}

// CompleteResponse implements LLMProvider.
func (p Perplexity) CompleteResponse(
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

	for i, key := range p.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		res, _, err := p.doRequest(ctx, req, client, nil, key)
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

	return p.tryWithBackup(ctx, req, client, nil, reqLog)
}

// doRequest implements LLMProvider.
func (p Perplexity) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	hisLen := len(req.History)
	requestMessages := make([]requestMessage, hisLen+2)
	for i, his := range req.History {
		requestMessages[i] = requestMessage(requestMessage{
			Role:    his.Role,
			Content: his.Content,
		})
	}

	if hisLen == 0 {
		requestMessages[0] = requestMessage(requestMessage{
			Role:    "system",
			Content: req.SystemMessage,
		})
		requestMessages[1] = requestMessage(requestMessage{
			Role:    "user",
			Content: req.UserMessage,
		})
	}
	if hisLen != 0 {
		requestMessages[hisLen+1] = requestMessage(requestMessage{
			Role:    "system",
			Content: req.SystemMessage,
		})
		requestMessages[hisLen+2] = requestMessage(requestMessage{
			Role:    "user",
			Content: req.UserMessage,
		})
	}

	apiReq := openAIRequest{
		Model:         req.Model.GetName(),
		Messages:      requestMessages,
		Stream:        true,
		StreamOptions: streamOptions{IncludeUsage: true},
		Temperature:   1.0,
	}

	var structuredOutput map[string]any
	switch m := req.Model.(type) {
	case models.SonarReasoningPro:
		structuredOutput = m.StructuredOutput
	case models.SonarReasoning:
		structuredOutput = m.StructuredOutput
	case models.SonarPro:
		structuredOutput = m.StructuredOutput
	case models.Sonar:
		structuredOutput = m.StructuredOutput
	}

	if len(structuredOutput) > 0 {
		apiReq.ResponseFormat = map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"schema": structuredOutput,
			},
		}
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return response.Completion{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		perplexityBaseUrl,
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
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			return response.Completion{}, resp.StatusCode, fmt.Errorf(
				"perplexity returned status %d: %s",
				resp.StatusCode,
				string(bodyBytes),
			)
		}
		return response.Completion{}, resp.StatusCode, fmt.Errorf(
			"perplexity returned status %d",
			resp.StatusCode,
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
			contentDelta := chunk.Choices[0].Delta.Content
			fullContent.WriteString(contentDelta)

			if chunkHandler != nil {
				if err := chunkHandler(contentDelta); err != nil {
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

	finalContent := fullContent.String()

	return response.Completion{
		Content: finalContent,
		Model:   req.Model.GetName(),
		Usage:   usage,
	}, 0, nil
}

func (p Perplexity) Name() string {
	return models.PerplexityProvider
}

// StreamResponse implements LLMProvider.
func (p Perplexity) StreamResponse(
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

	for i, key := range p.apiKeys {
		reqLog.Events = append(reqLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		res, _, err := p.doRequest(ctx, req, client, chunkHandler, key)
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

	return p.tryWithBackup(ctx, req, client, chunkHandler, reqLog)
}

// tryWithBackup implements LLMProvider.
func (p Perplexity) tryWithBackup(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	key := p.apiKeys[0]

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
			res, resCode, err := p.doRequest(
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

var _ LLMProvider = new(Perplexity)
