package providers_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"encoding/json"
	"io"
	"net/http/httptest"

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

func TestOpenAIProvider_ImageGeneration_Success(t *testing.T) {
	t.Parallel()

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request path and method
		assert.Equal(t, "/v1/images/generations", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Check auth header
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		// Check request body
		bodyBytes, err := io.ReadAll(r.Body)
		assert.NoError(t, err)        // Use assert in handler
		var reqPayload map[string]any // Use any instead of interface{}
		err = json.Unmarshal(bodyBytes, &reqPayload)
		assert.NoError(t, err) // Use assert in handler

		assert.Equal(t, models.Dalle3Alias, reqPayload["model"])
		assert.Equal(t, "A painting of a cat sitting on a table.", reqPayload["prompt"])
		assert.InDelta(t, float64(1), reqPayload["n"], 0.001) // Use InDelta for float comparison
		assert.Equal(t, models.Dalle3Size1024x1024, reqPayload["size"])
		assert.Equal(t, "url", reqPayload["response_format"])

		// Send mock response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"created": 1677649963,
			"data": [
				{
					"url": "https://example.com/image1.png",
					"revised_prompt": "A detailed painting of a fluffy cat lounging on a wooden table."
				}
			]
		}`))
	}))
	defer server.Close()

	// // Patch the base URL to use the mock server
	// // NOTE: This requires modifying the provider to allow URL injection/patching.
	// originalBaseURL := getOpenAIBaseURLForTest() // Using helper
	// setOpenAIBaseURLForTest(server.URL)          // Using helper
	// defer setOpenAIBaseURLForTest(originalBaseURL) // Restore original URL

	// Setup provider and client
	client := http.Client{Timeout: 5 * time.Second}
	openai := providers.NewOpenAI([]string{"test-key"})

	// Create request
	req := request.Completion{
		Model:       &models.Dalle3{}, // Use pointer for type assertion
		UserMessage: "A painting of a cat sitting on a table.",
		Tags:        map[string]string{"type": "testing"},
	}

	// Call CompleteResponse (image generation is non-streaming)
	res, err := openai.CompleteResponse(context.Background(), req, client, nil)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/image1.png", res.Content)
	assert.Equal(t, models.Dalle3Alias, res.Model)
	// assert.Equal(t, 0, res.Usage.TotalTokens) // Usage is expected to be 0 for images
}

func TestOpenAIProvider_ImageGeneration_Params(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/images/generations", r.URL.Path)
		bodyBytes, err := io.ReadAll(r.Body)
		assert.NoError(t, err)        // Use assert in handler
		var reqPayload map[string]any // Use any instead of interface{}
		err = json.Unmarshal(bodyBytes, &reqPayload)
		assert.NoError(t, err) // Use assert in handler

		assert.Equal(t, models.Dalle3Alias, reqPayload["model"])
		assert.Equal(t, "A futuristic cityscape.", reqPayload["prompt"])
		assert.Equal(t, models.Dalle3Size1792x1024, reqPayload["size"])
		assert.Equal(t, models.Dalle3QualityHD, reqPayload["quality"])
		assert.Equal(t, models.Dalle3StyleNatural, reqPayload["style"])
		assert.Equal(t, "user-1234", reqPayload["user"])

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"created": 1677649964, "data": [{"url": "https://example.com/image_params.png"}]}`))
	}))
	defer server.Close()

	// // Patch the base URL to use the mock server
	// originalBaseURL := getOpenAIBaseURLForTest()
	// setOpenAIBaseURLForTest(server.URL)
	// defer setOpenAIBaseURLForTest(originalBaseURL)

	client := http.Client{Timeout: 5 * time.Second}
	openai := providers.NewOpenAI([]string{"test-key"})

	req := request.Completion{
		Model: &models.Dalle3{
			Size:    models.Dalle3Size1792x1024,
			Quality: models.Dalle3QualityHD,
			Style:   models.Dalle3StyleNatural,
			User:    "user-1234",
			// N is implicitly 1
		},
		UserMessage: "A futuristic cityscape.",
	}

	res, err := openai.CompleteResponse(context.Background(), req, client, nil)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/image_params.png", res.Content)
	assert.Equal(t, models.Dalle3Alias, res.Model)
}

func TestOpenAIProvider_ImageGeneration_APIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/images/generations", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest) // Simulate a 400 Bad Request
		_, _ = w.Write([]byte(`{"error": {"message": "Invalid prompt", "type": "invalid_request_error"}}`))
	}))
	defer server.Close()

	// // Patch the base URL to use the mock server
	// originalBaseURL := getOpenAIBaseURLForTest()
	// setOpenAIBaseURLForTest(server.URL)
	// defer setOpenAIBaseURLForTest(originalBaseURL)

	client := http.Client{Timeout: 5 * time.Second}
	openai := providers.NewOpenAI([]string{"test-key"})

	req := request.Completion{
		Model:       &models.Dalle3{},
		UserMessage: "---", // Potentially invalid prompt
	}

	_, err := openai.CompleteResponse(context.Background(), req, client, nil)

	require.Error(t, err)
	// Check if the error message contains the expected status code and API error details
	assert.Contains(t, err.Error(), "status 400")
	assert.Contains(t, err.Error(), "Invalid prompt")
	assert.Contains(t, err.Error(), "invalid_request_error")
}

/*
// --- Helper functions for testing BaseURL ---
// These are needed because the original code doesn't easily allow URL injection.
// We simulate patching the URL conceptually.

var (
	testBaseURL = "https://api.openai.com/v1" // Default value
	mu sync.Mutex
)

// setOpenAIBaseURLForTest overrides the package-level base URL for testing.
// WARNING: This is not concurrency-safe without external locking if tests run in parallel
// modifying the same global variable. The parallel tests above are safe because they only read this conceptually.
// func setOpenAIBaseURLForTest(url string) {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	// This conceptually overrides the openAIBaseURL constant in the providers package
// 	// For this test file, we store it locally.
// 	testBaseURL = url
// 	// In real code, you'd need a mechanism to actually change the URL used by the provider,
// 	// e.g., reflection (unsafe) or ideally dependency injection.
// }

// // getOpenAIBaseURLForTest returns the current base URL used for testing.
// func getOpenAIBaseURLForTest() string {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	return testBaseURL
// }

// The actual http request creation in callImageGenerationAPI would need to be modified
// to use getOpenAIBaseURLForTest() instead of the hardcoded constant for these tests to work
// without real patching/DI.
*/
