package heimdall

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
)

const vertexAIBaseURL = "https://us-east1-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent"

type Location string

const (
	UsEastFive    Location = "us-east5"
	UsSoutOne     Location = "us-south1"
	UsCentralOne  Location = "us-central1"
	UsWestFour    Location = "us-west4"
	UsEastOne     Location = "us-east1"
	UsEastFour    Location = "us-east4"
	UsWestOne     Location = "us-west1"
	EuWestFour    Location = "europe-west4"
	EuWestNine    Location = "europe-west9"
	EuWestOne     Location = "europe-west1"
	EuSoutWestOne Location = "europe-southwest1"
	EuWestEight   Location = "europe-west8"
	EuNorthOne    Location = "europe-north1"
	EuCentralTwo  Location = "europe-central2"
)

type googleVertexAI struct {
	client    http.Client
	clientTwo *genai.Client
}

func newGoogleVertexAI(c http.Client, genC *genai.Client) googleVertexAI {
	return googleVertexAI{c, genC}
}

// TODO: Implement manual key checking
func (g googleVertexAI) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	key APIKey,
) (CompletionResponse, error) {
	var parts []genai.Part
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "user" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "file" {
			parts = append(parts, genai.FileData{
				MIMEType: string(msg.FileType),
				FileURI:  msg.Content,
			})
		}
	}

	model := g.clientTwo.GenerativeModel(req.Model.Name)

	res, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return CompletionResponse{}, err
	}

	rb, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return CompletionResponse{}, err
	}

	return CompletionResponse{
		Content: string(rb),
		Model:   req.Model,
		Usage: Usage{
			PromptTokens: int(
				res.UsageMetadata.PromptTokenCount,
			),
			CompletionTokens: int(
				res.UsageMetadata.CandidatesTokenCount,
			),
			TotalTokens: int(
				res.UsageMetadata.TotalTokenCount,
			),
		},
	}, nil
}

func (g googleVertexAI) streamResponse(
	ctx context.Context,
	req CompletionRequest,
	key APIKey,
	chunkHandler func(chunk string) error,
) (CompletionResponse, error) {
	var parts []genai.Part
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "user" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "file" {
			parts = append(parts, genai.FileData{
				MIMEType: string(msg.FileType),
				FileURI:  msg.Content,
			})
		}
	}

	model := g.clientTwo.GenerativeModel(req.Model.Name)

	streamIter := model.GenerateContentStream(ctx, parts...)
	var fullContent strings.Builder
	var usage Usage

	chunks := 0
	now := time.Now()

	for {
		if chunks == 0 && time.Since(now).Seconds() > 3.0 {
			return CompletionResponse{}, context.Canceled
		}

		responseChunk, err := streamIter.Next()
		if err != nil {
			return CompletionResponse{}, err
		}

		if len(responseChunk.Candidates) > 0 {
			rb, err := json.MarshalIndent(responseChunk, "", "  ")
			if err != nil {
				return CompletionResponse{}, err
			}

			fullContent.WriteString(string(rb))
		}

		chunks++

		if responseChunk.Candidates[0].FinishReason == genai.FinishReasonStop {
			usage = Usage{
				PromptTokens: int(
					responseChunk.UsageMetadata.PromptTokenCount,
				),
				CompletionTokens: int(
					responseChunk.UsageMetadata.CandidatesTokenCount,
				),
				TotalTokens: int(
					responseChunk.UsageMetadata.TotalTokenCount,
				),
			}
			break
		}

		if responseChunk.Candidates[0].FinishReason != genai.FinishReasonStop {
			break
		}
	}

	return CompletionResponse{
		Content: fullContent.String(),
		Model:   req.Model,
		Usage:   usage,
	}, nil
}
