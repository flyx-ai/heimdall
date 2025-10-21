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

func TestVertexAIModelsWithCompletion(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}

	projectID := os.Getenv("VERTEX_PROJECT_ID")
	apiKey := os.Getenv("GOOGLE_API_KEY")

	vertexai, err := providers.NewVertexAI(
		context.Background(),
		projectID,
		apiKey,
	)
	require.NoError(t, err, "error creating VertexAI provider", "error", err)

	systemInst := "you are a helpful assistant."
	userMsg := "please make a detailed analysis of the NVIDIA's current valuation."

	tests := []struct {
		name string
		req  request.Completion
	}{
		// NOTE: Gemini 1.5 test cases removed as models were retired by Google in 2025
		{
			name: "should complete request with VertexGemini20Flash",
			req: request.Completion{
				Model:         models.VertexGemini20Flash{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with VertexGemini20FlashLite",
			req: request.Completion{
				Model:         models.VertexGemini20FlashLite{},
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

			assert.NotEmpty(t, res.Content, "content should not be empty")
			assert.NotEmpty(t, res.Model, "model should not be empty")
		})
	}
}

func TestVertexAIModelsWithStreaming(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}

	projectID := os.Getenv("VERTEX_PROJECT_ID")
	apiKey := os.Getenv("GOOGLE_API_KEY")

	vertexai, err := providers.NewVertexAI(
		context.Background(),
		projectID,
		apiKey,
	)
	require.NoError(t, err, "error creating VertexAI provider", "error", err)

	systemInst := "you are a helpful assistant."
	userMsg := "please make a detailed analysis of the NVIDIA's current valuation."

	tests := []struct {
		name string
		req  request.Completion
	}{
		// NOTE: Gemini 1.5 test cases removed as models were retired by Google in 2025
		{
			name: "should stream request with VertexGemini20Flash",
			req: request.Completion{
				Model:         models.VertexGemini20Flash{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with VertexGemini20FlashLite",
			req: request.Completion{
				Model:         models.VertexGemini20FlashLite{},
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
			var chunkHandlerCollection string
			res, err := vertexai.StreamResponse(
				context.Background(),
				client,
				tt.req,
				func(chunk string) error {
					chunkHandlerCollection = chunkHandlerCollection + chunk
					return nil
				},
				nil,
			)
			require.NoError(
				t,
				err,
				"vertex StreamResponse returned an error",
			)

			assert.NotEmpty(
				t,
				chunkHandlerCollection,
				"chunkHandlerCollection should not be empty",
			)
			assert.NotEmpty(t, res.Content, "content should not be empty")
			assert.NotEmpty(t, res.Model, "model should not be empty")
		})
	}
}
