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

	apiKey := os.Getenv("GROK_API_KEY")
	if apiKey == "" {
		t.Skip("GROK_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	grokProvider := providers.NewGrok([]string{apiKey})

	req := request.Completion{
		Model:         models.Grok3Mini{},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
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
	require.NoError(t, err, "CompleteResponse returned an unexpected error")
	assert.NotEmpty(t, res.Content, "Expected non-empty content")
	assert.Equal(t, req.Model.GetName(), res.Model, "Model mismatch")
}

func TestGrokModelsWithStreaming(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("GROK_API_KEY")
	if apiKey == "" {
		t.Skip("GROK_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	grokProvider := providers.NewGrok([]string{apiKey})

	req := request.Completion{
		Model:         models.Grok3Mini{},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	var chunks []string
	res, err := grokProvider.StreamResponse(
		context.Background(),
		client,
		req,
		func(chunk string) error {
			chunks = append(chunks, chunk)
			return nil
		},
		nil,
	)
	require.NoError(t, err, "StreamResponse returned an unexpected error")
	assert.NotEmpty(t, res.Content, "Expected non-empty content")
	assert.NotEmpty(t, chunks, "Expected streaming chunks")
}

func TestGrokErrorHandling(t *testing.T) {
	t.Parallel()

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	grokProvider := providers.NewGrok([]string{"invalid-key"})

	req := request.Completion{
		Model:         models.Grok3Mini{},
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
