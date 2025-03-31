package providers_test

import (
	"context"
	"log/slog"
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

func TestVertexAIModelsWithCompletion(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}

	projectID := os.Getenv("VERTEX_PROJECT_ID")
	location := "us-west-1"
	credentialsJSON := os.Getenv("VERTEX_AI_KEY")

	vertexai, err := providers.NewVertexAI(
		context.Background(),
		projectID,
		location,
		credentialsJSON,
	)
	require.NoError(t, err, "error creating VertexAI provider", "error", err)

	msgs := []request.Message{
		{
			Role:    "system",
			Content: "you are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: "please make a detailed analysis of the NVIDIA's current valuation.",
		},
	}

	tests := []struct {
		name string
		req  request.Completion
	}{
		{
			name: "should complete request with VertexGemini15Flash",
			req: request.Completion{
				Model:       models.VertexGemini15FlashThinking{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with VertexGemini15Pro",
			req: request.Completion{
				Model:       models.VertexGemini15Pro{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with VertexGemini20Flash",
			req: request.Completion{
				Model:       models.VertexGemini20Flash{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with VertexGemini20FlashLite",
			req: request.Completion{
				Model:       models.VertexGemini20FlashLite{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := vertexai.CompleteResponse(
				context.Background(),
				tt.req,
				client,
				nil,
			)
			require.NoError(
				t,
				err,
				"vertex complete response returned an error",
			)

			slog.Info(
				"USAGE",
				"tkns",
				res.Usage.TotalTokens,
				"pt",
				res.Usage.PromptTokens,
				"ct",
				res.Usage.CompletionTokens,
			)
			assert.NotEmpty(t, res.Content, "content should not be empty")
		})
	}
}
