package heimdall

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const googleBaseUrl = "https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s"

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

type google struct {
	client http.Client
}

// Response represents the top-level response from the Gemini API
type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata usageMetadata     `json:"usageMetadata"`
	ModelVersion  string            `json:"modelVersion"`
}

// Candidate represents a response candidate
type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

// Content represents the content of a candidate
type geminiContent struct {
	Parts []geminiResponsePart `json:"parts"`
	Role  string               `json:"role"`
}

// Part represents a part of the content
type geminiResponsePart struct {
	Text string `json:"text"`
}

// UsageMetadata represents token usage information
type usageMetadata struct {
	PromptTokenCount        int             `json:"promptTokenCount"`
	CandidatesTokenCount    int             `json:"candidatesTokenCount"`
	TotalTokenCount         int             `json:"totalTokenCount"`
	PromptTokensDetails     []tokensDetails `json:"promptTokensDetails"`
	CandidatesTokensDetails []tokensDetails `json:"candidatesTokensDetails"`
}

// TokensDetails represents token details by modality
type tokensDetails struct {
	Modality   string `json:"modality"`
	TokenCount int    `json:"tokenCount"`
}

// TODO: Implement manual key checking
func (g google) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	key APIKey,
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
		fmt.Sprintf(googleBaseUrl, req.Model.Name, key.Key),
		bytes.NewReader(body))
	if err != nil {
		return CompletionResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return CompletionResponse{}, err
	case http.StatusBadRequest:
		return CompletionResponse{}, err
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
			slog.ErrorContext(ctx, "UNMARSHAL ERRRORRRR", "err", err)
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

// TODO: Implement manual key checking
func (g google) streamResponse(
	ctx context.Context,
	req CompletionRequest,
	key APIKey,
	chunkHandler func(chunk string) error,
) (*CompletionResponse, error) {
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
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf(googleBaseUrl, req.Model.Name, key.Key),
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
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
