package heimdall

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type APIKey struct {
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
	completeResponse(
		ctx context.Context,
		req CompletionRequest,
		key APIKey,
	) (*CompletionResponse, error)
	streamResponse(
		ctx context.Context,
		req CompletionRequest,
		key APIKey,
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

	openai := openai{client: c}
	google := google{client: c}

	return &Router{
		config: config,
		llms: map[Provider]LLM{
			ProviderOpenAI: openai,
			ProviderGoogle: google,
		},
	}
}

func (r *Router) Complete(
	ctx context.Context,
	req CompletionRequest,
) (*CompletionResponse, error) {
	models := append([]Model{req.Model}, req.Fallback...)
	var err error
	var resp *CompletionResponse

	for _, model := range models {
		resp, err = r.tryWithModel(ctx, req, model)
		if resp != nil && err == nil {
			break
		}

		continue
	}

	return resp, err
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
			key,
			chunkHandler,
		)
		if err != nil {
			slog.ErrorContext(ctx, "tryStreamWithModel", "error", err)
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
	key APIKey,
	chunkHandler func(chunk string) error,
) (*CompletionResponse, error) {
	var resp *CompletionResponse
	var err error

	switch provider {
	case ProviderOpenAI:
		resp, err = r.llms[ProviderOpenAI].streamResponse(
			ctx,
			req,
			key,
			chunkHandler,
		)
	case ProviderGoogle:
		resp, err = r.llms[ProviderGoogle].streamResponse(
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

func (r *Router) tryWithModel(
	ctx context.Context,
	req CompletionRequest,
	model Model,
) (*CompletionResponse, error) {
	keys := r.config.ProviderAPIKeys[model.Provider]
	var err error

	for i := range keys {
		key := keys[i]
		if !key.isAvailable() {
			continue
		}

		resp, err := r.completeResponse(
			ctx,
			req,
			model.Provider,
			key,
		)
		if err != nil {
			slog.ErrorContext(ctx, "tryWithModel", "error", err)
			continue
		}

		return resp, nil
	}

	return nil, err
}

func (r *Router) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	provider Provider,
	key APIKey,
) (*CompletionResponse, error) {
	var resp *CompletionResponse
	var err error

	switch provider {
	case ProviderOpenAI:
		resp, err = r.llms[ProviderOpenAI].completeResponse(
			ctx,
			req,
			key,
		)
	case ProviderGoogle:
		resp, err = r.llms[ProviderGoogle].completeResponse(
			ctx,
			req,
			key,
		)
	default:
		err = ErrUnsupportedProvider
	}

	return resp, err
}

func (ak *APIKey) isAvailable() bool {
	return ak.RequestRemaining > 1
}
