package heimdall

import (
	"context"
	"net/http"
	"time"
)

// APIKey is matched with a provider and will keep track of request limits returned from each provider.
// Some providers requires you to add a location and project id to the request (e.g. if you're using vertex ai). These are optional. The only required field are 'Key'.
type APIKey struct {
	Name             string
	Key              string
	Location         string
	ProjectID        string
	RequestsLimit    int
	requestsUsed     int
	RequestRemaining int
	ResetAt          time.Time
}

type RouterConfig struct {
	// ProviderAPIKeys map[Provider][]APIKey
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
		req CompletionRequest,
		key APIKey,
		chunkHandler func(chunk string) error,
	) (CompletionResponse, error)
	getApiKeys() []string
	name() string
	// setClient(c http.Client) *LLMProvider
}

type Router struct {
	// config RouterConfig
	providers map[string]LLMProvider
	// timeout   time.Duration
	client http.Client
	// llms   map[Provider]LLM
}

//	func (r *Router) ReqsStats() {
//		for provider, keys := range r.config.ProviderAPIKeys {
//			for _, key := range keys {
//				slog.Info(
//					"API KEY STATS",
//					"provider",
//					provider,
//					"key_name",
//					key.Name,
//					"requests_remaining",
//					key.RequestRemaining,
//					"requests_used",
//					key.requestsUsed,
//				)
//			}
//		}
//	}

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

func (r *Router) Stream(
	ctx context.Context,
	req CompletionRequest,
	chunkHandler func(chunk string) error,
) (CompletionResponse, error) {
	if chunkHandler == nil {
		return CompletionResponse{}, ErrNoChunkHandler
	}

	models := append([]Model{req.Model}, req.Fallback...)
	var resp CompletionResponse
	var err error

	for _, model := range models {
		resp, err = r.tryStreamWithModel(
			ctx,
			req,
			model,
			chunkHandler,
		)
		if err == nil {
			break
		}
	}

	req.Tags["request_type"] = "stream"

	return resp, err
}

func (r *Router) tryStreamWithModel(
	ctx context.Context,
	req CompletionRequest,
	model Model,
	chunkHandler func(chunk string) error,
) (CompletionResponse, error) {
	// keys := r.config.ProviderAPIKeys[model.Provider]
	var err error
	//
	// for i := range keys {
	// 	key := keys[i]
	// 	if !key.isAvailable() {
	// 		continue
	// 	}
	//
	// 	resp, err := r.streamResponse(
	// 		ctx,
	// 		req,
	// 		model.Provider,
	// 		key,
	// 		chunkHandler,
	// 	)
	// 	if err != nil {
	// 		slog.ErrorContext(ctx, "tryStreamWithModel", "error", err)
	// 		continue
	// 	}
	//
	// 	return resp, nil
	// }

	return CompletionResponse{}, err
}

func (r *Router) streamResponse(
	ctx context.Context,
	req CompletionRequest,
	provider Provider,
	key APIKey,
	chunkHandler func(chunk string) error,
) (CompletionResponse, error) {
	var resp CompletionResponse
	var err error

	switch provider {
	// case ProviderOpenAI:
	// 	resp, err = r.llms[ProviderOpenAI].streamResponse(
	// 		ctx,
	// 		req,
	// 		key,
	// 		chunkHandler,
	// 	)
	// case ProviderGoogle:
	// 	resp, err = r.llms[ProviderGoogle].streamResponse(
	// 		ctx,
	// 		req,
	// 		key,
	// 		chunkHandler,
	// 	)
	// case ProviderGoogleVertexAI:
	// 	resp, err = r.llms[ProviderGoogleVertexAI].streamResponse(
	// 		ctx,
	// 		req,
	// 		key,
	// 		chunkHandler,
	// 	)
	default:
		err = ErrUnsupportedProvider
	}

	return resp, err
}
