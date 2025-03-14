package heimdall

type Provider string

const (
	ProviderOpenAI         Provider = "openai"
	ProviderGoogle         Provider = "google"
	ProviderGoogleVertexAI Provider = "vertexai"
)
