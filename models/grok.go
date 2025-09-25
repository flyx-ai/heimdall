package models

const GrokProvider = "grok"

const (
	Grok2VisionAlias   = "grok-2-vision"
	Grok3Alias         = "grok-3"
	Grok3MiniAlias     = "grok-3-mini"
	Grok3FastAlias     = "grok-3-fast"
	Grok3MiniFastAlias = "grok-3-mini-fast"
	Grok4Alias         = "grok-4"
	Grok4FastAlias     = "grok-4-fast"
)

type GrokImagePayload struct {
	URL    string
	Detail string
}

type Grok2Vision struct {
	ImageFile []GrokImagePayload
}

func (g Grok2Vision) EstimateCost(text string) float64 {
	inputCostPerToken := 0.000002
	outputCostPerToken := 0.000002
	averageCost := (inputCostPerToken + outputCostPerToken) / 2
	return (float64(len(text)) / 4) * averageCost
}

func (Grok2Vision) GetName() string {
	return Grok2VisionAlias
}

func (Grok2Vision) GetProvider() string {
	return GrokProvider
}

var _ Model = new(Grok2Vision)

type Grok3 struct{}

func (g Grok3) EstimateCost(text string) float64 {
	inputCostPerToken := 0.000003
	outputCostPerToken := 0.000015
	averageCost := (inputCostPerToken + outputCostPerToken) / 2
	return (float64(len(text)) / 4) * averageCost
}

func (Grok3) GetName() string {
	return Grok3Alias
}

func (Grok3) GetProvider() string {
	return GrokProvider
}

var _ Model = new(Grok3)

type Grok3Mini struct{}

func (g Grok3Mini) EstimateCost(text string) float64 {
	inputCostPerToken := 0.0000003
	outputCostPerToken := 0.0000005
	averageCost := (inputCostPerToken + outputCostPerToken) / 2
	return (float64(len(text)) / 4) * averageCost
}

func (Grok3Mini) GetName() string {
	return Grok3MiniAlias
}

func (Grok3Mini) GetProvider() string {
	return GrokProvider
}

var _ Model = new(Grok3Mini)

type Grok3Fast struct{}

func (g Grok3Fast) EstimateCost(text string) float64 {
	inputCostPerToken := 0.000005
	outputCostPerToken := 0.000025
	averageCost := (inputCostPerToken + outputCostPerToken) / 2
	return (float64(len(text)) / 4) * averageCost
}

func (Grok3Fast) GetName() string {
	return Grok3FastAlias
}

func (Grok3Fast) GetProvider() string {
	return GrokProvider
}

var _ Model = new(Grok3Fast)

type Grok3MiniFast struct{}

func (g Grok3MiniFast) EstimateCost(text string) float64 {
	inputCostPerToken := 0.0000006
	outputCostPerToken := 0.000004
	averageCost := (inputCostPerToken + outputCostPerToken) / 2
	return (float64(len(text)) / 4) * averageCost
}

func (Grok3MiniFast) GetName() string {
	return Grok3MiniFastAlias
}

func (Grok3MiniFast) GetProvider() string {
	return GrokProvider
}

var _ Model = new(Grok3MiniFast)

type Grok4 struct{}

func (g Grok4) EstimateCost(text string) float64 {
	inputCostPerToken := 0.000002
	outputCostPerToken := 0.000002
	averageCost := (inputCostPerToken + outputCostPerToken) / 2
	return (float64(len(text)) / 4) * averageCost
}

func (Grok4) GetName() string {
	return Grok4Alias
}

func (Grok4) GetProvider() string {
	return GrokProvider
}

var _ Model = new(Grok4)

type Grok4Fast struct{}

func (g Grok4Fast) EstimateCost(text string) float64 {
	inputCostPerToken := 0.0000002
	outputCostPerToken := 0.0000005
	averageCost := (inputCostPerToken + outputCostPerToken) / 2
	return (float64(len(text)) / 4) * averageCost
}

func (Grok4Fast) GetName() string {
	return Grok4FastAlias
}

func (Grok4Fast) GetProvider() string {
	return GrokProvider
}

var _ Model = new(Grok4Fast)
