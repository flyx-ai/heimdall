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

	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

const anthropicBaseUrl = "https://api.anthropic.com/v1"

type Anthropic struct {
	apiKeys []string
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

type anthropicStreamResponse struct {
	Type    string `json:"type"`
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	ID           string  `json:"id"`
	Model        string  `json:"model"`
	Role         string  `json:"role"`
	StopReason   string  `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Delta struct {
		Type    string `json:"type"`
		Text    string `json:"text"`
		Message string `json:"message"`
	} `json:"delta"`
}

// CompleteResponse implements LLMProvider.
func (a Anthropic) CompleteResponse(
	ctx context.Context,
	req request.CompletionRequest,
	client http.Client,
	requestLog *response.Logging,
) (response.CompletionResponse, error) {
	for i, key := range a.apiKeys {
		requestLog.Events = append(requestLog.Events, response.Event{
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
		requestLog.Events = append(requestLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"request could not be completed, err: %v",
				err,
			),
		})
	}

	return a.tryWithBackup(ctx, req, client, nil, requestLog)
}

// doRequest implements LLMProvider.
func (a Anthropic) doRequest(
	ctx context.Context,
	req request.CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.CompletionResponse, int, error) {
	var systemMsg string
	messages := []anthropicMsg{}
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemMsg = msg.Content
		}
		if msg.Role == "user" || msg.Role == "assistant" {
			messages = append(messages, anthropicMsg{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	apiReq := anthropicRequest{
		System:      systemMsg,
		Model:       req.Model.GetName(),
		Messages:    messages,
		Stream:      true,
		MaxTokens:   4096,
		Temperature: 1.0,
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return response.CompletionResponse{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/messages", anthropicBaseUrl),
		bytes.NewReader(body))
	if err != nil {
		return response.CompletionResponse{}, 0, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", key)
	httpReq.Header.Set("Anthropic-Version", "2023-06-01")

	resp, err := client.Do(httpReq)
	if err != nil {
		return response.CompletionResponse{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response.CompletionResponse{}, resp.StatusCode, err
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder
	var lastResponse *anthropicStreamResponse
	chunks := 0
	now := time.Now()

	for {
		if chunks == 0 && time.Since(now).Seconds() > 3.0 {
			return response.CompletionResponse{}, 0, context.Canceled
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return response.CompletionResponse{}, 0, err
		}

		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		line = bytes.TrimSpace(line)

		if bytes.HasPrefix(line, []byte("event:")) {
			continue
		}

		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))

		if bytes.Equal(bytes.TrimSpace(data), []byte("[DONE]")) {
			break
		}

		var chunk anthropicStreamResponse
		if err := json.Unmarshal(data, &chunk); err != nil {
			return response.CompletionResponse{}, 0, err
		}

		switch chunk.Type {
		case "message_start":
			lastResponse = &chunk
		case "content_block_start":
			continue
		case "content_block_delta":
			if chunk.Delta.Text != "" {
				fullContent.WriteString(chunk.Delta.Text)
				if chunkHandler != nil {
					if err := chunkHandler(chunk.Delta.Text); err != nil {
						return response.CompletionResponse{}, 0, err
					}
				}
			}
		case "message_delta":
			if lastResponse != nil {
				lastResponse.Usage = chunk.Usage
				lastResponse.StopReason = chunk.StopReason
			}
		case "message_stop":
			if lastResponse != nil {
				lastResponse.Usage = chunk.Usage
				lastResponse.StopReason = chunk.StopReason
			}
		}

		chunks++
	}

	return response.CompletionResponse{
		Content: fullContent.String(),
		Model:   req.Model.GetName(),
		Usage: response.Usage{
			CompletionTokens: lastResponse.Usage.OutputTokens,
			PromptTokens:     lastResponse.Usage.InputTokens,
		},
	}, 0, nil
}

// getApiKeys implements LLMProvider.
func (a Anthropic) GetApiKeys() []string {
	return a.apiKeys
}

// name implements LLMProvider.
func (a Anthropic) Name() string {
	return "anthropic"
}

// StreamResponse implements LLMProvider.
func (a Anthropic) StreamResponse(
	ctx context.Context,
	client http.Client,
	req request.CompletionRequest,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.CompletionResponse, error) {
	for i, key := range a.apiKeys {
		requestLog.Events = append(requestLog.Events, response.Event{
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
		requestLog.Events = append(requestLog.Events, response.Event{
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
	req request.CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.CompletionResponse, error) {
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
			return response.CompletionResponse{}, ctx.Err()
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
				return response.CompletionResponse{}, err
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
				return response.CompletionResponse{}, ctx.Err()
			case <-timer.C:
				continue
			}
		}
	}

	return response.CompletionResponse{}, fmt.Errorf(
		"max retries exceeded: %w",
		lastErr,
	)
}

var _ LLMProvider = new(Anthropic)
