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

// TODO: test with tools as well
func TestGoogleModelsWithCompletion(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	google := providers.NewGoogle([]string{os.Getenv("GOOGLE_API_KEY")})

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
			name: "should complete request with gemini 1.5 flash",
			req: request.Completion{
				Model:       models.Gemini15Flash{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gemini 1.5 pro",
			req: request.Completion{
				Model:       models.Gemini15Pro{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gemini 2.0 flash",
			req: request.Completion{
				Model:       models.Gemini20Flash{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gemini 2.0 flash lite",
			req: request.Completion{
				Model:       models.Gemini20FlashLite{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gemini 2.5 pro experimental",
			req: request.Completion{
				Model:       models.Gemini25ProExp{},
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
			res, err := google.CompleteResponse(
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

func TestGoogleModelsWithStreaming(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	google := providers.NewGoogle([]string{os.Getenv("GOOGLE_API_KEY")})

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
			name: "should stream request with gemini 1.5 flash",
			req: request.Completion{
				Model:       models.Gemini15Flash{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream  request with gemini 1.5 pro",
			req: request.Completion{
				Model:       models.Gemini15Pro{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with gemini 2.0 flash",
			req: request.Completion{
				Model:       models.Gemini20Flash{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with gemini 2.0 flash lite",
			req: request.Completion{
				Model:       models.Gemini20FlashLite{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		// {
		// 	name: "should complete request with gemini 2.5 pro experimental",
		// 	req: request.Completion{
		// 		Model:       models.Gemini25ProExp{},
		// 		Messages:    msgs,
		// 		Temperature: 1,
		// 		Tags: map[string]string{
		// 			"type": "testing",
		// 		},
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var chunkHandlerCollection string
			res, err := google.StreamResponse(
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
				"StreamResponse returned an unexpected error",
				"error",
				err,
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
