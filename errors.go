package heimdall

import "errors"

var (
	ErrRateLimitHit        = errors.New("rate limit exceeded")
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrNoChunkHandler      = errors.New(
		"a chunk handler must be provided to stream response",
	)
)
