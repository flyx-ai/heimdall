package heimdall

import (
	"net/http"
	"time"

	"github.com/flyx-ai/heimdall/providers"
)

type RouterConfig struct {
	Providers []providers.LLMProvider
	Timeout   time.Duration
}

type Router struct {
	providers map[string]providers.LLMProvider
	client    http.Client
}

func New(timeout time.Duration, llmProviders []providers.LLMProvider) *Router {
	c := http.Client{
		Timeout: timeout,
	}

	providers := make(map[string]providers.LLMProvider, len(llmProviders))
	for _, provider := range llmProviders {
		providers[provider.Name()] = provider
	}

	return &Router{
		providers,
		c,
	}
}
