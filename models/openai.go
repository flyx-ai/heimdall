package models

const OpenaiProvider = "openai"

type O3Mini struct {
	// StructuredOutput represents a subset of the JSON Schema Language. Refer to openai documentation for complete and up-to-date information. An example structure could be:
	//
	//  var schema = map[string]any{
	//  	"name": "navidia_valuation",
	//  	"schema": map[string]any{
	//  		"type": "object",
	//  		"properties": map[string]any{
	//  			"final_answer": map[string]any{"type": "string"},
	//  			"valuation": map[string]any{
	//  				"type": "number",
	//  			},
	//  		},
	//  	},
	//  }
	StructuredOutput map[string]any
}

func (o O3Mini) GetStructuredOutput() map[string]any {
	return o.StructuredOutput
}

func (o O3Mini) GetName() string {
	return "o3-mini"
}

func (o O3Mini) GetProvider() string {
	return OpenaiProvider
}

var (
	_ Model            = new(O3Mini)
	_ StructuredOutput = new(O3Mini)
)

type O1 struct {
	// StructuredOutput represents a subset of the JSON Schema Language. Refer to openai documentation for complete and up-to-date information. An example structure could be:
	//
	//  var schema = map[string]any{
	//  	"name": "navidia_valuation",
	//  	"schema": map[string]any{
	//  		"type": "object",
	//  		"properties": map[string]any{
	//  			"final_answer": map[string]any{"type": "string"},
	//  			"valuation": map[string]any{
	//  				"type": "number",
	//  			},
	//  		},
	//  	},
	//  }
	StructuredOutput map[string]any
}

func (o O1) GetStructuredOutput() map[string]any {
	return o.StructuredOutput
}

func (o O1) GetName() string {
	return "o1"
}

func (o O1) GetProvider() string {
	return OpenaiProvider
}

var (
	_ Model            = new(O1)
	_ StructuredOutput = new(O1)
)

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

type GPT4O struct {
	// StructuredOutput represents a subset of the JSON Schema Language. Refer to openai documentation for complete and up-to-date information. An example structure could be:
	//
	//  var schema = map[string]any{
	//  	"name": "navidia_valuation",
	//  	"schema": map[string]any{
	//  		"type": "object",
	//  		"properties": map[string]any{
	//  			"final_answer": map[string]any{"type": "string"},
	//  			"valuation": map[string]any{
	//  				"type": "number",
	//  			},
	//  		},
	//  	},
	//  }
	StructuredOutput map[string]any
}

func (g GPT4O) GetStructuredOutput() map[string]any {
	return g.StructuredOutput
}

func (g GPT4O) GetName() string {
	return "gpt-4o"
}

func (g GPT4O) GetProvider() string {
	return OpenaiProvider
}

var (
	_ Model            = new(GPT4O)
	_ StructuredOutput = new(GPT4O)
)

type GPT4OMini struct {
	// StructuredOutput represents a subset of the JSON Schema Language. Refer to openai documentation for complete and up-to-date information. An example structure could be:
	//
	//  var schema = map[string]any{
	//  	"name": "navidia_valuation",
	//  	"schema": map[string]any{
	//  		"type": "object",
	//  		"properties": map[string]any{
	//  			"final_answer": map[string]any{"type": "string"},
	//  			"valuation": map[string]any{
	//  				"type": "number",
	//  			},
	//  		},
	//  	},
	//  }
	StructuredOutput map[string]any
}

func (g GPT4OMini) GetStructuredOutput() map[string]any {
	return g.StructuredOutput
}

func (g GPT4OMini) GetName() string {
	return "gpt-4o-mini"
}

func (g GPT4OMini) GetProvider() string {
	return OpenaiProvider
}

var (
	_ Model            = new(GPT4OMini)
	_ StructuredOutput = new(GPT4OMini)
)
