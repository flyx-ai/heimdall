package heimdall

type Provider string

const (
	ProviderOpenAI     Provider = "openai"
	ProviderAnthropic  Provider = "anthropic"
	ProviderGoogle     Provider = "google"
	ProviderPerplexity Provider = "perplexity"

	anthropicBaseURL  = "https://api.anthropic.com/v1"
	googleBaseUrl     = "https://generativelanguage.googleapis.com/v1beta:chatCompletions"
	perplexityBaseUrl = "https://api.perplexity.ai/chat/completions"
)
