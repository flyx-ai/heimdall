package models

const googleProvider = "google"

type GoogleTool []map[string]map[string]any

type GoogleModel interface {
	Model
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

type Gemini15FlashThink struct{}

func (gm Gemini15FlashThink) GetProvider() string {
	return googleProvider
}

func (gm Gemini15FlashThink) GetName() string {
	return "gemini-1.5-flash-002"
}

var _ Model = new(Gemini15FlashThink)

type Gemini15Pro struct{}

// GetName implements Model.
func (g *Gemini15Pro) GetName() string {
	return "gemini-1.5-pro-002"
}

// GetProvider implements Model.
func (g *Gemini15Pro) GetProvider() string {
	return googleProvider
}

var _ Model = new(Gemini15Pro)

type Gemini10ProVision struct{}

// GetName implements Model.
func (g *Gemini10ProVision) GetName() string {
	return "gemini-1.0-pro-vision-001"
}

// GetProvider implements Model.
func (g *Gemini10ProVision) GetProvider() string {
	return googleProvider
}

var _ Model = new(Gemini10ProVision)

type Gemini10Pro struct{}

// GetName implements Model.
func (g *Gemini10Pro) GetName() string {
	return "gemini-1.0-pro-002"
}

// GetProvider implements Model.
func (g *Gemini10Pro) GetProvider() string {
	return googleProvider
}

var _ Model = new(Gemini10Pro)

type Gemini20Flash struct {
	Tools GoogleTool
}

// GetName implements GoogleModel.
func (g *Gemini20Flash) GetName() string {
	return "gemini-2.0-flash-001"
}

// GetProvider implements GoogleModel.
func (g *Gemini20Flash) GetProvider() string {
	return googleProvider
}

// GetTools implements GoogleModel.
func (g *Gemini20Flash) GetTools() GoogleTool {
	return g.Tools
}

var _ GoogleModel = new(Gemini20Flash)

type Gemini20FlashLite struct {
	Tools GoogleTool
}

// GetName implements GoogleModel.
func (g *Gemini20FlashLite) GetName() string {
	return "gemini-2.0-flash-lite-001"
}

// GetProvider implements GoogleModel.
func (g *Gemini20FlashLite) GetProvider() string {
	return googleProvider
}

// GetTools implements GoogleModel.
func (g *Gemini20FlashLite) GetTools() GoogleTool {
	return g.Tools
}

var _ GoogleModel = new(Gemini20FlashLite)
