package response

import (
	"time"

	"github.com/flyx-ai/heimdall/models"
)

type Event struct {
	Timestamp   time.Time
	Description string
}

type Logging struct {
	Completed bool
	Start     time.Time
	End       time.Time
	Events    []Event
	Model     models.Model
	SystemMsg string
	UserMsg   string
	Response  string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
type Completion struct {
	Content    string
	Model      string
	Usage      Usage
	RequestLog Logging
}
