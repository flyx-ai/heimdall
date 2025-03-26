package providers

import (
	"context"
	"net/http"

	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

type LLMProvider interface {
	CompleteResponse(
		ctx context.Context,
		req request.Completion,
		client http.Client,
		requestLog *response.Logging,
	) (response.Completion, error)
	StreamResponse(
		ctx context.Context,
		client http.Client,
		req request.Completion,
		chunkHandler func(chunk string) error,
		requestLog *response.Logging,
	) (response.Completion, error)
	tryWithBackup(
		ctx context.Context,
		req request.Completion,
		client http.Client,
		chunkHandler func(chunk string) error,
		requestLog *response.Logging,
	) (response.Completion, error)
	doRequest(
		ctx context.Context,
		req request.Completion,
		client http.Client,
		chunkHandler func(chunk string) error,
		key string,
	) (response.Completion, int, error)
	Name() string
}
