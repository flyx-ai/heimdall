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

const openAIBaseURL = "https://api.openai.com/v1"

type openAIRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAIRequest struct {
	Model         string                 `json:"model"`
	Messages      []openAIRequestMessage `json:"messages"`
	Stream        bool                   `json:"stream"`
	StreamOptions streamOptions          `json:"stream_options"`
	Temperature   float32                `json:"temperature,omitempty"`
	TopP          float32                `json:"top_p,omitempty"`
	// ResponseFormat map[string]interface{} `json:"response_format",omitempty`
}

type Openai struct {
	Client http.Client
}

func (oa Openai) StreamResponse(
	ctx context.Context,
	req CompletionRequest,
	key string,
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
		fmt.Sprintf("%s/chat/completions", openAIBaseURL),
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

	resp, err := oa.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
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
