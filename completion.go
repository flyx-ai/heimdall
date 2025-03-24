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

	req.Tags["request_type"] = "completion"

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

	// models := append([]Model{req.Model}, req.Fallback...)
	var err error
	resp := CompletionResponse{}

	for _, model := range req.Models {
		if r.providers[model.Provider().name()] == nil {
			requestLog.Events = append(requestLog.Events, Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"attempting tryWithModel using model: %s but provider: %s not registered on router. attempting with next model.",
					model.Name,
					model.Provider.name(),
				),
			})

			continue
		}
		requestLog.Events = append(requestLog.Events, Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting tryWithModel using model: %s",
				// model.Name,
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
	provider := r.providers[model.Provider.name()]
	res, err := provider.completeResponse(ctx, req, r.client, requestLog)
	if err != nil {
		return CompletionResponse{}, err
	}

	return res, nil
}
