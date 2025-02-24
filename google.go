package heimdall

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const googleBaseUrl = "https://generativelanguage.googleapis.com/v1beta:chatCompletions"

type googleRequest struct {
	Contents []googleContent `json:"contents"`
}

type googleContent struct {
	Role  string              `json:"role,omitempty"`
	Parts []googleContentPart `json:"parts"`
}

type googleContentPart struct {
	Text string `json:"text"`
}

type googleStreamResponse struct {
	Candidates []googleCandidate `json:"candidates"`
}

type googleCandidate struct {
	Content      googleContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type Google struct {
	Client http.Client
}

// TODO: Implement manual key checking
func (g Google) StreamResponse(
	ctx context.Context,
	req CompletionRequest,
	key APIKey,
	chunkHandler func(chunk string) error,
) (*CompletionResponse, error) {
	messages := make([]openAIRequestMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openAIRequestMessage(msg)
	}

	apiReq := openAIRequest{
		Model:         req.Model.Name,
		Messages:      messages,
		Stream:        true,
		StreamOptions: streamOptions{IncludeUsage: true},
		Temperature:   req.Temperature,
		TopP:          req.TopP,
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		googleBaseUrl,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key.Key)

	resp, err := g.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		fmt.Println(key, ":", values)
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return nil, ErrRateLimitHit
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder
	var usage Usage
	chunks := 0
	now := time.Now()

	for {
		if chunks == 0 && time.Since(now).Seconds() > 3.0 {
			return nil, context.Canceled
		}
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read line: %w", err)
		}

		line = strings.TrimPrefix(line, "data: ")
		line = strings.TrimSpace(line)
		if line == "" || line == "[DONE]" {
			continue
		}

		var chunk openAIChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return nil, fmt.Errorf("unmarshal chunk: %w", err)
		}

		if len(chunk.Choices) > 0 {
			fullContent.WriteString(chunk.Choices[0].Delta.Content)
			if err := chunkHandler(chunk.Choices[0].Delta.Content); err != nil {
				return nil, fmt.Errorf("handle chunk: %w", err)
			}
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

	return &CompletionResponse{
		Content: fullContent.String(),
		Model:   req.Model,
		Usage:   usage,
	}, nil
}
