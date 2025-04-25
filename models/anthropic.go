package models

const AnthropicProvider = "anthropic"

const (
	AnthropicClaude3OpusAlias    = "claude-3-opus-latest"
	AnthropicClaude35SonnetAlias = "claude-3-5-sonnet-latest"
	AnthropicClaude35HaikuAlias  = "claude-3-5-haiku-latest"
	AnthropicClaude37SonnetAlias = "claude-3-7-sonnet-latest"
)

type (
	AnthropicImageType string
	AnthropicPdfType   string
)

const (
	AnthropicImageJpeg AnthropicImageType = "image/jpeg"
	AnthropicPdf       AnthropicPdfType   = "application/pdf"
)

type Claude3Opus struct {
	ImageFile map[AnthropicImageType]string
	PdfFiles  []map[AnthropicPdfType]string
}

func (c Claude3Opus) GetName() string {
	return AnthropicClaude3OpusAlias
}

func (c Claude3Opus) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude3Opus)

type Claude35Sonnet struct {
	ImageFile map[AnthropicImageType]string
	PdfFiles  []map[AnthropicPdfType]string
}

func (c Claude35Sonnet) GetName() string {
	return AnthropicClaude35SonnetAlias
}

func (c Claude35Sonnet) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude35Sonnet)

type Claude35Haiku struct {
	ImageFile map[AnthropicImageType]string
	PdfFiles  []map[AnthropicPdfType]string
}

func (c Claude35Haiku) GetName() string {
	return AnthropicClaude35HaikuAlias
}

func (c Claude35Haiku) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude35Haiku)

type Claude37Sonnet struct {
	ImageFile map[AnthropicImageType]string
	PdfFiles  []map[AnthropicPdfType]string
}

func (c Claude37Sonnet) GetName() string {
	return AnthropicClaude37SonnetAlias
}

func (c Claude37Sonnet) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude37Sonnet)
