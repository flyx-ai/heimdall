package models

const PerplexityProvider = "perplexity"

type SonarReasoningPro struct{}

func (s SonarReasoningPro) GetName() string {
	return "sonar-reasoning-pro"
}

func (s SonarReasoningPro) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarReasoningPro)

type SonarReasoning struct{}

func (s SonarReasoning) GetName() string {
	return "sonar-reasoning"
}

func (s SonarReasoning) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarReasoning)

type SonarPro struct{}

func (s SonarPro) GetName() string {
	return "sonar-pro"
}

func (s SonarPro) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarPro)

type Sonar struct{}

func (s Sonar) GetName() string {
	return "sonar"
}

func (s Sonar) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(Sonar)
