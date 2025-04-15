package models

const OpenaiProvider = "openai"

const (
	O3MiniAlias    = "o3-mini-2025-01-31"
	GPT4OAlias     = "gpt-4o-2024-11-20"
	GPT4OMiniAlias = "gpt-4o-mini-2024-07-18"
	O1Alias        = "o1-2024-12-17"
	O1MiniAlias    = "o1-mini-2024-09-12"
	O1PreviewAlias = "o1-preview-2024-09-12"
	GPT4Alias      = "gpt-4-0613"
	GPT4TurboAlias = "gpt-4-turbo"
	GPT41Alias     = "gpt-4.1-2025-04-14"
)

type OpenaiImagePayload struct {
	// Url can be ether that, an url or a base64 encoding of the image .
	// If using base64, it must follow this format: data:image/jpeg;base64,{base64_image}
	Url string
	// Detail determines the level detail to use when processing and understanding the image. Can be either: high, low or auto. If nothing is specified, it will default to auto.
	Detail string
}

type GPT41 struct {
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
	// PdfFile let's you include a PDF file in your request to the LLM.
	// The expected format:
	//
	// map["file-name.pdf"]"data:application/pdf;base64," + encodedString
	// Only provide a pdf file or an image file, not both.
	PdfFile map[string]string
	// ImageFile enables vision for the request
	ImageFile []OpenaiImagePayload
}

func (GPT41) GetName() string {
	return GPT41Alias
}

func (GPT41) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O3Mini)

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

func (o O3Mini) GetName() string {
	return O3MiniAlias
}

func (o O3Mini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O3Mini)

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

	// PdfFile let's you include a PDF file in your request to the LLM.
	// The expected format:
	//
	// map["file-name.pdf"]"data:application/pdf;base64," + encodedString
	// Only provide a pdf file or an image file, not both.
	PdfFile map[string]string
	// ImageFile enables vision for the request
	ImageFile []OpenaiImagePayload
}

func (o O1) GetName() string {
	return O1Alias
}

func (o O1) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O1)

type O1Mini struct{}

func (o O1Mini) GetName() string {
	return O1MiniAlias
}

func (o O1Mini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O1Mini)

type O1Preview struct{}

func (o O1Preview) GetName() string {
	return O1PreviewAlias
}

func (o O1Preview) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O1Preview)

type GPT4 struct{}

func (g GPT4) GetName() string {
	return GPT4Alias
}

func (g GPT4) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4)

type GPT4Turbo struct{}

func (g GPT4Turbo) GetName() string {
	return GPT4TurboAlias
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

	// PdfFile let's you include a PDF file in your request to the LLM.
	// The expected format:
	//
	// map["file-name.pdf"]"data:application/pdf;base64," + encodedString
	// Only provide a pdf file or an image file, not both.
	PdfFile map[string]string
	// ImageFile enables vision for the request
	ImageFile []OpenaiImagePayload
}

func (g GPT4O) GetName() string {
	return GPT4OAlias
}

func (g GPT4O) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4O)

type (
	GPT4OMini struct {
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

		// PdfFile let's you include a PDF file in your request to the LLM.
		// The expected format:
		//
		// map["file-name.pdf"]"data:application/pdf;base64," + encodedString
		// Only provide a pdf file or an image file, not both.
		PdfFile map[string]string
		// ImageFile enables vision for the request
		ImageFile []OpenaiImagePayload
	}
)

func (g GPT4OMini) GetName() string {
	return GPT4OMiniAlias
}

func (g GPT4OMini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4OMini)
