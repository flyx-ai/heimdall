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

func TestAnthropicModelsWithCompletion(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	anthropicProvider := providers.NewAnthropic([]string{apiKey})

	req := request.Completion{
		Model:         models.Claude35Haiku{},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	res, err := anthropicProvider.CompleteResponse(
		context.Background(),
		req,
		client,
		nil,
	)
	require.NoError(t, err, "CompleteResponse returned an unexpected error")
	assert.NotEmpty(t, res.Content, "content should not be empty")
	assert.NotEmpty(t, res.Model, "model should not be empty")
}

func TestClaude46OpusWithCompletion(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	anthropicProvider := providers.NewAnthropic([]string{apiKey})

	req := request.Completion{
		Model: models.Claude46Opus{
			MaxOutputTokens: 8192,
		},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	res, err := anthropicProvider.CompleteResponse(
		context.Background(),
		req,
		client,
		nil,
	)
	require.NoError(t, err, "CompleteResponse returned an unexpected error")
	assert.NotEmpty(t, res.Content, "content should not be empty")
	assert.NotEmpty(t, res.Model, "model should not be empty")
}

func TestClaude46OpusWithExtendedContext(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	anthropicProvider := providers.NewAnthropic([]string{apiKey})

	req := request.Completion{
		Model: models.Claude46Opus{
			ExtendedContext: true,
		},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	res, err := anthropicProvider.CompleteResponse(
		context.Background(),
		req,
		client,
		nil,
	)
	require.NoError(t, err, "CompleteResponse returned an unexpected error")
	assert.NotEmpty(t, res.Content, "content should not be empty")
	assert.NotEmpty(t, res.Model, "model should not be empty")
}

func TestAnthropicModelsWithStreaming(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	client := http.Client{
		Timeout: 2 * time.Minute,
	}
	anthropicProvider := providers.NewAnthropic([]string{apiKey})

	req := request.Completion{
		Model:         models.Claude35Haiku{},
		SystemMessage: "you are a helpful assistant.",
		UserMessage:   "Say hello in one sentence.",
		Temperature:   1,
		Tags: map[string]string{
			"type": "testing",
		},
	}

	var chunkHandlerCollection string
	res, err := anthropicProvider.StreamResponse(
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
