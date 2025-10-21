package models

const VertexProvider = "vertexai"

// NOTE: VertexGemini15FlashThinking and VertexGemini15Pro types have been removed as these models were retired by Google in 2025

type VertexGemini20Flash struct{}

func (v VertexGemini20Flash) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000010
}

func (v VertexGemini20Flash) GetName() string {
	return "gemini-2.0-flash-001"
}

func (v VertexGemini20Flash) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini20Flash)

type VertexGemini20FlashLite struct{}

func (v VertexGemini20FlashLite) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000000075
}

func (v VertexGemini20FlashLite) GetName() string {
	return "gemini-2.0-flash-lite-001"
}

func (v VertexGemini20FlashLite) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini20FlashLite)
