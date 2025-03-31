package models

const OpenaiProvider = "openai"

type O3Mini struct{}

func (o O3Mini) GetName() string {
	return "o1-mini"
}

func (o O3Mini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O3Mini)

type O1 struct{}

func (o O1) GetName() string {
	return "o1"
}

func (o O1) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O1)

type O1Mini struct{}

func (o O1Mini) GetName() string {
	return "o1-mini"
}

func (o O1Mini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O1Mini)

type O1Preview struct{}

func (o O1Preview) GetName() string {
	return "o1-preview"
}

func (o O1Preview) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O1Preview)

type GPT4 struct{}

func (g GPT4) GetName() string {
	return "gpt-4"
}

func (g GPT4) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4)

type GPT4Turbo struct{}

func (g GPT4Turbo) GetName() string {
	return "gpt-4-turbo"
}

func (g GPT4Turbo) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4Turbo)

type GPT4O struct{}

func (g GPT4O) GetName() string {
	return "gpt-4o"
}

func (g GPT4O) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4O)

type GPT4OMini struct{}

func (g GPT4OMini) GetName() string {
	return "gpt-4o-mini"
}

func (g GPT4OMini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4OMini)
