package heimdall

import (
	"context"
	"fmt"
	"time"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

func (r *Router) Stream(
	ctx context.Context,
	req request.Completion,
	chunkHandler func(chunk string) error,
) (response.Completion, error) {
	now := time.Now()

	if chunkHandler == nil {
		return response.Completion{}, ErrNoChunkHandler
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

	models := append([]models.Model{req.Model}, req.Fallback...)
	var resp response.Completion
	var err error

	requestLog := response.Logging{
		Events: []response.Event{
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
		if r.providers[model.GetProvider()] == nil {
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"attempting tryStreamWithModel using model: %s but provider: %s not registered on router. attempting with next model.",
					model.GetName(),
					model.GetProvider(),
				),
			})

			continue
		}

		requestLog.Events = append(requestLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting tryStreamWithModel using model: %s",
				model.GetName(),
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
	req request.Completion,
	model models.Model,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	provider := r.providers[model.GetProvider()]
	res, err := provider.StreamResponse(
		ctx,
		r.client,
		req,
		chunkHandler,
		requestLog,
	)
	if err != nil {
		return response.Completion{}, err
	}

	return res, nil
}
