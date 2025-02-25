package heimdall

import (
	"net/http"
	"strconv"
	"time"
)

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
type CompletionResponse struct {
	Content    string
	Model      Model
	Usage      Usage
	RequestLog Logging
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

func parseOpenAICompatRateLimit(resp *http.Response) RateLimit {
	return RateLimit{
		Remaining: parseInt(resp.Header.Get("x-ratelimit-remaining-requests")),
		Limit:     parseInt(resp.Header.Get("x-ratelimit-limit-requests")),
		Reset: time.Now().
			Add(time.Duration(parseInt(resp.Header.Get("x-ratelimit-reset-requests"))) * time.Second),
	}
}
