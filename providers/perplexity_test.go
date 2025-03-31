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

// func TestPerplexityModelsWithCompletion(t *testing.T) {
// 	t.Parallel()
//
// 	client := http.Client{
// 		Timeout: 2 * time.Minute,
// 	}
// 	perplexity := providers.NewPerplexity(
// 		[]string{os.Getenv("PERPLEXITY_API_KEY")},
// 	)
//
// 	msgs := []request.Message{
// 		{
// 			Role:    "system",
// 			Content: "you are a helpful assistant.",
// 		},
// 		{
// 			Role:    "user",
// 			Content: "please make a detailed analysis of the NVIDIA's current valuation.",
// 		},
// 	}
//
// 	tests := []struct {
// 		name string
// 		req  request.Completion
// 	}{
// 		{
// 			name: "should complete request with Sonar",
// 			req: request.Completion{
// 				Model:       models.Sonar{},
// 				Messages:    msgs,
// 				Temperature: 1,
// 				Tags: map[string]string{
// 					"type": "testing",
// 				},
// 			},
// 		},
// 		{
// 			name: "should complete request with SonarPro",
// 			req: request.Completion{
// 				Model:       models.SonarPro{},
// 				Messages:    msgs,
// 				Temperature: 1,
// 				Tags: map[string]string{
// 					"type": "testing",
// 				},
// 			},
// 		},
// 		{
// 			name: "should complete request with SonarReasoning",
// 			req: request.Completion{
// 				Model:       models.SonarReasoning{},
// 				Messages:    msgs,
// 				Temperature: 1,
// 				Tags: map[string]string{
// 					"type": "testing",
// 				},
// 			},
// 		},
// 		{
// 			name: "should complete request with SonarReasoningPro",
// 			req: request.Completion{
// 				Model:       models.SonarReasoningPro{},
// 				Messages:    msgs,
// 				Temperature: 1,
// 				Tags: map[string]string{
// 					"type": "testing",
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			res, err := perplexity.CompleteResponse(
// 				context.Background(),
// 				tt.req,
// 				client,
// 				nil,
// 			)
// 			require.NoError(
// 				t,
// 				err,
// 				"CompleteResponse returned an unexpected error",
// 				"error",
// 				err,
// 			)
//
// 			assert.NotEmpty(t, res.Content, "content should not be empty")
// 		})
// 	}
// }

func TestPerplexityModelsWithStreaming(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	perplexity := providers.NewPerplexity(
		[]string{os.Getenv("PERPLEXITY_API_KEY")},
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
			name: "should stream request with Sonar",
			req: request.Completion{
				Model:       models.Sonar{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with SonarPro",
			req: request.Completion{
				Model:       models.SonarPro{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with SonarReasoning",
			req: request.Completion{
				Model:       models.SonarReasoning{},
				Messages:    msgs,
				Temperature: 1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with SonarReasoningPro",
			req: request.Completion{
				Model:       models.SonarReasoningPro{},
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
			var chunkHandlerCollection string
			res, err := perplexity.StreamResponse(
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
