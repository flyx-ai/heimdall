package providers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

// MockRoundTripper for intercepting HTTP requests
type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
	RequestBody   []byte
}

func (m *MockRoundTripper) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	m.RequestBody = bodyBytes // Store the request body
	req.Body = io.NopCloser(
		bytes.NewBuffer(bodyBytes),
	) // Restore the body for the actual request if needed

	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	// Default mock response (can be customized)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(
			bytes.NewBufferString(
				"data: {\"choices\": [{\"delta\": {\"content\": \"mock response\"}}], \"usage\": {\"prompt_tokens\": 10, \"completion_tokens\": 5, \"total_tokens\": 15}}\n\ndata: [DONE]\n",
			),
		),
		Header: make(http.Header),
	}, nil
}

func TestOpenAIWebSearchToolParameter(t *testing.T) {
	mockRT := &MockRoundTripper{}
	client := http.Client{Transport: mockRT}
	openai := providers.NewOpenAI([]string{"fake-key"})

	baseUserMsg := "What's the weather like?"

	t.Run("WebSearchEnabled_GPT4O", func(t *testing.T) {
		reqWithWebSearch := request.Completion{
			Model: models.GPT4O{
				EnableWebSearch: true,
			},
			UserMessage: baseUserMsg,
			Tags:        make(map[string]string), // Initialize Tags map
		}

		_, err := openai.CompleteResponse(
			context.Background(),
			reqWithWebSearch,
			client,
			nil,
		)
		require.NoError(t, err)

		// Check the captured request body
		var requestPayload map[string]any
		err = json.Unmarshal(mockRT.RequestBody, &requestPayload)
		require.NoError(t, err, "Failed to unmarshal request body")

		// Assert tools array is present and correct
		tools, ok := requestPayload["tools"].([]any)
		assert.True(t, ok, "tools field should be present")
		require.Len(t, tools, 1, "tools array should have one element")
		tool, ok := tools[0].(map[string]any)
		assert.True(t, ok, "tool should be a map")
		assert.Equal(
			t,
			models.OpenAIToolTypeWebSearch,
			tool["type"],
			"tool type should be web_search",
		)

		// Assert tool_choice is present and correct
		toolChoice, ok := requestPayload["tool_choice"].(string)
		assert.True(t, ok, "tool_choice field should be present and a string")
		assert.Equal(t, "auto", toolChoice, "tool_choice should be 'auto'")
	})

	t.Run("WebSearchDisabled_GPT4O", func(t *testing.T) {
		reqWithoutWebSearch := request.Completion{
			Model: models.GPT4O{ // EnableWebSearch defaults to false
			},
			UserMessage: baseUserMsg,
			Tags:        make(map[string]string), // Initialize Tags map
		}

		_, err := openai.CompleteResponse(
			context.Background(),
			reqWithoutWebSearch,
			client,
			nil,
		)
		require.NoError(t, err)

		// Check the captured request body
		var requestPayload map[string]any
		err = json.Unmarshal(mockRT.RequestBody, &requestPayload)
		require.NoError(t, err, "Failed to unmarshal request body")

		// Assert tools and tool_choice are absent
		_, toolsExist := requestPayload["tools"]
		assert.False(t, toolsExist, "tools field should NOT be present")
		_, toolChoiceExist := requestPayload["tool_choice"]
		assert.False(
			t,
			toolChoiceExist,
			"tool_choice field should NOT be present",
		)
	})

	// Add similar tests for GPT4OMini, GPT41, O1 if desired

	// Test a model without the flag
	t.Run("WebSearchUnsupportedModel_GPT4", func(t *testing.T) {
		reqUnsupported := request.Completion{
			Model:       models.GPT4{}, // Does not have EnableWebSearch field
			UserMessage: baseUserMsg,
			Tags:        make(map[string]string), // Initialize Tags map
		}

		_, err := openai.CompleteResponse(
			context.Background(),
			reqUnsupported,
			client,
			nil,
		)
		// Skip error check due to known panic issue
		require.NoError(t, err)

		// Check the captured request body
		var requestPayload map[string]any
		err = json.Unmarshal(mockRT.RequestBody, &requestPayload)
		require.NoError(t, err, "Failed to unmarshal request body")

		// Assert tools and tool_choice are absent
		_, toolsExist := requestPayload["tools"]
		assert.False(
			t,
			toolsExist,
			"tools field should NOT be present on unsupported model",
		)
		_, toolChoiceExist := requestPayload["tool_choice"]
		assert.False(
			t,
			toolChoiceExist,
			"tool_choice field should NOT be present on unsupported model",
		)
	})
}

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
