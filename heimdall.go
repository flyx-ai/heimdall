package heimdall

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type APIKey struct {
	Key              string
	RequestsLimit    int
	requestsUsed     int
	requestRemaining int
	ResetAt          time.Time
	mu               sync.Mutex
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
	mu     sync.RWMutex
	llms   map[Provider]LLM
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
		return errors.New("a chunk handler must be provided to stream response")
	}

	models := append([]Model{req.Model}, req.Fallback...)
	// groupingID := uuid.New()
	var resp *CompletionResponse
	var err error

	for _, model := range models {
		resp, err = r.tryStreamWithModel(
			ctx,
			req,
			model,
			// groupingID,
			chunkHandler,
		)
		if resp != nil && err == nil {
			break
		}
	}

	// var systemMsg string
	// var userMsg string
	// messages := make([]openAIMsg, len(req.Messages))
	// for i, msg := range req.Messages {
	// 	if msg.Role == "system" {
	// 		systemMsg = msg.Content
	// 	}
	// 	if msg.Role == "user" {
	// 		userMsg = msg.Content
	// 	}
	// 	messages[i] = openAIMsg(msg)
	// }

	req.Tags["request_type"] = "stream"

	// tags, err := json.Marshal(req.Tags)
	// if err != nil {
	// 	slog.ErrorContext(ctx, "log final request and response", "error", err)
	// }

	// if _, err := clients.RiverQueue.Insert(ctx, jobs.HeimdallRequestArgs{
	// 	ID:           groupingID,
	// 	Success:      err == nil,
	// 	TopP:         req.TopP,
	// 	Temperature:  req.Temperature,
	// 	SystemPrompt: systemMsg,
	// 	UserPrompt:   userMsg,
	// 	Response:     resp.Content,
	// 	Model:        resp.Model.Name,
	// 	Tags:         tags,
	// }, &river.InsertOpts{}); err != nil {
	// 	slog.ErrorContext(ctx, "log final request and response", "error", err)
	// }

	return err
}

func (r *Router) tryStreamWithModel(
	ctx context.Context,
	req CompletionRequest,
	model Model,
	// groupingID uuid.UUID,
	chunkHandler func(chunk string) error,
) (*CompletionResponse, error) {
	keys := r.config.ProviderAPIKeys[model.Provider]
	var err error

	for i := range keys {
		key := &keys[i]
		if !key.isAvailable() {
			//}
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
	// case ProviderGoogle:
	// 	resp, err = c.streamResponseGoogle(ctx, req, chunkHandler)
	// case ProviderPerplexity:
	// 	resp, err = c.streamResponsePerplexity(ctx, req, chunkHandler)
	// case ProviderAnthropic:
	// 	resp, err = c.streamResponseAnthropic(ctx, req, chunkHandler)
	default:
		err = fmt.Errorf("unsupported provider: %s", provider)
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

// func (r *Router) getKeysForProvider(provider Provider) []*APIKey {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()
//
// 	var keys []*APIKey
// 	for _, key := range r.config.Keys {
// 		if key.Provider == provider {
// 			keys = append(keys, key)
// 		}
// 	}
// 	return keys
// }

func (ak *APIKey) handleKeyError(key *APIKey, err error) {
	key.mu.Lock()
	defer key.mu.Unlock()

	if errors.Is(err, ErrRateLimitHit) {
		key.requestsUsed = key.RequestsLimit
	} else {
		key.requestsUsed++
	}
}
