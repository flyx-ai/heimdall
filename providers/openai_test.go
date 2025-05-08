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

func TestOpenAIModelsWithCompletion(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	openai := providers.NewOpenAI([]string{os.Getenv("OPENAI_API_KEY")})

	systemInst := "you are a helpful assistant."
	userMsg := "please make a detailed analysis of the NVIDIA's current valuation."

	tests := []struct {
		name string
		req  request.Completion
	}{
		{
			name: "should complete request with GPT4",
			req: request.Completion{
				Model:         models.GPT4{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with GPT4Turbo",
			req: request.Completion{
				Model:         models.GPT4Turbo{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with GPT4O",
			req: request.Completion{
				Model:         models.GPT4O{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with GPT4OMini",
			req: request.Completion{
				Model:         models.GPT4OMini{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with O1",
			req: request.Completion{
				Model:         models.O1{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with O1Mini",
			req: request.Completion{
				Model:         models.O1Mini{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with O1Preview",
			req: request.Completion{
				Model:         models.O1Preview{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gpt 4.1",
			req: request.Completion{
				Model:         models.GPT41{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gpt 4.1 mini",
			req: request.Completion{
				Model:         models.GPT41Mini{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gpt 4.1 nano",
			req: request.Completion{
				Model:         models.GPT41Nano{},
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
			res, err := openai.CompleteResponse(
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

func TestOpenAIModelsWithStreaming(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	openai := providers.NewOpenAI([]string{os.Getenv("OPENAI_API_KEY")})

	systemInst := "you are a helpful assistant."
	userMsg := "please make a detailed analysis of the NVIDIA's current valuation."

	tests := []struct {
		name string
		req  request.Completion
	}{
		{
			name: "should stream request with GPT4",
			req: request.Completion{
				Model:         models.GPT4{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with GPT4Turbo",
			req: request.Completion{
				Model:         models.GPT4Turbo{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with GPT4O",
			req: request.Completion{
				Model:         models.GPT4O{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with GPT4OMini",
			req: request.Completion{
				Model:         models.GPT4OMini{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with O1",
			req: request.Completion{
				Model:         models.O1{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with O1Mini",
			req: request.Completion{
				Model:         models.O1Mini{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with O1Preview",
			req: request.Completion{
				Model:         models.O1Preview{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should complete request with gpt 4.1",
			req: request.Completion{
				Model:         models.GPT41{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with gpt 4.1 mini",
			req: request.Completion{
				Model:         models.GPT41Mini{},
				SystemMessage: systemInst,
				UserMessage:   userMsg,
				Temperature:   1,
				Tags: map[string]string{
					"type": "testing",
				},
			},
		},
		{
			name: "should stream request with gpt 4.1 nano",
			req: request.Completion{
				Model:         models.GPT41Nano{},
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
			res, err := openai.StreamResponse(
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

// TestOpenAIStructuredOutput tests the structured output functionality
// func TestOpenAIStructuredOutput(t *testing.T) {
// 	t.Parallel()
//
// 	client := http.Client{
// 		Timeout: 2 * time.Minute,
// 	}
// 	openai := providers.NewOpenAI([]string{os.Getenv("OPENAI_API_KEY")})
//
// 	// Define a test schema for structured output
// 	schema := map[string]any{
// 		"type": "object",
// 		"properties": map[string]any{
// 			"sentiment": map[string]any{
// 				"type": "string",
// 				"enum": []string{"positive", "negative", "neutral"},
// 			},
// 			"summary": map[string]any{
// 				"type": "string",
// 			},
// 			"key_points": map[string]any{
// 				"type": "array",
// 				"items": map[string]any{
// 					"type": "string",
// 				},
// 			},
// 		},
// 		"required": []string{"sentiment", "summary", "key_points"},
// 	}
//
// 	// Create requests for models that support structured output
// 	tests := []struct {
// 		name  string
// 		model models.Model
// 	}{
// 		// {
// 		// 	name:  "GPT-4o with structured output",
// 		// 	model: &models.GPT4O{StructuredOutput: schema},
// 		// },
// 		{
// 			name:  "GPT-4.1 with structured output",
// 			model: &models.GPT41{StructuredOutput: schema},
// 		},
// 		// {
// 		// 	name:  "O1 with structured output",
// 		// 	model: &models.O1{StructuredOutput: schema},
// 		// },
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := request.Completion{
// 				Model:         tt.model,
// 				SystemMessage: "You are a helpful assistant that analyzes text.",
// 				UserMessage:   "Analyze the sentiment of: 'I love this product, it's amazing!'",
// 				Temperature:   0.0,
// 				Tags: map[string]string{
// 					"type": "testing",
// 				},
// 			}
//
// 			res, err := openai.CompleteResponse(
// 				context.Background(),
// 				req,
// 				client,
// 				nil,
// 			)
// 			require.NoError(
// 				t,
// 				err,
// 				"CompleteResponse returned an unexpected error",
// 			)
//
// 			assert.NotEmpty(t, res.Content, "content should not be empty")
// 			assert.Contains(
// 				t,
// 				res.Content,
// 				"sentiment",
// 				"response should contain the sentiment field",
// 			)
// 			assert.Contains(
// 				t,
// 				res.Content,
// 				"summary",
// 				"response should contain the summary field",
// 			)
// 			assert.Contains(
// 				t,
// 				res.Content,
// 				"key_points",
// 				"response should contain the key_points field",
// 			)
// 		})
// 	}
// }

// TestOpenAIImageGeneration tests the image generation functionality
func TestOpenAIImageGeneration(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	openai := providers.NewOpenAI([]string{os.Getenv("OPENAI_API_KEY")})

	// Test with different image generation configurations
	tests := []struct {
		name        string
		imageModel  *models.GPTImage
		userMessage string
	}{
		{
			name: "basic image generation",
			imageModel: &models.GPTImage{
				Size: models.GPTImageSize1024x1024,
			},
			userMessage: "A serene landscape with mountains and a lake at sunset.",
		},
		{
			name: "image with high quality",
			imageModel: &models.GPTImage{
				Size:    models.GPTImageSize1024x1024,
				Quality: models.GPTImageQualityHigh,
			},
			userMessage: "A detailed close-up of a butterfly on a flower.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := request.Completion{
				Model:       tt.imageModel,
				UserMessage: tt.userMessage,
				Tags: map[string]string{
					"type": "testing",
				},
			}

			res, err := openai.CompleteResponse(
				context.Background(),
				req,
				client,
				nil,
			)
			require.NoError(
				t,
				err,
				"Image generation returned an unexpected error",
			)

			// The response should contain a base64-encoded image
			assert.NotEmpty(t, res.Content, "image content should not be empty")
			assert.True(
				t,
				len(res.Content) > 1000,
				"image content should be substantial",
			)
		})
	}
}

// TestOpenAIPDFHandling tests the PDF file handling functionality
func TestOpenAIPDFHandling(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	openai := providers.NewOpenAI([]string{os.Getenv("OPENAI_API_KEY")})

	// In a real test, you would need a base64-encoded PDF file
	// This is a placeholder - you'd need to replace with actual base64 content
	pdfData := map[string]string{
		"test.pdf": "data:application/pdf;base64,JVBERi0xLjEKJcKlwrHDqwoKMSAwIG9iagogIDw8IC9UeXBlIC9DYXRhbG9nCiAgICAgL1BhZ2VzIDIgMCBSCiAgPj4KZW5kb2JqCgoyIDAgb2JqCiAgPDwgL1R5cGUgL1BhZ2VzCiAgICAgL0tpZHMgWzMgMCBSXQogICAgIC9Db3VudCAxCiAgICAgL01lZGlhQm94IFswIDAgMzAwIDE0NF0KICA+PgplbmRvYmoKCjMgMCBvYmoKICA8PCAgL1R5cGUgL1BhZ2UKICAgICAgL1BhcmVudCAyIDAgUgogICAgICAvUmVzb3VyY2VzCiAgICAgICA8PCAvRm9udAogICAgICAgICAgIDw8IC9GMQogICAgICAgICAgICAgICA8PCAvVHlwZSAvRm9udAogICAgICAgICAgICAgICAgICAvU3VidHlwZSAvVHlwZTEKICAgICAgICAgICAgICAgICAgL0Jhc2VGb250IC9UaW1lcy1Sb21hbgogICAgICAgICAgICAgICA+PgogICAgICAgICAgID4+CiAgICAgICA+PgogICAgICAvQ29udGVudHMgNCAwIFIKICA+PgplbmRvYmoKCjQgMCBvYmoKICA8PCAvTGVuZ3RoIDU1ID4+CnN0cmVhbQogIEJUCiAgICAvRjEgMTggVGYKICAgIDAgMCBUZAogICAgKEhlbGxvIFdvcmxkKSBUagogIEVUCmVuZHN0cmVhbQplbmRvYmoKCnhyZWYKMCA1CjAwMDAwMDAwMDAgNjU1MzUgZiAKMDAwMDAwMDAxOCAwMDAwMCBuIAowMDAwMDAwMDc3IDAwMDAwIG4gCjAwMDAwMDAxNzggMDAwMDAgbiAKMDAwMDAwMDQ1NyAwMDAwMCBuIAp0cmFpbGVyCiAgPDwgIC9Sb290IDEgMCBSCiAgICAgIC9TaXplIDUKICA+PgpzdGFydHhyZWYKNTY1CiUlRU9GCg==", // truncated for brevity
	}

	// Test with models that support PDF handling
	tests := []struct {
		name  string
		model models.Model
	}{
		{
			name:  "GPT-4o with PDF",
			model: &models.GPT4O{PdfFile: pdfData},
		},
		{
			name:  "GPT-4.1 with PDF",
			model: &models.GPT41{PdfFile: pdfData},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := request.Completion{
				Model:         tt.model,
				SystemMessage: "You are a helpful assistant.",
				UserMessage:   "Summarize the content of this PDF document.",
				Temperature:   0.7,
				Tags: map[string]string{
					"type": "testing",
				},
			}

			res, err := openai.CompleteResponse(
				context.Background(),
				req,
				client,
				nil,
			)
			require.NoError(t, err, "PDF handling returned an unexpected error")
			assert.NotEmpty(
				t,
				res.Content,
				"response content should not be empty",
			)
		})
	}
}
