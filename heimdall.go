package heimdall

import (
	"context"
	"fmt"
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
	) (CompletionResponse, error)
	streamResponse(
		ctx context.Context,
		req CompletionRequest,
		key APIKey,
		chunkHandler func(chunk string) error,
	) (CompletionResponse, error)
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
		Timeout: config.Timeout * time.Second,
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
) (CompletionResponse, error) {
	now := time.Now()
	var systemMsg string
	var userMsg string
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemMsg = msg.Content
		}
		if msg.Role == "user" {
			userMsg = msg.Content
		}
	}
	requestLog := Logging{
		Events: []Event{
			{
				Timestamp:   now,
				Description: "start of call to Complete",
			},
		},
		SystemMsg: systemMsg,
		UserMsg:   userMsg,
		Start:     now,
	}

	models := append([]Model{req.Model}, req.Fallback...)
	var err error
	resp := CompletionResponse{}

	for _, model := range models {
		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting tryWithModel using model: %s",
				model.Name,
			),
		})
		resp, err = r.tryWithModel(ctx, req, model, &requestLog)
		if err == nil {
			break
		}

		continue
	}

	if err == nil {
		requestLog.Response = resp.Content
		requestLog.Completed = true
	}
	if err != nil {
		requestLog.Completed = false
	}

	requestLog.End = time.Now()

	resp.RequestLog = requestLog

	return resp, err
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
	requestLog *Logging,
) (CompletionResponse, error) {
	keys := r.config.ProviderAPIKeys[model.Provider]
	var err error

	for i := range keys {
		key := keys[i]

		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to call completeResponse using key: %s and model: %s",
				key.Name,
				model.Name,
			),
		})

		if !key.isAvailable() {
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"key: %s not available for provider: %s",
					key.Name,
					model.Provider,
				),
			})

			continue
		}

		resp, err := r.completeResponse(
			ctx,
			req,
			model.Provider,
			key,
			requestLog,
		)
		if err != nil {
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"call to completeResponse failed with err: %s and model: %s",
					err.Error(),
					model.Name,
				),
			})
			continue
		}

		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"call to completeResponse with key: %s and model: %s finised successfully",
				key.Name,
				model.Name,
			),
		})

		return resp, nil
	}

	requestLog.Events = append(requestLog.Events, Event{
		Timestamp: time.Now(),
		Description: fmt.Sprintf(
			"call to tryWithModel failed with err: %v",
			err,
		),
	})

	return CompletionResponse{}, err
}

func (r *Router) completeResponse(
	ctx context.Context,
	req CompletionRequest,
	provider Provider,
	key APIKey,
	requestLog *Logging,
) (CompletionResponse, error) {
	var resp CompletionResponse
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
		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"unsupported provider: %s passed to completeResponse",
				provider,
			),
		})
		err = ErrUnsupportedProvider
	}

	return resp, err
}

func (ak *APIKey) isAvailable() bool {
	return ak.RequestRemaining > 1
}
