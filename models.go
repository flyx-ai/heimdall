package heimdall

type Model struct {
	Provider Provider
	Name     string
}

var ModelO3Mini Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1-mini",
}

var ModelO1 Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1",
}

var ModelO1Mini Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1-mini",
}

var ModelO1Preview Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1-preview",
}

var ModelGPT4 Model = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4",
}

var ModelGPT4Turbo Model = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4-turbo",
}

var ModelGPT4OModel = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4o",
}

var ModelGPT4OMini Model = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4o-mini",
}

var ModelClaude3Opus Model = Model{
	Provider: ProviderAnthropic,
	Name:     "claude-3-opus-20240229",
}

var ModelClaude3Sonnet = Model{
	Provider: ProviderAnthropic,
	Name:     "claude-3-sonnet-20240229",
}

var ModelClaude3Haiku = Model{
	Provider: ProviderAnthropic,
	Name:     "claude-3-haiku-20240307",
}

var ModelGemini15FlashThinking = Model{
	Provider: ProviderGoogle,
	Name:     "gemini-1.5-flash-002",
}

var ModelGemini15Pro = Model{
	Provider: ProviderGoogle,
	Name:     "gemini-1.5-pro-002",
}

var ModelGemini10ProVision = Model{
	Provider: ProviderGoogle,
	Name:     "gemini-1.0-pro-vision-001",
}

var ModelGemini10Pro = Model{
	Provider: ProviderGoogle,
	Name:     "gemini-1.0-pro-002",
}

var ModelSonarReasoningPro = Model{
	Provider: ProviderPerplexity,
	Name:     "sonar-reasoning-pro",
}

var ModelSonarReasoning = Model{
	Provider: ProviderPerplexity,
	Name:     "sonar-reasoning",
}

var ModelSonarPro = Model{
	Provider: ProviderPerplexity,
	Name:     "sonar-pro",
}

var ModelSonar = Model{
	Provider: ProviderPerplexity,
	Name:     "sonar",
}

func GetModelByName(name string) Model {
	switch name {
	case ModelO3Mini.Name:
		return ModelO3Mini
	case ModelO1.Name:
		return ModelO1
	case ModelO1Mini.Name:
		return ModelO1Mini
	case ModelO1Preview.Name:
		return ModelO1Preview
	case ModelGPT4.Name:
		return ModelGPT4
	case ModelGPT4Turbo.Name:
		return ModelGPT4Turbo
	case ModelGPT4OModel.Name:
		return ModelGPT4OModel
	case ModelGPT4OMini.Name:
		return ModelGPT4OMini
	case ModelClaude3Opus.Name:
		return ModelClaude3Opus
	case ModelClaude3Sonnet.Name:
		return ModelClaude3Sonnet
	case ModelClaude3Haiku.Name:
		return ModelClaude3Haiku
	case ModelGemini15FlashThinking.Name:
		return ModelGemini15FlashThinking
	case ModelGemini15Pro.Name:
		return ModelGemini15Pro
	case ModelGemini10ProVision.Name:
		return ModelGemini10ProVision
	case ModelGemini10Pro.Name:
		return ModelGemini10Pro
	}
	return ModelGPT4
}
