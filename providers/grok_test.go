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

func TestGrokModelsWithCompletion(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	grokProvider := providers.NewGrok([]string{os.Getenv("GROK_API_KEY")})

	systemInst := "you are a helpful assistant."
	userMsg := "please make a detailed analysis of the NVIDIA's current valuation."

	tests := []struct {
		name string
		req  request.Completion
	}{
		{
			name: "should complete request with grok-3",
			req: request.Completion{
				Model:         models.Grok3{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with grok-3-mini",
			req: request.Completion{
				Model:         models.Grok3Mini{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with grok-3-fast",
			req: request.Completion{
				Model:         models.Grok3Fast{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with grok-3-mini-fast",
			req: request.Completion{
				Model:         models.Grok3MiniFast{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with grok-4",
			req: request.Completion{
				Model:         models.Grok4{},
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
			res, err := grokProvider.CompleteResponse(
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

			assert.NotEmpty(t, res.Content, "Expected non-empty content")
			assert.Equal(t, tt.req.Model.GetName(), res.Model, "Model mismatch")
			assert.Positive(
				t,
				res.Usage.TotalTokens,
				"Expected positive token usage",
			)
		})
	}
}

func TestGrokModelsWithStreaming(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	grokProvider := providers.NewGrok([]string{os.Getenv("GROK_API_KEY")})

	systemInst := "you are a helpful assistant."
	userMsg := "please make a detailed analysis of the NVIDIA's current valuation."

	tests := []struct {
		name string
		req  request.Completion
	}{
		{
			name: "should stream request with grok-3",
			req: request.Completion{
				Model:         models.Grok3{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with grok-3-mini",
			req: request.Completion{
				Model:         models.Grok3Mini{},
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
			var chunks []string
			chunkHandler := func(chunk string) error {
				chunks = append(chunks, chunk)
				return nil
			}

			res, err := grokProvider.StreamResponse(
				context.Background(),
				client,
				tt.req,
				chunkHandler,
				nil,
			)
			require.NoError(
				t,
				err,
				"StreamResponse returned an unexpected error",
				"error",
				err,
			)

			assert.NotEmpty(t, res.Content, "Expected non-empty content")
			assert.Equal(t, tt.req.Model.GetName(), res.Model, "Model mismatch")
			assert.Positive(
				t,
				res.Usage.TotalTokens,
				"Expected positive token usage",
			)
			assert.NotEmpty(t, chunks, "Expected streaming chunks")
		})
	}
}

func TestGrok2VisionWithImage(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	grokProvider := providers.NewGrok([]string{os.Getenv("GROK_API_KEY")})

	systemInst := "you are a helpful assistant that analyzes images."
	userMsg := "What is in this image?"

	grok2Vision := &models.Grok2Vision{
		ImageFile: []models.GrokImagePayload{
			{
				URL:    "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/640px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg",
				Detail: "high",
			},
		},
	}

	req := request.Completion{
		Model:         grok2Vision,
		SystemMessage: systemInst,
		UserMessage:   userMsg,
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	res, err := grokProvider.CompleteResponse(
		context.Background(),
		req,
		client,
		nil,
	)
	require.NoError(
		t,
		err,
		"CompleteResponse returned an unexpected error for vision model",
		"error",
		err,
	)

	assert.NotEmpty(
		t,
		res.Content,
		"Expected non-empty content for vision response",
	)
	assert.Equal(
		t,
		models.Grok2VisionAlias,
		res.Model,
		"Model mismatch for vision model",
	)
	assert.Positive(
		t,
		res.Usage.TotalTokens,
		"Expected positive token usage for vision model",
	)
}

func TestGrokWithHistory(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	grokProvider := providers.NewGrok([]string{os.Getenv("GROK_API_KEY")})

	systemInst := "you are a helpful assistant."
	userMsg := "What was my previous question about?"

	history := []request.Message{
		{
			Role:    "user",
			Content: "Tell me about quantum computing",
		},
		{
			Role:    "assistant",
			Content: "Quantum computing is a revolutionary technology...",
		},
	}

	req := request.Completion{
		Model:         models.Grok3Mini{},
		SystemMessage: systemInst,
		UserMessage:   userMsg,
		History:       history,
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	res, err := grokProvider.CompleteResponse(
		context.Background(),
		req,
		client,
		nil,
	)
	require.NoError(
		t,
		err,
		"CompleteResponse returned an unexpected error with history",
		"error",
		err,
	)

	assert.NotEmpty(t, res.Content, "Expected non-empty content with history")
	assert.Contains(
		t,
		res.Content,
		"quantum",
		"Expected response to reference previous context",
	)
}

func TestGrokErrorHandling(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	// Test with invalid API key
	grokProvider := providers.NewGrok([]string{"invalid-key"})

	req := request.Completion{
		Model:         models.Grok3{},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Hello",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	_, err := grokProvider.CompleteResponse(
		context.Background(),
		req,
		client,
		nil,
	)
	require.Error(t, err, "Expected error with invalid API key")
}
