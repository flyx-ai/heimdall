package models

const VertexProvider = "vertexai"

// NOTE: VertexGemini15FlashThinking and VertexGemini15Pro types have been removed as these models were retired by Google in 2025

// Gemini 2.0 Models

type VertexGemini20Flash struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	Thinking         ThinkBudget
}

func (v VertexGemini20Flash) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000010
}

func (v VertexGemini20Flash) GetInputCostPer1M() float64 {
	return 0.10
}

func (v VertexGemini20Flash) GetOutputCostPer1M() float64 {
	return 0.40
}

func (v VertexGemini20Flash) GetName() string {
	return "gemini-2.0-flash-001"
}

func (v VertexGemini20Flash) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini20Flash)
var _ CostBreakdown = new(VertexGemini20Flash)

type VertexGemini20FlashLite struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	Thinking         ThinkBudget
}

func (v VertexGemini20FlashLite) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000000075
}

func (v VertexGemini20FlashLite) GetInputCostPer1M() float64 {
	return 0.075
}

func (v VertexGemini20FlashLite) GetOutputCostPer1M() float64 {
	return 0.30
}

func (v VertexGemini20FlashLite) GetName() string {
	return "gemini-2.0-flash-lite-001"
}

func (v VertexGemini20FlashLite) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini20FlashLite)
var _ CostBreakdown = new(VertexGemini20FlashLite)

// Gemini 2.5 Models (GA)

type VertexGemini25Pro struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	Thinking         ThinkBudget
}

func (v VertexGemini25Pro) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000125
}

func (v VertexGemini25Pro) GetInputCostPer1M() float64 {
	return 1.25
}

func (v VertexGemini25Pro) GetOutputCostPer1M() float64 {
	return 10.0
}

func (v VertexGemini25Pro) GetName() string {
	return "gemini-2.5-pro"
}

func (v VertexGemini25Pro) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini25Pro)
var _ CostBreakdown = new(VertexGemini25Pro)

type VertexGemini25Flash struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	Thinking         ThinkBudget
}

func (v VertexGemini25Flash) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000015
}

func (v VertexGemini25Flash) GetInputCostPer1M() float64 {
	return 0.15
}

func (v VertexGemini25Flash) GetOutputCostPer1M() float64 {
	return 0.60
}

func (v VertexGemini25Flash) GetName() string {
	return "gemini-2.5-flash"
}

func (v VertexGemini25Flash) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini25Flash)
var _ CostBreakdown = new(VertexGemini25Flash)

type VertexGemini25FlashLite struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	Thinking         ThinkBudget
}

func (v VertexGemini25FlashLite) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000000075
}

func (v VertexGemini25FlashLite) GetInputCostPer1M() float64 {
	return 0.075
}

func (v VertexGemini25FlashLite) GetOutputCostPer1M() float64 {
	return 0.30
}

func (v VertexGemini25FlashLite) GetName() string {
	return "gemini-2.5-flash-lite"
}

func (v VertexGemini25FlashLite) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini25FlashLite)
var _ CostBreakdown = new(VertexGemini25FlashLite)

type VertexGemini25FlashImage struct {
	NumberOfImages int
	AspectRatio    AspectRatio
	ImageFile      []GoogleImagePayload
	PdfFiles       []GooglePdf
	Files          []GoogleFilePayload
}

func (v VertexGemini25FlashImage) EstimateCost(text string) float64 {
	// 1K/2K image = 1120 tokens at $30/1M = $0.0336 per image
	numImages := v.NumberOfImages
	if numImages == 0 {
		numImages = 1
	}
	return float64(numImages) * 0.0336
}

func (v VertexGemini25FlashImage) GetInputCostPer1M() float64 {
	return 0.30
}

func (v VertexGemini25FlashImage) GetOutputCostPer1M() float64 {
	return 30.0
}

func (v VertexGemini25FlashImage) GetName() string {
	return "gemini-2.5-flash-image"
}

func (v VertexGemini25FlashImage) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini25FlashImage)
var _ CostBreakdown = new(VertexGemini25FlashImage)

// Gemini 3 Models (Preview)

type VertexGemini3ProPreview struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	ThinkingLevel    ThinkingLevel
	MediaResolution  MediaResolution
}

func (v VertexGemini3ProPreview) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000002
}

func (v VertexGemini3ProPreview) GetInputCostPer1M() float64 {
	return 2.0
}

func (v VertexGemini3ProPreview) GetOutputCostPer1M() float64 {
	return 12.0
}

func (v VertexGemini3ProPreview) GetName() string {
	return "gemini-3-pro-preview"
}

func (v VertexGemini3ProPreview) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini3ProPreview)
var _ CostBreakdown = new(VertexGemini3ProPreview)

type VertexGemini3FlashPreview struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	ThinkingLevel    ThinkingLevel
	MediaResolution  MediaResolution
}

func (v VertexGemini3FlashPreview) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.0000005
}

func (v VertexGemini3FlashPreview) GetInputCostPer1M() float64 {
	return 0.50
}

func (v VertexGemini3FlashPreview) GetOutputCostPer1M() float64 {
	return 3.0
}

func (v VertexGemini3FlashPreview) GetName() string {
	return "gemini-3-flash-preview"
}

func (v VertexGemini3FlashPreview) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini3FlashPreview)
var _ CostBreakdown = new(VertexGemini3FlashPreview)

type VertexGemini3ProImagePreview struct {
	NumberOfImages  int
	AspectRatio     AspectRatio
	ImageFile       []GoogleImagePayload
	PdfFiles        []GooglePdf
	Files           []GoogleFilePayload
	ThinkingLevel   ThinkingLevel
	MediaResolution MediaResolution
}

func (v VertexGemini3ProImagePreview) EstimateCost(text string) float64 {
	// 1K/2K image = 1120 tokens at $120/1M = $0.134 per image
	numImages := v.NumberOfImages
	if numImages == 0 {
		numImages = 1
	}
	return float64(numImages) * 0.134
}

func (v VertexGemini3ProImagePreview) GetInputCostPer1M() float64 {
	return 2.0
}

func (v VertexGemini3ProImagePreview) GetOutputCostPer1M() float64 {
	return 120.0
}

func (v VertexGemini3ProImagePreview) GetName() string {
	return "gemini-3-pro-image-preview"
}

func (v VertexGemini3ProImagePreview) GetProvider() string {
	return VertexProvider
}

var _ Model = new(VertexGemini3ProImagePreview)
var _ CostBreakdown = new(VertexGemini3ProImagePreview)
