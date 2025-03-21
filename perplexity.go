package heimdall

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

// completeResponse implements LLMProvider.
func (p Perplexity) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	requestLog *Logging,
) (CompletionResponse, error) {
	for i, key := range p.apiKeys {
		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		response, _, err := p.doRequest(ctx, req, client, nil, key)
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

	return p.tryWithBackup(ctx, req, client, nil, requestLog)
}

// doRequest implements LLMProvider.
func (p Perplexity) doRequest(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (CompletionResponse, int, error) {
	messages := make([]openAIRequestMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openAIRequestMessage(openAIRequestMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	apiReq := openAIRequest{
		Model:         req.Model.Name,
		Messages:      messages,
		Stream:        true,
		StreamOptions: streamOptions{IncludeUsage: true},
		Temperature:   1.0,
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return CompletionResponse{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		perplexityBaseUrl,
		bytes.NewReader(body))
	if err != nil {
		return CompletionResponse{}, 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

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
			return CompletionResponse{}, 0, fmt.Errorf("read line: %w", err)
		}

		line = strings.TrimPrefix(line, "data: ")
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		var chunk openAIChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return CompletionResponse{}, 0, fmt.Errorf(
				"unmarshal chunk: %w",
				err,
			)
		}

		if len(chunk.Choices) > 0 {
			fullContent.WriteString(chunk.Choices[0].Delta.Content)
		}

		chunks++
		if chunk.Usage.TotalTokens != 0 {
			usage = Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}
	}

	return CompletionResponse{
		Content: fullContent.String(),
		Model:   req.Model,
		Usage:   usage,
	}, 0, nil
}

// getApiKeys implements LLMProvider.
func (p Perplexity) getApiKeys() []string {
	panic("unimplemented")
}

// name implements LLMProvider.
func (p Perplexity) name() string {
	panic("unimplemented")
}

// streamResponse implements LLMProvider.
func (p Perplexity) streamResponse(
	ctx context.Context,
	client http.Client,
	req CompletionRequest,
	chunkHandler func(chunk string) error,
	requestLog *Logging,
) (CompletionResponse, error) {
	for i, key := range p.apiKeys {
		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with key_number: %v",
				i,
			),
		})
		response, _, err := p.doRequest(ctx, req, client, chunkHandler, key)
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

	return p.tryWithBackup(ctx, req, client, chunkHandler, requestLog)
}

// tryWithBackup implements LLMProvider.
func (p Perplexity) tryWithBackup(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *Logging,
) (CompletionResponse, error) {
	key := p.apiKeys[0]

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
			response, resCode, err := p.doRequest(
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

var _ LLMProvider = new(Perplexity)
