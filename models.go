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

var ModelGPT4O = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4o",
}

var ModelGPT4OMini Model = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4o-mini",
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

var ModelGemini20Flash = Model{
	Provider: ProviderGoogle,
	Name:     "gemini-2.0-flash-001",
}

var ModelGemini20FlashLite = Model{
	Provider: ProviderGoogle,
	Name:     "gemini-2.0-flash-lite-001",
}

var ModelVertexGemini15FlashThinking = Model{
	Provider: ProviderGoogleVertexAI,
	Name:     "gemini-1.5-flash-002",
}

var ModelVertexGemini15Pro = Model{
	Provider: ProviderGoogleVertexAI,
	Name:     "gemini-1.5-pro-002",
}

var ModelVertexGemini10ProVision = Model{
	Provider: ProviderGoogleVertexAI,
	Name:     "gemini-1.0-pro-vision-001",
}

var ModelVertexGemini10Pro = Model{
	Provider: ProviderGoogleVertexAI,
	Name:     "gemini-1.0-pro-002",
}

var ModelVertexGemini20Flash = Model{
	Provider: ProviderGoogleVertexAI,
	Name:     "gemini-2.0-flash-001",
}

var ModelVertexGemini20FlashLite = Model{
	Provider: ProviderGoogleVertexAI,
	Name:     "gemini-2.0-flash-lite-001",
}
