package models

const GoogleProvider = "google"

const (
	// NOTE: Gemini 1.5 models (gemini-1.5-flash-002, gemini-1.5-pro-002) have been retired by Google as of 2025
	Gemini20FlashModel      = "gemini-2.0-flash-001"
	Gemini20FlashLiteModel  = "gemini-2.0-flash-lite-001"
	Gemini25FlashModel      = "gemini-2.5-flash"
	Gemini25FlashLiteModel  = "gemini-2.5-flash-lite"
	Gemini25ProModel        = "gemini-2.5-pro"
	Gemini25FlashImageModel = "gemini-2.5-flash-image"
	Gemini3ProModel         = "gemini-3-pro-preview"
	Gemini3ProImageModel    = "gemini-3-pro-image-preview"
	Gemini3FlashModel       = "gemini-3-flash-preview"
)

type ThinkBudget string

const (
	HighThinkBudget   ThinkBudget = "thinking_budget.high"
	MediumThinkBudget ThinkBudget = "thinking_budget.medium"
	LowThinkBudget    ThinkBudget = "thinking_budget.low"
)

type ThinkingLevel string

const (
	HighThinkingLevel ThinkingLevel = "high"
	LowThinkingLevel  ThinkingLevel = "low"
)

type MediaResolution string

const (
	HighMediaResolution   MediaResolution = "high"
	MediumMediaResolution MediaResolution = "medium"
	LowMediaResolution    MediaResolution = "low"
)

type GoogleTool []map[string]map[string]any

var GoogleSearchTool = map[string]map[string]any{
	"google_search": {},
}

type DynamicRetrievalConf struct {
	Mode             string `json:"mode"`
	DynamicThreshold int64  `json:"dynamic_threshold"`
}

var GoogleSearchRetrievalTool = map[string]map[string]any{
	"google_search_retrieval": {
		"dynamic_retrieval_config": DynamicRetrievalConf{
			Mode:             "MODE_DYNAMIC",
			DynamicThreshold: 1,
		},
	},
}

type GoogleImagePayload struct {
	MimeType string
	// Data can be either a base64 encoded payload or a file_uri.
	// If you pass a base64 encoded image you must omit the `data:image/<type>;base64,` part
	Data string
}

type GoogleFilePayload struct {
	MimeType string
	// Data can be either a base64 encoded payload or a file_uri.
	// If you pass a base64 encoded file you must omit the `data:<mimetype>;base64,` part
	Data string
}

type (
	// GooglePdf represents a PDF input, either as a URI or base64 data
	// The string can be either:
	// - A file URI (starts with "https://")
	// - Base64 encoded PDF data (with or without the data:application/pdf;base64, prefix)
	GooglePdf string
)

// NOTE: Gemini15Pro and Gemini15Flash types have been removed as these models were retired by Google in 2025

type Gemini20Flash struct {
	Tools GoogleTool
	// StructuredOutput represents a subset of the OpenAPI 3.0 Schema Object. Refer to gemini documentation for complete and up-to-date information. An example structure could be:
	//
	// 	var schemaGoogle = map[string]any{
	// 		"type": "object",
	// 		"properties": map[string]any{
	// 			"final_answer": map[string]any{"type": "string"},
	// 			"valuation": map[string]any{
	// 				"type": "number",
	// 			},
	// 		},
	// 	}
	StructuredOutput map[string]any
	// PdfFiles accepts one or more PDFs, either URIs or base64 data
	PdfFiles  []GooglePdf
	ImageFile []GoogleImagePayload
	// Files accepts any file type with URI and mime type
	Files    []GoogleFilePayload
	Thinking ThinkBudget
}

func (g Gemini20Flash) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.0000001
}

func (g Gemini20Flash) GetName() string {
	return Gemini20FlashModel
}

func (g Gemini20Flash) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini20Flash)

type Gemini20FlashLite struct {
	Tools GoogleTool
	// StructuredOutput represents a subset of the OpenAPI 3.0 Schema Object. Refer to gemini documentation for complete and up-to-date information. An example structure could be:
	//
	// 	var schemaGoogle = map[string]any{
	// 		"type": "object",
	// 		"properties": map[string]any{
	// 			"final_answer": map[string]any{"type": "string"},
	// 			"valuation": map[string]any{
	// 				"type": "number",
	// 			},
	// 		},
	// 	}
	StructuredOutput map[string]any
	// PdfFiles accepts one or more PDFs, either URIs or base64 data
	PdfFiles  []GooglePdf
	ImageFile []GoogleImagePayload
	// Files accepts any file type with URI and mime type
	Files    []GoogleFilePayload
	Thinking ThinkBudget
}

func (g Gemini20FlashLite) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000000075
}

func (g Gemini20FlashLite) GetName() string {
	return Gemini20FlashLiteModel
}

func (g Gemini20FlashLite) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini20FlashLite)

type Gemini25FlashPreview struct {
	Tools GoogleTool
	// StructuredOutput represents a subset of the OpenAPI 3.0 Schema Object. Refer to gemini documentation for complete and up-to-date information. An example structure could be:
	//
	// 	var schemaGoogle = map[string]any{
	// 		"type": "object",
	// 		"properties": map[string]any{
	// 			"final_answer": map[string]any{"type": "string"},
	// 			"valuation": map[string]any{
	// 				"type": "number",
	// 			},
	// 		},
	// 	}
	StructuredOutput map[string]any
	// PdfFiles accepts one or more PDFs, either URIs or base64 data
	PdfFiles  []GooglePdf
	ImageFile []GoogleImagePayload
	// Files accepts any file type with URI and mime type
	Files    []GoogleFilePayload
	Thinking ThinkBudget
}

func (g Gemini25FlashPreview) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.0000001
}

