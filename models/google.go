package models

const GoogleProvider = "google"

type GoogleModelName string

const (
	Gemini15ProModel       GoogleModelName = "gemini-1.5-pro-002"
	Gemini15FlashModel     GoogleModelName = "gemini-1.5-flash-002"
	Gemini20FlashModel     GoogleModelName = "gemini-2.0-flash-001"
	Gemini20FlashLiteModel GoogleModelName = "gemini-2.0-flash-lite-001"
	Gemini25ProExpModel    GoogleModelName = "gemini-2.5-pro-exp-03-25"
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
	StructuredOutput map[string]any
}

func (g Gemini15Pro) GetStructuredOutput() map[string]any {
	return g.StructuredOutput
}

func (g Gemini15Pro) GetName() string {
	return string(Gemini15ProModel)
}

func (g Gemini15Pro) GetProvider() string {
	return GoogleProvider
}

var (
	_ Model            = new(Gemini15Pro)
	_ StructuredOutput = new(Gemini15Pro)
)

type Gemini15Flash struct{}

func (g Gemini15Flash) GetName() string {
	return string(Gemini15FlashModel)
}

func (g Gemini15Flash) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini15Flash)

type Gemini20Flash struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
}

func (g Gemini20Flash) GetStructuredOutput() map[string]any {
	return g.StructuredOutput
}

func (g Gemini20Flash) GetName() string {
	return string(Gemini20FlashModel)
}

func (g Gemini20Flash) GetProvider() string {
	return GoogleProvider
}

func (g Gemini20Flash) GetTools() GoogleTool {
	return g.Tools
}

var (
	_ Model            = new(Gemini20Flash)
	_ StructuredOutput = new(Gemini20Flash)
	_ GoogleTools      = new(Gemini20Flash)
)

type Gemini20FlashLite struct {
	Tools            GoogleTool
	StructuredOutput map[string]any
}

func (g Gemini20FlashLite) GetStructuredOutput() map[string]any {
	return g.StructuredOutput
}

func (g Gemini20FlashLite) GetName() string {
	return string(Gemini20FlashLiteModel)
}

func (g Gemini20FlashLite) GetProvider() string {
	return GoogleProvider
}

func (g Gemini20FlashLite) GetTools() GoogleTool {
	return g.Tools
}

var (
	_ Model            = new(Gemini20FlashLite)
	_ StructuredOutput = new(Gemini20FlashLite)
	_ GoogleTools      = new(Gemini20FlashLite)
)

type Gemini25ProExp struct{}

func (g Gemini25ProExp) GetName() string {
	return string(Gemini25ProExpModel)
}

func (g Gemini25ProExp) GetProvider() string {
	return GoogleProvider
}

var _ Model = new(Gemini25ProExp)
