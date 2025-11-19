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
	req request.Completion,
) (response.Completion, error) {
	now := time.Now()

	req.Tags["request_type"] = "completion"

	requestLog := response.Logging{
		Events: []response.Event{
			{
				Timestamp:   now,
				Description: "start of call to Complete",
			},
		},
		SystemMsg: req.SystemMessage,
		UserMsg:   req.UserMessage,
		Start:     now,
	}

	models := append([]models.Model{req.Model}, req.Fallback...)
	var err error
	resp := response.Completion{}

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
	}

	requestLog.Completed = err == nil
	if err == nil {
		requestLog.Response = resp.Content
	}

	requestLog.End = time.Now()

	resp.RequestLog = requestLog

	return resp, err
}

func (r *Router) tryWithModel(
	ctx context.Context,
	req request.Completion,
	model models.Model,
	requestLog *response.Logging,
) (response.Completion, error) {
	provider := r.providers[model.GetProvider()]
	return provider.CompleteResponse(ctx, req, r.client, requestLog)
}
