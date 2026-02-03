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

func TestGoogleModelsWithCompletion(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	google := providers.NewGoogle([]string{apiKey})

	req := request.Completion{
		Model:         models.Gemini20FlashLite{},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	res, err := google.CompleteResponse(
		context.Background(),
		req,
		client,
		nil,
	)
	require.NoError(t, err, "CompleteResponse returned an unexpected error")
	assert.NotEmpty(t, res.Content, "content should not be empty")
	assert.NotEmpty(t, res.Model, "model should not be empty")
}

func TestGoogleModelsWithStreaming(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	google := providers.NewGoogle([]string{apiKey})

	req := request.Completion{
		Model:         models.Gemini20FlashLite{},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	var chunkHandlerCollection string
	res, err := google.StreamResponse(
		context.Background(),
		client,
		req,
		func(chunk string) error {
			chunkHandlerCollection = chunkHandlerCollection + chunk
			return nil
		},
		nil,
	)
	require.NoError(t, err, "StreamResponse returned an unexpected error")
	assert.NotEmpty(t, chunkHandlerCollection, "chunkHandlerCollection should not be empty")
	assert.NotEmpty(t, res.Content, "content should not be empty")
	assert.NotEmpty(t, res.Model, "model should not be empty")
}
