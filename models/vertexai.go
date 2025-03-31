package models

const VertexProvider = "vertexai"

type VertexGemini15FlashThinking struct{}

func (v VertexGemini15FlashThinking) GetName() string {
	return "gemini-1.5-flash-002"
}

func (v VertexGemini15FlashThinking) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini15FlashThinking)

type VertexGemini15Pro struct{}

func (v VertexGemini15Pro) GetName() string {
	return "gemini-1.5-pro-002"
}

func (v VertexGemini15Pro) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini15Pro)

type VertexGemini20Flash struct{}

func (v VertexGemini20Flash) GetName() string {
	return "gemini-2.0-flash-001"
}

func (v VertexGemini20Flash) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini20Flash)

type VertexGemini20FlashLite struct{}

func (v VertexGemini20FlashLite) GetName() string {
	return "gemini-2.0-flash-lite-001"
}

func (v VertexGemini20FlashLite) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini20FlashLite)
