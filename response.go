package heimdall

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
