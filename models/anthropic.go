package models

const AnthropicProvider = "anthropic"

const (
	AnthropicClaude3OpusAlias    = "claude-3-opus-latest"
	AnthropicClaude35SonnetAlias = "claude-3-5-sonnet-latest"
	AnthropicClaude35HaikuAlias  = "claude-3-5-haiku-latest"
	AnthropicClaude37SonnetAlias = "claude-3-7-sonnet-latest"
	AnthropicClaude4SonnetAlias  = "claude-sonnet-4-20250514"
	AnthropicClaude4OpusAlias    = "claude-opus-4-20250514"
)

type (
	AnthropicImageType string
	AnthropicPdf       string
)

const (
	AnthropicImageJpeg AnthropicImageType = "image/jpeg"
	AnthropicImagePng  AnthropicImageType = "image/png"
	AnthropicImageGif  AnthropicImageType = "image/gif"
	AnthropicImageWebp AnthropicImageType = "image/webp"
)

type Claude3Opus struct {
	ImageFile map[AnthropicImageType]string
	PdfFiles  []AnthropicPdf
}

func (c Claude3Opus) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000015
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
	PdfFiles  []AnthropicPdf
}

func (c Claude35Sonnet) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000003
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
	PdfFiles  []AnthropicPdf
}

func (c Claude35Haiku) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.0000008
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
	PdfFiles  []AnthropicPdf
}

func (c Claude37Sonnet) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000003
}

func (c Claude37Sonnet) GetName() string {
	return AnthropicClaude37SonnetAlias
}

func (c Claude37Sonnet) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude37Sonnet)

type Claude4Sonnet struct {
	ImageFile map[AnthropicImageType]string
	PdfFiles  []AnthropicPdf
}

func (c Claude4Sonnet) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000003
}

func (c Claude4Sonnet) GetName() string {
	return AnthropicClaude4SonnetAlias
}

func (c Claude4Sonnet) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude4Sonnet)

type Claude4Opus struct {
	ImageFile map[AnthropicImageType]string
	PdfFiles  []AnthropicPdf
}

func (c Claude4Opus) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000015
}

func (c Claude4Opus) GetName() string {
	return AnthropicClaude4OpusAlias
}

func (c Claude4Opus) GetProvider() string {
	return AnthropicProvider
}

var _ Model = new(Claude4Opus)
