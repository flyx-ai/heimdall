package providers_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/providers"
	"github.com/flyx-ai/heimdall/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelsWithHistory(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	anthropicProvider := providers.NewAnthropic(
		[]string{os.Getenv("ANTHROPIC_API_KEY")},
	)
	google := providers.NewGoogle([]string{os.Getenv("GOOGLE_API_KEY")})
	openai := providers.NewOpenAI([]string{os.Getenv("OPENAI_API_KEY")})
	perplexity := providers.NewPerplexity(
		[]string{os.Getenv("PERPLEXITY_API_KEY")},
	)
	projectID := os.Getenv("VERTEX_PROJECT_ID")
	location := "us-west-1"
	credentialsJSON := os.Getenv("VERTEX_AI_KEY")

	vertexai, err := providers.NewVertexAI(
		context.Background(),
		projectID,
		location,
		credentialsJSON,
	)
	require.NoError(
		t,
		err,
		"setting up vertexai returned an unexpected error",
		"error",
		err,
	)

	systemInst := "you are a helpful assistant."
	userMsg := "please make a detailed analysis of the NVIDIA's current valuation."

	tests := []struct {
		name string
		req  request.Completion
	}{
		{
			name: "should complete request with claude-3-haiku",
			req: request.Completion{
				Model:         models.Claude35Haiku{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				History: []request.Message{
					{
						Role:    "model",
						Content: "i was a very helpful assistant",
					},
				},
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with claude-35-sonnet",
			req: request.Completion{
				Model:         models.Claude35Sonnet{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with claude-3-opus",
			req: request.Completion{
				Model:         models.Claude3Opus{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with claude-37-sonnet",
			req: request.Completion{
				Model:         models.Claude37Sonnet{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := anthropicProvider.CompleteResponse(
				context.Background(),
				tt.req,
				client,
				nil,
			)
			require.NoError(
				t,
				err,
				"CompleteResponse returned an unexpected error",
				"error",
				err,
			)

			assert.NotEmpty(t, res.Content, "content should not be empty")
			assert.NotEmpty(t, res.Model, "model should not be empty")
		})
	}
}
