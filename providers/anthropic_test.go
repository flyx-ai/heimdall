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
func TestAnthropicModelsWithCompletion(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	anthropicProvider := providers.NewAnthropicClient(
		[]string{os.Getenv("ANTHROPIC_API_KEY")},
	)

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
			name: "should complete request with claude-3-haiku",
			req: request.Completion{
				Model:       models.Claude3Haiku{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with claude-3-sonnet",
			req: request.Completion{
				Model:       models.Claude3Sonnet{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with claude-3-opus",
			req: request.Completion{
				Model:       models.Claude3Opus{},
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
		})
	}
}
