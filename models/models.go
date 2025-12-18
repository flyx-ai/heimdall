package models

type Model interface {
	GetProvider() string
	GetName() string
	EstimateCost(text string) float64
}

type CostBreakdown interface {
	GetInputCostPer1M() float64
	GetOutputCostPer1M() float64
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
		AnthropicClaude4SonnetAlias,
		AnthropicClaude4OpusAlias,
		AnthropicClaude45HaikuAlias,
		AnthropicClaude45OpusAlias,

		// NOTE: Gemini 1.5 models have been retired by Google as of 2025
		Gemini20FlashModel,
		Gemini20FlashLiteModel,
		Gemini25FlashModel,
		Gemini25ProModel,
		Gemini3ProModel,
		Gemini3ProImageModel,
		Gemini3FlashModel,

		O3MiniAlias,
		GPT4OAlias,
		GPT4OMiniAlias,
		O1Alias,
		GPT4Alias,
		GPT4TurboAlias,
		GPT41Alias,
		GPT5Alias,
		GPT5MiniAlias,
		GPT5NanoAlias,
		GPT5ChatAlias,
		GPT51Alias,
		GPT51ChatAlias,
		GPT51CodexAlias,
		GPT51CodexMiniAlias,

		"sonar-reasoning-pro",
		"sonar-reasoning",
		"sonar-pro",
		"sonar",

		"gemini-2.0-flash-001",
		"gemini-2.0-flash-lite-001",

		Grok2VisionAlias,
		Grok3Alias,
		Grok3MiniAlias,
		Grok3FastAlias,
		Grok3MiniFastAlias,
		Grok4Alias,
		Grok4FastAlias,

		Gemini25FlashImageModel,
	}
}
