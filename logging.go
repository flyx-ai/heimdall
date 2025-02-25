package heimdall

import "time"

type Event struct {
	Timestamp   time.Time
	Description string
}

type Logging struct {
	Completed bool
	Start     time.Time
	End       time.Time
	Events    []Event
	Model     Model
	SystemMsg string
	UserMsg   string
	Response  string
}
