package heimdall

import (
	"context"
	"fmt"
	"time"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

func (r *Router) Complete(
	ctx context.Context,
	req request.CompletionRequest,
) (response.CompletionResponse, error) {
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

	requestLog := response.Logging{
		Events: []response.Event{
			{
				Timestamp:   now,
				Description: "start of call to Complete",
			},
		},
		SystemMsg: systemMsg,
		UserMsg:   userMsg,
		Start:     now,
	}

	models := append([]models.Model{req.Model}, req.Fallback...)
	var err error
	resp := response.CompletionResponse{}

	for _, model := range models {
		if r.providers[model.GetProvider()] == nil {
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"attempting tryWithModel using model: %s but provider: %s not registered on router. attempting with next model.",
					model.GetName(),
					model.GetProvider(),
				),
			})

			continue
		}
		requestLog.Events = append(requestLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting tryWithModel using model: %s",
				model.GetName(),
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
	req request.CompletionRequest,
	model models.Model,
	requestLog *response.Logging,
) (response.CompletionResponse, error) {
	provider := r.providers[model.GetProvider()]
	res, err := provider.CompleteResponse(ctx, req, r.client, requestLog)
	if err != nil {
		return response.CompletionResponse{}, err
	}

	return res, nil
}
