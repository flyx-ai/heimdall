package models

const AnthropicProvider = "anthropic"

type Claude3Opus struct{}

func (c Claude3Opus) GetName() string {
	return "claude-3-opus-latest"
}

func (c Claude3Opus) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude3Opus)

type Claude35Sonnet struct{}

func (c Claude35Sonnet) GetName() string {
	return "claude-3-5-sonnet-latest"
}

func (c Claude35Sonnet) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude35Sonnet)

type Claude35Haiku struct{}

func (c Claude35Haiku) GetName() string {
	return "claude-3-5-haiku-latest"
}

func (c Claude35Haiku) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude35Haiku)

type Claude37Sonnet struct{}

func (c Claude37Sonnet) GetName() string {
	return "claude-3-7-sonnet-latest"
}

func (c Claude37Sonnet) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude37Sonnet)