func (g Gemini25FlashPreview) GetName() string {
	return Gemini25FlashModel
}

func (g Gemini25FlashPreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini25FlashPreview)

type Gemini25ProPreview struct {
	Tools GoogleTool
	// StructuredOutput represents a subset of the OpenAPI 3.0 Schema Object. Refer to gemini documentation for complete and up-to-date information. An example structure could be:
	//
	// 	var schemaGoogle = map[string]any{
	// 		"type": "object",
	// 		"properties": map[string]any{
	// 			"final_answer": map[string]any{"type": "string"},
	// 			"valuation": map[string]any{
	// 				"type": "number",
	// 			},
	// 		},
	// 	}
	StructuredOutput map[string]any
	// PdfFiles accepts one or more PDFs, either URIs or base64 data
	PdfFiles  []GooglePdf
	ImageFile []GoogleImagePayload
	// Files accepts any file type with URI and mime type
	Files    []GoogleFilePayload
	Thinking ThinkBudget
}

func (g Gemini25ProPreview) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000125
}

func (g Gemini25ProPreview) GetName() string {
	return Gemini25ProModel
}

func (g Gemini25ProPreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini25ProPreview)

// AspectRatio represents the supported aspect ratios for image generation
type AspectRatio string

const (
	AspectRatio1x1  AspectRatio = "1:1"
	AspectRatio3x4  AspectRatio = "3:4"
	AspectRatio4x3  AspectRatio = "4:3"
	AspectRatio9x16 AspectRatio = "9:16"
	AspectRatio16x9 AspectRatio = "16:9"
)

// PersonGeneration controls whether images can contain people
type PersonGeneration string

const (
	PersonGenerationDontAllow  PersonGeneration = "dont_allow"
	PersonGenerationAllowAdult PersonGeneration = "allow_adult"
	PersonGenerationAllowAll   PersonGeneration = "allow_all"
)

// Gemini25FlashImage represents the Gemini 2.5 Flash image generation model
// This model generates images conversationally within the chat interface
type Gemini25FlashImage struct {
	// NumberOfImages specifies how many images to generate (1-4)
	NumberOfImages int
	// AspectRatio specifies the image aspect ratio
	AspectRatio AspectRatio
	// ImageFile accepts input images for image-to-image generation or reference
	ImageFile []GoogleImagePayload
	// PdfFiles accepts one or more PDFs, either URIs or base64 data
	PdfFiles []GooglePdf
	// Files accepts any file type with URI and mime type
	Files []GoogleFilePayload
}

func (g Gemini25FlashImage) EstimateCost(text string) float64 {
	// Gemini 2.5 Flash Image: $30.00 per 1M output tokens
	// Each image = 1290 output tokens = $0.039 per image
	numImages := g.NumberOfImages
	if numImages == 0 {
		numImages = 1
	}
	return float64(numImages) * 0.039
}

func (g Gemini25FlashImage) GetInputCostPer1M() float64 {
	// Official pricing: https://ai.google.dev/gemini-api/docs/pricing
	return 0.30
}

func (g Gemini25FlashImage) GetOutputCostPer1M() float64 {
	// Official pricing: $0.039 per image = 1290 tokens = $30.23 per 1M tokens
	return 30.23
}

func (g Gemini25FlashImage) GetName() string {
	return Gemini25FlashImageModel
}

func (g Gemini25FlashImage) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini25FlashImage)
var _ CostBreakdown = new(Gemini25FlashImage)

type Gemini3ProPreview struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	ThinkingLevel    ThinkingLevel
	MediaResolution  MediaResolution
}

func (g Gemini3ProPreview) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000002
}

func (g Gemini3ProPreview) GetInputCostPer1M() float64 {
	return 2.0
}

func (g Gemini3ProPreview) GetOutputCostPer1M() float64 {
	return 12.0
}

func (g Gemini3ProPreview) GetName() string {
	return Gemini3ProModel
}

func (g Gemini3ProPreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini3ProPreview)
var _ CostBreakdown = new(Gemini3ProPreview)

type Gemini3ProImagePreview struct {
	NumberOfImages  int
	AspectRatio     AspectRatio
	ImageFile       []GoogleImagePayload
	PdfFiles        []GooglePdf
	Files           []GoogleFilePayload
	ThinkingLevel   ThinkingLevel
	MediaResolution MediaResolution
}

func (g Gemini3ProImagePreview) EstimateCost(text string) float64 {
	numImages := g.NumberOfImages
	if numImages == 0 {
		numImages = 1
	}
	return float64(numImages) * 0.134
}

func (g Gemini3ProImagePreview) GetInputCostPer1M() float64 {
	return 2.0
}

func (g Gemini3ProImagePreview) GetOutputCostPer1M() float64 {
	return 30.0
}

func (g Gemini3ProImagePreview) GetName() string {
	return Gemini3ProImageModel
}

func (g Gemini3ProImagePreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini3ProImagePreview)
var _ CostBreakdown = new(Gemini3ProImagePreview)

type Gemini3FlashPreview struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []GooglePdf
	ImageFile        []GoogleImagePayload
	Files            []GoogleFilePayload
	ThinkingLevel    ThinkingLevel
	MediaResolution  MediaResolution
}

func (g Gemini3FlashPreview) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.0000005
}

func (g Gemini3FlashPreview) GetInputCostPer1M() float64 {
	return 0.50
}

func (g Gemini3FlashPreview) GetOutputCostPer1M() float64 {
	return 3.0
}

func (g Gemini3FlashPreview) GetName() string {
	return Gemini3FlashModel
}

func (g Gemini3FlashPreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini3FlashPreview)
var _ CostBreakdown = new(Gemini3FlashPreview)
