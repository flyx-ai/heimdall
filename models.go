package heimdall

type Model struct {
	Provider Provider
	Name     string
}

var ModelO3Mini Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1-mini",
}

var ModelO1 Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1",
}

var ModelO1Mini Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1-mini",
}

var ModelO1Preview Model = Model{
	Provider: ProviderOpenAI,
	Name:     "o1-preview",
}

var ModelGPT4 Model = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4",
}

var ModelGPT4Turbo Model = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4-turbo",
}

var ModelGPT4OModel = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4o",
}

var ModelGPT4OMini Model = Model{
	Provider: ProviderOpenAI,
	Name:     "gpt-4o-mini",
}
