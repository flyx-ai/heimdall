package models

type Model interface {
	GetProvider() string
	GetName() string
	EstimateCost(text string) float64
}

type StructuredOutput interface {
	GetStructuredOutput() map[string]any
}

type FileReader interface {
	GetFileData() map[string][]byte
}

// GetAll returns all model names
func GetAll() []string {
	return []string{
		AnthropicClaude3OpusAlias,
		AnthropicClaude35SonnetAlias,
		AnthropicClaude35HaikuAlias,
		AnthropicClaude37SonnetAlias,

		Gemini15FlashModel,
		Gemini15ProModel,
		Gemini20FlashModel,
		Gemini20FlashLiteModel,
		Gemini25FlashPreviewModel,
		Gemini25ProPreviewModel,

		O3MiniAlias,
		GPT4OAlias,
		GPT4OMiniAlias,
		O1Alias,
		GPT4Alias,
		GPT4TurboAlias,
		GPT41Alias,

		"sonar-reasoning-pro",
		"sonar-reasoning",
		"sonar-pro",
		"sonar",

		"gemini-1.5-flash-002",
		"gemini-1.5-pro-002",
		"gemini-2.0-flash-001",
		"gemini-2.0-flash-lite-001",

		GrokBetaAlias,
		Grok2Alias,
		Grok2MiniAlias,
		Grok2VisionAlias,
		Grok3Alias,
		Grok3MiniAlias,
		Grok3FastAlias,
		Grok3MiniFastAlias,
		Grok4Alias,
	}
}
