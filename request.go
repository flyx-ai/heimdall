package heimdall

type CompletionRequest struct {
	Model       Model
	Messages    []Message
	Fallback    []Model
	Temperature float32
	TopP        float32
	Tags        map[string]string `json:"tags"`
}

type Message struct {
	Role    string
	Content string
}
