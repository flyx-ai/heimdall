package models

const OpenRouterProvider = "openrouter"

type OpenRouterImagePayload struct {
	Url    string
	Detail string
}

type OpenRouterModel struct {
	ModelName        string
	ImageFile        []OpenRouterImagePayload
	PdfFile          map[string]string
	StructuredOutput map[string]any
}

func (o OpenRouterModel) EstimateCost(text string) float64 {
	return 0
}

func (o OpenRouterModel) GetName() string {
	return o.ModelName
}

func (o OpenRouterModel) GetProvider() string {
	return OpenRouterProvider
}

var _ Model = new(OpenRouterModel)
