package heimdall

import (
	"context"
	"net/http"
	"time"
)

// const openAIBaseURL = "https://api.openai.com/v1"
//
// type openAIRequestMessage struct {
// 	Role    string `json:"role"`
// 	Content string `json:"content"`
// }
//
// type openAIChunk struct {
// 	Choices []struct {
// 		Delta struct {
// 			Content string `json:"content"`
// 		} `json:"delta"`
// 	} `json:"choices"`
// 	Usage struct {
// 		PromptTokens     int `json:"prompt_tokens"`
// 		CompletionTokens int `json:"completion_tokens"`
// 		TotalTokens      int `json:"total_tokens"`
// 	} `json:"usage"`
// }
//
// type streamOptions struct {
// 	IncludeUsage bool `json:"include_usage"`
// }
//
// type openAIRequest struct {
// 	Model         string                 `json:"model"`
// 	Messages      []openAIRequestMessage `json:"messages"`
// 	Stream        bool                   `json:"stream"`
// 	StreamOptions streamOptions          `json:"stream_options"`
// 	Temperature   float32                `json:"temperature,omitempty"`
// 	TopP          float32                `json:"top_p,omitempty"`
// }

type Openai struct {
	client http.Client
}

// doRequest implements LLMProvider.
func (oa Openai) doRequest(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (CompletionResponse, int, error) {
	panic("unimplemented")
}

// getApiKeys implements LLMProvider.
func (oa Openai) getApiKeys() []string {
	panic("unimplemented")
}

// name implements LLMProvider.
func (oa Openai) name() string {
	panic("unimplemented")
}

// tryWithBackup implements LLMProvider.
func (oa Openai) tryWithBackup(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *Logging,
) (CompletionResponse, error) {
	panic("unimplemented")
}

type RateLimit struct {
	Remaining int
	Limit     int
	Reset     time.Time
}

func (oa Openai) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	client http.Client,
	requestLog *Logging,
) (CompletionResponse, error) {
	// messages := make([]openAIRequestMessage, len(req.Messages))
	// for i, msg := range req.Messages {
	// 	messages[i] = openAIRequestMessage(openAIRequestMessage{
	// 		Role:    msg.Role,
	// 		Content: msg.Content,
	// 	})
	// }
	//
	// apiReq := openAIRequest{
	// 	Model:         req.Model.Name,
	// 	Messages:      messages,
	// 	Stream:        true,
	// 	StreamOptions: streamOptions{IncludeUsage: true},
	// 	Temperature:   1.0,
	// }
	//
	// body, err := json.Marshal(apiReq)
	// if err != nil {
	// 	return CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	// }
	//
	// httpReq, err := http.NewRequestWithContext(ctx, "POST",
	// 	fmt.Sprintf("%s/chat/completions", openAIBaseURL),
	// 	bytes.NewReader(body))
	// if err != nil {
	// 	return CompletionResponse{}, fmt.Errorf("create request: %w", err)
	// }
	//
	// httpReq.Header.Set("Content-Type", "application/json")
	// httpReq.Header.Set("Authorization", "Bearer "+key.Key)
	//
	// resp, err := oa.client.Do(httpReq)
	// if err != nil {
	// 	return CompletionResponse{}, fmt.Errorf("do request: %w", err)
	// }
	// defer resp.Body.Close()
	//
	// rateLimit := parseOpenAICompatRateLimit(resp)
	//
	// used := rateLimit.Limit - rateLimit.Remaining
	// remaining := rateLimit.Remaining
	// reset := rateLimit.Reset
	//
	// if key.requestsUsed < used {
	// 	key.requestsUsed = used
	// }
	//
	// if key.RequestRemaining > remaining {
	// 	key.RequestRemaining = remaining
	// }
	//
	// // TODO: fix this logic
	// if key.ResetAt.Before(reset) {
	// 	key.ResetAt = rateLimit.Reset
	// }
	//
	// switch resp.StatusCode {
	// case http.StatusTooManyRequests:
	// 	return CompletionResponse{}, ErrRateLimitHit
	// }
	//
	// reader := bufio.NewReader(resp.Body)
	// var fullContent strings.Builder
	// var usage Usage
	// chunks := 0
	// now := time.Now()
	//
	// for {
	// 	if chunks == 0 && time.Since(now).Seconds() > 3.0 {
	// 		return CompletionResponse{}, context.Canceled
	// 	}
	// 	line, err := reader.ReadString('\n')
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	if err != nil {
	// 		return CompletionResponse{}, fmt.Errorf("read line: %w", err)
	// 	}
	//
	// 	line = strings.TrimPrefix(line, "data: ")
	// 	line = strings.TrimSpace(line)
	// 	if line == "" || line == "[DONE]" {
	// 		continue
	// 	}
	//
	// 	var chunk openAIChunk
	// 	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
	// 		return CompletionResponse{}, fmt.Errorf("unmarshal chunk: %w", err)
	// 	}
	//
	// 	if len(chunk.Choices) > 0 {
	// 		fullContent.WriteString(chunk.Choices[0].Delta.Content)
	// 	}
	//
	// 	chunks++
	// 	if chunk.Usage.TotalTokens != 0 {
	// 		usage = Usage{
	// 			PromptTokens:     chunk.Usage.PromptTokens,
	// 			CompletionTokens: chunk.Usage.CompletionTokens,
	// 			TotalTokens:      chunk.Usage.TotalTokens,
	// 		}
	// 	}
	// }
	//

	return CompletionResponse{
		// Content: fullContent.String(),
		// Model:   req.Model,
		// Usage:   usage,
	}, nil
}

func (oa Openai) streamResponse(
	ctx context.Context,
	client http.Client,
	req CompletionRequest,
	chunkHandler func(chunk string) error,
	requestLog *Logging,
) (CompletionResponse, error) {
	// messages := make([]openAIRequestMessage, len(req.Messages))
	// for i, msg := range req.Messages {
	// 	messages[i] = openAIRequestMessage(openAIRequestMessage{
	// 		Role:    msg.Role,
	// 		Content: msg.Content,
	// 	})
	// }
	//
	// apiReq := openAIRequest{
	// 	Model:         req.Model.Name,
	// 	Messages:      messages,
	// 	Stream:        true,
	// 	StreamOptions: streamOptions{IncludeUsage: true},
	// 	Temperature:   req.Temperature,
	// 	TopP:          req.TopP,
	// }
	//
	// body, err := json.Marshal(apiReq)
	// if err != nil {
	// 	return CompletionResponse{}, err
	// }
	//
	// httpReq, err := http.NewRequestWithContext(ctx, "POST",
	// 	fmt.Sprintf("%s/chat/completions", openAIBaseURL),
	// 	bytes.NewReader(body))
	// if err != nil {
	// 	return CompletionResponse{}, err
	// }
	//
	// httpReq.Header.Set("Content-Type", "application/json")
	// httpReq.Header.Set("Authorization", "Bearer "+key.Key)
	//
	// resp, err := oa.client.Do(httpReq)
	// if err != nil {
	// 	return CompletionResponse{}, err
	// }
	//
	// defer resp.Body.Close()
	//
	// rateLimit := parseOpenAICompatRateLimit(resp)
	//
	// used := rateLimit.Limit - rateLimit.Remaining
	// remaining := rateLimit.Remaining
	// reset := rateLimit.Reset
	//
	// if key.requestsUsed < used {
	// 	key.requestsUsed = used
	// }
	//
	// if key.RequestRemaining > remaining {
	// 	key.RequestRemaining = remaining
	// }
	//
	// // TODO: fix this logic
	// if key.ResetAt.Before(reset) {
	// 	key.ResetAt = rateLimit.Reset
	// }
	//
	// switch resp.StatusCode {
	// case http.StatusTooManyRequests:
	// 	return CompletionResponse{}, ErrRateLimitHit
	// }
	//
	// reader := bufio.NewReader(resp.Body)
	// var fullContent strings.Builder
	// var usage Usage
	// chunks := 0
	// now := time.Now()
	//
	// for {
	// 	if chunks == 0 && time.Since(now).Seconds() > 3.0 {
	// 		return CompletionResponse{}, context.Canceled
	// 	}
	// 	line, err := reader.ReadString('\n')
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	if err != nil {
	// 		return CompletionResponse{}, err
	// 	}
	//
	// 	line = strings.TrimPrefix(line, "data: ")
	// 	line = strings.TrimSpace(line)
	// 	if line == "" || line == "[DONE]" {
	// 		continue
	// 	}
	//
	// 	var chunk openAIChunk
	// 	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
	// 		return CompletionResponse{}, err
	// 	}
	//
	// 	if len(chunk.Choices) > 0 {
	// 		fullContent.WriteString(chunk.Choices[0].Delta.Content)
	// 		if err := chunkHandler(chunk.Choices[0].Delta.Content); err != nil {
	// 			return CompletionResponse{}, err
	// 		}
	// 	}
	//
	// 	chunks++
	// 	if chunk.Usage.TotalTokens != 0 {
	// 		usage = Usage{
	// 			PromptTokens:     chunk.Usage.PromptTokens,
	// 			CompletionTokens: chunk.Usage.CompletionTokens,
	// 			TotalTokens:      chunk.Usage.TotalTokens,
	// 		}
	// 	}
	// }
	//

	return CompletionResponse{
		// Content: fullContent.String(),
		// Model:   req.Model,
		// Usage:   usage,
	}, nil
}

var _ LLMProvider = new(Openai)
