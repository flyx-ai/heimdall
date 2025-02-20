package heimdall

import "errors"

var (
	ErrNoAvailableKeys     = errors.New("no available API keys")
	ErrRequestFailed       = errors.New("request failed")
	ErrRateLimitHit        = errors.New("rate limit exceeded")
	ErrSlowResponse        = errors.New("slow response")
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrNoChunkHandler      = errors.New(
		"a chunk handler must be provided to stream response",
	)
)
