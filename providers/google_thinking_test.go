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

func TestGemini25WithThinking(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set, skipping test")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	google := providers.NewGoogle([]string{apiKey})

	tests := []struct {
		name           string
		model          models.Model
		expectThoughts bool
		userMsg        string
	}{
		{
			name: "Gemini 2.5 Flash with high thinking budget",
			model: models.Gemini25FlashPreview{
				Thinking: models.HighThinkBudget,
			},
			expectThoughts: true,
			userMsg:        "Solve this step by step: If a train travels 120 miles in 2 hours, and then 180 miles in 3 hours, what is its average speed for the entire journey?",
		},
		{
			name: "Gemini 2.5 Pro with medium thinking budget",
			model: models.Gemini25ProPreview{
				Thinking: models.MediumThinkBudget,
			},
			expectThoughts: true,
			userMsg:        "Analyze the pros and cons of using microservices architecture versus monolithic architecture for a startup.",
		},
		{
			name: "Gemini 2.5 Flash with low thinking budget",
			model: models.Gemini25FlashPreview{
				Thinking: models.LowThinkBudget,
			},
			expectThoughts: false,
			userMsg:        "What is 2 + 2?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := request.Completion{
				Model:         tt.model,
				SystemMessage: "You are a helpful assistant that explains your reasoning process.",
				UserMessage:   tt.userMsg,
				Temperature:   0.7,
				Tags: map[string]string{
					"test": "thinking",
				},
			}

			res, err := google.CompleteResponse(
				context.Background(),
				req,
				client,
				nil,
			)
			require.NoError(t, err, "CompleteResponse returned an error")

			assert.NotEmpty(t, res.Content, "Content should not be empty")
			assert.NotEmpty(t, res.Model, "Model should not be empty")

			if tt.expectThoughts {
				assert.NotEmpty(t, res.Thoughts, "Thoughts should not be empty when thinking budget is set")
				t.Logf("Thoughts: %s", res.Thoughts)
			} else {
				assert.Empty(t, res.Thoughts, "Thoughts should be empty when thinking budget is 0")
			}

			t.Logf("Response: %s", res.Content)
		})
	}
}

func TestGemini25StreamingWithThinking(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set, skipping test")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	google := providers.NewGoogle([]string{apiKey})

	req := request.Completion{
		Model: models.Gemini25FlashPreview{
			Thinking: models.HighThinkBudget,
		},
		SystemMessage: "You are a helpful math tutor.",
		UserMessage:   "Explain how to solve quadratic equations using the quadratic formula, with an example.",
		Temperature:   0.7,
		Tags: map[string]string{
			"test": "streaming-thinking",
		},
	}

	var streamedContent string
	res, err := google.StreamResponse(
		context.Background(),
		client,
		req,
		func(chunk string) error {
			streamedContent += chunk
			return nil
		},
		nil,
	)

	require.NoError(t, err, "StreamResponse returned an error")
	assert.NotEmpty(t, res.Content, "Content should not be empty")
	assert.NotEmpty(t, res.Thoughts, "Thoughts should not be empty with thinking budget")
	assert.Equal(t, streamedContent, res.Content, "Streamed content should match final content")

	t.Logf("Thoughts: %s", res.Thoughts)
	t.Logf("Content: %s", res.Content)
}
