package models

const PerplexityProvider = "perplexity"

type SonarReasoningPro struct{}

// GetName implements Model.
func (s *SonarReasoningPro) GetName() string {
	return "sonar-reasoning-pro"
}

// GetProvider implements Model.
func (s *SonarReasoningPro) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarReasoningPro)

type SonarReasoning struct{}

// GetName implements Model.
func (s *SonarReasoning) GetName() string {
	return "sonar-reasoning"
}

// GetProvider implements Model.
func (s *SonarReasoning) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarReasoning)

type SonarPro struct{}

// GetName implements Model.
func (s *SonarPro) GetName() string {
	return "sonar-pro"
}

// GetProvider implements Model.
func (s *SonarPro) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(SonarPro)

type Sonar struct{}

// GetName implements Model.
func (s *Sonar) GetName() string {
	return "sonar"
}

// GetProvider implements Model.
func (s *Sonar) GetProvider() string {
	return PerplexityProvider
}

var _ Model = new(Sonar)
