package models

const PerplexityProvider = "perplexity"

type SonarReasoningPro struct{}

func (s SonarReasoningPro) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000002
}

func (s SonarReasoningPro) GetName() string {
	return "sonar-reasoning-pro"
}

func (s SonarReasoningPro) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarReasoningPro)

type SonarReasoning struct{}

func (s SonarReasoning) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000001
}

func (s SonarReasoning) GetName() string {
	return "sonar-reasoning"
}

func (s SonarReasoning) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarReasoning)

type SonarPro struct{}

func (s SonarPro) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000003
}

func (s SonarPro) GetName() string {
	return "sonar-pro"
}

func (s SonarPro) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarPro)

type Sonar struct{}

func (s Sonar) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.000001
}

func (s Sonar) GetName() string {
	return "sonar"
}

func (s Sonar) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(Sonar)
