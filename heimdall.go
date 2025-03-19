package heimdall

import (
	"context"
	"net/http"
	"time"
)

type RouterConfig struct {
	Providers []LLMProvider
	Timeout   time.Duration
}

type LLMProvider interface {
	completeResponse(
		ctx context.Context,
		req CompletionRequest,
		client http.Client,
		requestLog *Logging,
	) (CompletionResponse, error)
	streamResponse(
		ctx context.Context,
		client http.Client,
		req CompletionRequest,
		chunkHandler func(chunk string) error,
		requestLog *Logging,
	) (CompletionResponse, error)
	tryWithBackup(
		ctx context.Context,
		req CompletionRequest,
		client http.Client,
		chunkHandler func(chunk string) error,
		requestLog *Logging,
	) (CompletionResponse, error)
	doRequest(
		ctx context.Context,
		req CompletionRequest,
		client http.Client,
		chunkHandler func(chunk string) error,
		key string,
	) (CompletionResponse, int, error)
	getApiKeys() []string
	name() string
}

type Router struct {
	providers map[string]LLMProvider
	client    http.Client
}

func New(timeout time.Duration, llmProviders []LLMProvider) *Router {
	c := http.Client{
		Timeout: timeout,
	}

	providers := make(map[string]LLMProvider, len(llmProviders))
	for _, provider := range llmProviders {
		providers[provider.name()] = provider
	}

	return &Router{
		providers,
		c,
	}
}
