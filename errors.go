package heimdall

import "errors"

var (
	ErrNoAvailableKeys     = errors.New("no available API keys")
	errInternalServer      = errors.New("request returned status code 500")
	errTooManyRequests     = errors.New("request returned status code 429")
	ErrRequestFailed       = errors.New("request failed")
	ErrRateLimitHit        = errors.New("rate limit exceeded")
	errBadRequest          = errors.New("bad request")
	ErrSlowResponse        = errors.New("slow response")
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrNoChunkHandler      = errors.New(
		"a chunk handler must be provided to stream response",
	)
	ErrProjectIDOrLocationMissing = errors.New(
		"the model or provider passed requires you to use an api key that specifies location and project id",
	)
)
