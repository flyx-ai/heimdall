package heimdall

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ModelNew string

type Google struct {
	apiKeys []string
	// client  http.Client
	models []ModelNew
}

func NewGoogle(keys []string) Google {
	return Google{keys, nil}
}

// completeResponse implements LLMProvider.
func (g Google) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	requestLog *Logging,
) (CompletionResponse, error) {
	maxRetries := 5
	initialBackoff := 100 * time.Millisecond
	maxBackoff := 10 * time.Second

	for i, key := range g.apiKeys {
		var lastErr error
		for attempt := range maxRetries {
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"attempting to complete request. attemp: %v, key_number: %v",
					attempt,
					i,
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
				response, err := g.doRequest(ctx, req, client, key)
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

				if !isRetryableError(response) {
					return CompletionResponse{}, err
				}

				lastErr = err

				backoff := min(initialBackoff*time.Duration(
					1<<attempt,
				), maxBackoff)

				jitter := time.Duration(
					float64(backoff) * (0.8 + 0.4*rand.Float64()),
				)

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

	return CompletionResponse{}, errors.New("could not complete request")
}

// getApiKeys implements LLMProvider.
func (g Google) getApiKeys() []string {
	return g.apiKeys
}

// name implements LLMProvider.
func (g Google) name() string {
	return string(ProviderGoogle)
}

// streamResponse implements LLMProvider.
func (g Google) streamResponse(
	ctx context.Context,
	req CompletionRequest,
	key APIKey,
	chunkHandler func(chunk string) error,
) (CompletionResponse, error) {
	panic("unimplemented")
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if httpErr, ok := err.(*url.Error); ok {
		if httpErr.Temporary() {
			return true
		}
	}

	if strings.Contains(err.Error(), "429") ||
		strings.Contains(err.Error(), "500") ||
		strings.Contains(err.Error(), "503") {
		return true
	}

	return false
}

func (g *Google) doRequest(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	key string,
) (CompletionResponse, error) {
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
		return CompletionResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf(googleBaseUrl, req.Model.Name, key),
		bytes.NewReader(body))
	if err != nil {
		return CompletionResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		return CompletionResponse{}, errInternalServer
	case http.StatusTooManyRequests:
		return CompletionResponse{}, errTooManyRequests
	case http.StatusBadRequest:
		return CompletionResponse{}, errBadRequest
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder
	var usage Usage
	chunks := 0
	now := time.Now()

	for {
		if chunks == 0 && time.Since(now).Seconds() > 3.0 {
			return CompletionResponse{}, context.Canceled
		}
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return CompletionResponse{}, err
		}

		line = strings.TrimPrefix(line, "data: ")
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		var responseChunk geminiResponse
		if err := json.Unmarshal([]byte(line), &responseChunk); err != nil {
			return CompletionResponse{}, err
		}

		if len(responseChunk.Candidates) > 0 {
			fullContent.WriteString(
				responseChunk.Candidates[0].Content.Parts[0].Text,
			)
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
	}, nil
}

var _ LLMProvider = new(Google)
