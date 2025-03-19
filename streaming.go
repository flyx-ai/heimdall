package heimdall

import (
	"context"
	"fmt"
	"time"
)

func (r *Router) Stream(
	ctx context.Context,
	req CompletionRequest,
	chunkHandler func(chunk string) error,
) (CompletionResponse, error) {
	now := time.Now()

	if chunkHandler == nil {
		return CompletionResponse{}, ErrNoChunkHandler
	}

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

	req.Tags["request_type"] = "stream"

	models := append([]Model{req.Model}, req.Fallback...)
	var resp CompletionResponse
	var err error

	requestLog := Logging{
		Events: []Event{
			{
				Timestamp:   now,
				Description: "start of call to Stream",
			},
		},
		SystemMsg: systemMsg,
		UserMsg:   userMsg,
		Start:     now,
	}

	for _, model := range models {
		if r.providers[model.Provider.name()] == nil {
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"attempting tryStreamWithModel using model: %s but provider: %s not registered on router. attempting with next model.",
					model.Name,
					model.Provider.name(),
				),
			})

			continue
		}

		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting tryStreamWithModel using model: %s",
				model.Name,
			),
		})
		resp, err = r.tryStreamWithModel(
			ctx,
			req,
			model,
			chunkHandler,
			&requestLog,
		)
		if err == nil {
			break
		}
	}

	return resp, err
}

func (r *Router) tryStreamWithModel(
	ctx context.Context,
	req CompletionRequest,
	model Model,
	chunkHandler func(chunk string) error,
	requestLog *Logging,
) (CompletionResponse, error) {
	provider := r.providers[model.Provider.name()]
	res, err := provider.streamResponse(
		ctx,
		r.client,
		req,
		chunkHandler,
		requestLog,
	)
	if err != nil {
		return CompletionResponse{}, err
	}

	return res, nil
}
