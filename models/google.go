package models

const GoogleProvider = "google"

const (
	Gemini15ProModel         = "gemini-1.5-pro-002"
	Gemini15FlashModel       = "gemini-1.5-flash-002"
	Gemini20FlashModel       = "gemini-2.0-flash-001"
	Gemini20FlashLiteModel   = "gemini-2.0-flash-lite-001"
	Gemini25ProPreviewpModel = "gemini-2.5-pro-preview-03-25"
)

type GoogleTool []map[string]map[string]any

type GoogleTools interface {
	GetTools() GoogleTool
}

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
	PdfFile          map[string]string
	StructuredOutput map[string]any
}

func (g Gemini15Pro) GetName() string {
	return Gemini15ProModel
}

func (g Gemini15Pro) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini15Pro)

type Gemini15Flash struct{}

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
	PdfFile          map[string]string
	StructuredOutput map[string]any
}

func (g Gemini20Flash) GetName() string {
	return Gemini20FlashModel
}

func (g Gemini20Flash) GetProvider() string {
	return GoogleProvider
}

func (g Gemini20Flash) GetTools() GoogleTool {
	return g.Tools
}

var (
	_ Model       = new(Gemini20Flash)
	_ GoogleTools = new(Gemini20Flash)
)

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
	PdfFile          map[string]string
	StructuredOutput map[string]any
}

func (g Gemini20FlashLite) GetName() string {
	return Gemini20FlashLiteModel
}

func (g Gemini20FlashLite) GetProvider() string {
	return GoogleProvider
}

func (g Gemini20FlashLite) GetTools() GoogleTool {
	return g.Tools
}

var (
	_ Model       = new(Gemini20FlashLite)
	_ GoogleTools = new(Gemini20FlashLite)
)

type Gemini25ProPreview struct{}

func (g Gemini25ProPreview) GetName() string {
	return Gemini25ProPreviewpModel
}

func (g Gemini25ProPreview) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini25ProPreview)
