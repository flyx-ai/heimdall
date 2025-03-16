package heimdall

import (
	"context"
	"fmt"
	"time"
)

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

func (r *Router) tryWithModel(
	ctx context.Context,
	req CompletionRequest,
	model Model,
	requestLog *Logging,
) (CompletionResponse, error) {
	provider := r.providers[string(model.Provider)]
	res, err := provider.completeResponse(ctx, req, r.client, requestLog)
	if err != nil {
		return CompletionResponse{}, err
	}

	return res, nil
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
	// case ProviderOpenAI:
	// 	resp, err = r.llms[ProviderOpenAI].completeResponse(
	// 		ctx,
	// 		req,
	// 		key,
	// 	)
	// case ProviderGoogle:
	// 	resp, err = r.llms[ProviderGoogle].completeResponse(
	// 		ctx,
	// 		req,
	// 		key,
	// 	)
	// case ProviderGoogleVertexAI:
	// 	resp, err = r.llms[ProviderGoogleVertexAI].completeResponse(
	// 		ctx,
	// 		req,
	// 		key,
	// 	)
	// default:
	// 	requestLog.Events = append(requestLog.Events, Event{
	// 		Timestamp: time.Now(),
	// 		Description: fmt.Sprintf(
	// 			"unsupported provider: %s passed to completeResponse",
	// 			provider,
	// 		),
	// 	})
	// 	err = ErrUnsupportedProvider
	}

	return resp, err
}
