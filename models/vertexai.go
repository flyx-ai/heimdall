package models

const vertexProvider = "google"

type VertexGemini15FlashThinking struct{}

// GetName implements Model.
func (v *VertexGemini15FlashThinking) GetName() string {
	return "gemini-1.5-flash-002"
}

// GetProvider implements Model.
func (v *VertexGemini15FlashThinking) GetProvider() string {
	return vertexProvider
}

var _ Model = new(VertexGemini15FlashThinking)

type VertexGemini15Pro struct{}

// GetName implements Model.
func (v *VertexGemini15Pro) GetName() string {
	return "gemini-1.5-pro-002"
}

// GetProvider implements Model.
func (v *VertexGemini15Pro) GetProvider() string {
	return vertexProvider
}

var _ Model = new(VertexGemini15Pro)

type VertexGemini10ProVision struct{}

// GetName implements Model.
func (v *VertexGemini10ProVision) GetName() string {
	return "gemini-1.0-pro-vision-001"
}

// GetProvider implements Model.
func (v *VertexGemini10ProVision) GetProvider() string {
	return vertexProvider
}

var _ Model = new(VertexGemini10ProVision)

type VertexGemini10Pro struct{}

// GetName implements Model.
func (v *VertexGemini10Pro) GetName() string {
	return "gemini-1.0-pro-002"
}

// GetProvider implements Model.
func (v *VertexGemini10Pro) GetProvider() string {
	return vertexProvider
}

var _ Model = new(VertexGemini10Pro)

type VertexGemini20Flash struct{}

// GetName implements Model.
func (v *VertexGemini20Flash) GetName() string {
	return "gemini-2.0-flash-001"
}

// GetProvider implements Model.
func (v *VertexGemini20Flash) GetProvider() string {
	return vertexProvider
}

var _ Model = new(VertexGemini20Flash)

type VertexGemini20FlashLite struct{}

// GetName implements Model.
func (v *VertexGemini20FlashLite) GetName() string {
	return "gemini-2.0-flash-lite-001"
}

// GetProvider implements Model.
func (v *VertexGemini20FlashLite) GetProvider() string {
	return vertexProvider
}

var _ Model = new(VertexGemini20FlashLite)
