package heimdall

type Model struct {
	Provider LLMProvider
	Name     string
}

var ModelO3Mini Model = Model{
	Provider: Openai{},
	Name:     "o1-mini",
}

var ModelO1 Model = Model{
	Provider: Openai{},
	Name:     "o1",
}

var ModelO1Mini Model = Model{
	Provider: Openai{},
	Name:     "o1-mini",
}

var ModelO1Preview Model = Model{
	Provider: Openai{},
	Name:     "o1-preview",
}

var ModelGPT4 Model = Model{
	Provider: Openai{},
	Name:     "gpt-4",
}

var ModelGPT4Turbo Model = Model{
	Provider: Openai{},
	Name:     "gpt-4-turbo",
}

var ModelGPT4O = Model{
	Provider: Openai{},
	Name:     "gpt-4o",
}

var ModelGPT4OMini Model = Model{
	Provider: Openai{},
	Name:     "gpt-4o-mini",
}

var ModelGemini15FlashThinking = Model{
	Provider: Google{},
	Name:     "gemini-1.5-flash-002",
}

var ModelGemini15Pro = Model{
	Provider: Google{},
	Name:     "gemini-1.5-pro-002",
}

var ModelGemini10ProVision = Model{
	Provider: Google{},
	Name:     "gemini-1.0-pro-vision-001",
}

var ModelGemini10Pro = Model{
	Provider: Google{},
	Name:     "gemini-1.0-pro-002",
}

var ModelGemini20Flash = Model{
	Provider: Google{},
	Name:     "gemini-2.0-flash-001",
}

var ModelGemini20FlashLite = Model{
	Provider: Google{},
	Name:     "gemini-2.0-flash-lite-001",
}

var ModelVertexGemini15FlashThinking = Model{
	Provider: Google{},
	Name:     "gemini-1.5-flash-002",
}

var ModelVertexGemini15Pro = Model{
	Provider: Google{},
	Name:     "gemini-1.5-pro-002",
}

var ModelVertexGemini10ProVision = Model{
	Provider: Google{},
	Name:     "gemini-1.0-pro-vision-001",
}

var ModelVertexGemini10Pro = Model{
	Provider: Google{},
	Name:     "gemini-1.0-pro-002",
}

var ModelVertexGemini20Flash = Model{
	Provider: Google{},
	Name:     "gemini-2.0-flash-001",
}

var ModelVertexGemini20FlashLite = Model{
	Provider: Google{},
	Name:     "gemini-2.0-flash-lite-001",
}

var ModelClaude3Opus Model = Model{
	Provider: Anthropic{},
	Name:     "claude-3-opus-latest",
}

var ModelClaude3Sonnet = Model{
	Provider: Anthropic{},
	Name:     "claude-3-sonnet-latest",
}

var ModelClaude3Haiku = Model{
	Provider: Anthropic{},
	Name:     "claude-3-haiku-latest",
}

var ModelSonarReasoningPro = Model{
	Provider: Perplexity{},
	Name:     "sonar-reasoning-pro",
}

var ModelSonarReasoning = Model{
	Provider: Perplexity{},
	Name:     "sonar-reasoning",
}

var ModelSonarPro = Model{
	Provider: Perplexity{},
	Name:     "sonar-pro",
}

var ModelSonar = Model{
	Provider: Perplexity{},
	Name:     "sonar",
}
