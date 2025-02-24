package heimdall

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type APIKey struct {
	Key           string
	RequestsLimit int
	requestsUsed  int
	ResetAt       time.Time
	mu            sync.Mutex
	Name             string
	Key              string
	RequestsLimit    int
	requestsUsed     int
	RequestRemaining int
	ResetAt          time.Time
}

type RouterConfig struct {
	ProviderAPIKeys map[Provider][]APIKey
	Timeout         time.Duration
}

type LLM interface {
	StreamResponse(
		ctx context.Context,
		req CompletionRequest,
		key string,
		chunkHandler func(chunk string) error,
	) (*CompletionResponse, error)
}

type Router struct {
	config RouterConfig
	llms   map[Provider]LLM
}

func (r *Router) ReqsStats() {
	for provider, keys := range r.config.ProviderAPIKeys {
		for _, key := range keys {
			slog.Info(
				"API KEY STATS",
				"provider",
				provider,
				"key_name",
				key.Name,
				"requests_remaining",
				key.RequestRemaining,
				"requests_used",
				key.requestsUsed,
			)
		}
	}
}
func New(config RouterConfig) *Router {
	c := http.Client{
		Timeout: config.Timeout,
	}

	openai := Openai{Client: c}

	return &Router{
		config: config,
		llms: map[Provider]LLM{
			ProviderOpenAI: openai,
		},
	}
}

func (r *Router) Stream(
	ctx context.Context,
	req CompletionRequest,
	chunkHandler func(chunk string) error,
) error {
	if chunkHandler == nil {
		return ErrNoChunkHandler
	}

	models := append([]Model{req.Model}, req.Fallback...)
	var resp *CompletionResponse
	var err error

	for _, model := range models {
		resp, err = r.tryStreamWithModel(
			ctx,
			req,
			model,
			chunkHandler,
		)
		if resp != nil && err == nil {
			break
		}
	}

	req.Tags["request_type"] = "stream"

	return err
}

func (r *Router) tryStreamWithModel(
	ctx context.Context,
	req CompletionRequest,
	model Model,
	chunkHandler func(chunk string) error,
) (*CompletionResponse, error) {
	keys := r.config.ProviderAPIKeys[model.Provider]
	var err error

	for i := range keys {
		key := keys[i]
		if !key.isAvailable() {
			continue
		}

		resp, err := r.streamResponse(
			ctx,
			req,
			model.Provider,
			key.Key,
			chunkHandler,
		)
		if err != nil {
			if errors.Is(err, ErrRateLimitHit) {
				key.handleRateLimit()
				continue
			}
			if errors.Is(err, context.Canceled) {
				continue
			}

			continue
		}

		return resp, nil
	}

	return nil, err
}

func (r *Router) streamResponse(
	ctx context.Context,
	req CompletionRequest,
	provider Provider,
	key string,
	chunkHandler func(chunk string) error,
) (*CompletionResponse, error) {
	var resp *CompletionResponse
	var err error

	switch provider {
	case ProviderOpenAI:
		resp, err = r.llms[ProviderOpenAI].StreamResponse(
			ctx,
			req,
			key,
			chunkHandler,
		)
	default:
		err = ErrUnsupportedProvider
	}

	return resp, err
}

func (ak *APIKey) isAvailable() bool {
	ak.mu.Lock()
	defer ak.mu.Unlock()

	if time.Now().After(ak.ResetAt) {
		ak.requestsUsed = 0
		ak.ResetAt = time.Now().Add(24 * time.Hour)
	}

	return ak.requestsUsed < ak.RequestsLimit
}

func (ak *APIKey) handleRateLimit() {
	ak.mu.Lock()
	defer ak.mu.Unlock()

	ak.requestsUsed = ak.RequestsLimit
}
