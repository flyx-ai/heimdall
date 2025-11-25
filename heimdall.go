package heimdall

import (
	"context"
	"net/http"
	"time"

	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

type LLMProvider interface {
	CompleteResponse(
		ctx context.Context,
		req request.Completion,
		client http.Client,
		requestLog *response.Logging,
	) (response.Completion, error)
	StreamResponse(
		ctx context.Context,
		client http.Client,
		req request.Completion,
		chunkHandler func(chunk string) error,
		requestLog *response.Logging,
	) (response.Completion, error)
	Name() string
}

type RouterConfig struct {
	Providers []LLMProvider
	Timeout   time.Duration
}

type Router struct {
	providers map[string]LLMProvider
	client    http.Client
}

func New(timeout time.Duration, llmProviders []LLMProvider) *Router {
	c := http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout: timeout,
			IdleConnTimeout:       timeout,
		},
	}

	providers := make(map[string]LLMProvider, len(llmProviders))
	for _, provider := range llmProviders {
		providers[provider.Name()] = provider
	}

	return &Router{
		providers,
		c,
	}
}
