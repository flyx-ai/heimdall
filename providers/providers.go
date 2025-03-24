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
		req request.CompletionRequest,
		client http.Client,
		requestLog *response.Logging,
	) (response.CompletionResponse, error)
	StreamResponse(
		ctx context.Context,
		client http.Client,
		req request.CompletionRequest,
		chunkHandler func(chunk string) error,
		requestLog *response.Logging,
	) (response.CompletionResponse, error)
	tryWithBackup(
		ctx context.Context,
		req request.CompletionRequest,
		client http.Client,
		chunkHandler func(chunk string) error,
		requestLog *response.Logging,
	) (response.CompletionResponse, error)
	doRequest(
		ctx context.Context,
		req request.CompletionRequest,
		client http.Client,
		chunkHandler func(chunk string) error,
		key string,
	) (response.CompletionResponse, int, error)
	GetApiKeys() []string
	Name() string
}
