package models

const AnthropicProvider = "anthropic"

type Claude3Opus struct{}

// GetName implements Model.
func (c *Claude3Opus) GetName() string {
	return "claude-3-opus-latest"
}

// GetProvider implements Model.
func (c *Claude3Opus) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude3Opus)

type Claude3Sonnet struct{}

// GetName implements Model.
func (c *Claude3Sonnet) GetName() string {
	return "claude-3-sonnet-latest"
}

// GetProvider implements Model.
func (c *Claude3Sonnet) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude3Sonnet)

type Claude3Haiku struct{}

// GetName implements Model.
func (c *Claude3Haiku) GetName() string {
	return "claude-3-haiku-latest"
}

// GetProvider implements Model.
func (c *Claude3Haiku) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude3Haiku)
