package models

const GoogleProvider = "google"

const (
	Gemini15FlashModel = "gemini-1.5-flash-002"
	Gemini15ProModel   = "gemini-1.5-pro-002"
	// Gemini20FlashModel        = "gemini-2.0-flash-001" // Added GenerateImage field
	Gemini20FlashModel        = "gemini-2.0-flash-exp-image-generation" // Added GenerateImage field
	Gemini20FlashLiteModel    = "gemini-2.0-flash-lite-001"
	Gemini25FlashPreviewModel = "gemini-2.5-flash-preview-04-17"
	Gemini25ProPreviewModel   = "gemini-2.5-pro-preview-03-25"
)

type ThinkBudget string

const (
	HighThinkBudget   ThinkBudget = "thinking_budget.high"
	MediumThinkBudget ThinkBudget = "thinking_budget.medium"
	LowThinkBudget    ThinkBudget = "thinking_budget.low"
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

type (
	// GooglePdf represents a PDF input, either as a URI or base64 data
	// The string can be either:
	// - A file URI (starts with "https://")
	// - Base64 encoded PDF data (with or without the data:application/pdf;base64, prefix)
	GooglePdf string
)

type Gemini15Pro struct {
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
	Thinking  ThinkBudget
}

func (g Gemini15Pro) GetName() string {
	return Gemini15ProModel
}

func (g Gemini15Pro) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini15Pro)

type Gemini15Flash struct {
	Thinking ThinkBudget
}

func (g Gemini15Flash) GetName() string {
	return Gemini15FlashModel
}

func (g Gemini15Flash) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini15Flash)

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
	PdfFiles      []GooglePdf
	ImageFile     []GoogleImagePayload
	Thinking      ThinkBudget
	GenerateImage bool `json:"generate_image,omitempty"` // Flag to request image generation
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
	Thinking  ThinkBudget
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
	Thinking  ThinkBudget
}

func (g Gemini25FlashPreview) GetName() string {
	return Gemini25FlashPreviewModel
}

func (g Gemini25FlashPreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini25ProPreview)

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
	Thinking  ThinkBudget
}

func (g Gemini25ProPreview) GetName() string {
	return Gemini25ProPreviewModel
}

func (g Gemini25ProPreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini25ProPreview)
