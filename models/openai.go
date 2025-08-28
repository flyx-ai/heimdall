package models

const OpenaiProvider = "openai"

const (
	O3MiniAlias    = "o3-mini-2025-01-31"
	GPT4OAlias     = "gpt-4o-2024-11-20"
	GPT4OMiniAlias = "gpt-4o-mini-2024-07-18"
	O1Alias        = "o1-2024-12-17"
	GPT4Alias      = "gpt-4-0613"
	GPT4TurboAlias = "gpt-4-turbo"
	GPT41Alias     = "gpt-4.1-2025-04-14"
	GPT41MiniAlias = "gpt-4.1-mini-2025-04-14"
	GPT41NanoAlias = "gpt-4.1-nano-2025-04-14"
	GPT5Alias      = "gpt-5-2025-08-07"
	GPT5MiniAlias  = "gpt-5-mini-2025-08-07"
	GPT5NanoAlias  = "gpt-5-nano-2025-08-07"
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

func (g GPT41) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000200
}

func (GPT41) GetName() string {
	return GPT41Alias
}

func (GPT41) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT41)

type GPT41Mini struct {
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

func (g GPT41Mini) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000040
}

func (GPT41Mini) GetName() string {
	return GPT41MiniAlias
}

func (GPT41Mini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT41Mini)

type GPT41Nano struct {
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

func (g GPT41Nano) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000010
}

func (GPT41Nano) GetName() string {
	return GPT41NanoAlias
}

func (GPT41Nano) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT41Nano)

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

func (o O3Mini) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000110
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

func (o O1) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00001500
}

func (o O1) GetName() string {
	return O1Alias
}

func (o O1) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(O1)

type GPT4 struct{}

func (g GPT4) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00006000
}

func (g GPT4) GetName() string {
	return GPT4Alias
}

func (g GPT4) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4)

type GPT4Turbo struct{}

func (g GPT4Turbo) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00001000
}

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

func (g GPT4O) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000250
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

func (g GPT4OMini) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000015
}

func (g GPT4OMini) GetName() string {
	return GPT4OMiniAlias
}

func (g GPT4OMini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT4OMini)

type GPT5 struct{}

func (g GPT5) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000125
}

func (g GPT5) GetName() string {
	return GPT5Alias
}

func (g GPT5) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT5)

type GPT5Mini struct{}

func (g GPT5Mini) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 0.00000025
}

func (g GPT5Mini) GetName() string {
	return GPT5MiniAlias
}

func (g GPT5Mini) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT5Mini)

type GPT5Nano struct{}

func (g GPT5Nano) EstimateCost(text string) float64 {
	return (float64(len(text)) / 4) * 5e-8
}

func (g GPT5Nano) GetName() string {
	return GPT5NanoAlias
}

func (g GPT5Nano) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPT5Nano)

const ImageModelAlias = "gpt-image-1"

const (
	GPTImageSize1024x1024 = "1024x1024"
	GPTImageSize1792x1024 = "1792x1024"
	GPTImageSize1024x1792 = "1024x1792"

	GPTImageQualityHigh   = "high"
	GPTImageQualityMedium = "medium"
	GPTImageQualityLow    = "low"
)

type GPTImage struct {
	// Allows to set transparency for the background of the generated image(s).
	// Must be one of transparent, opaque or auto (default value).
	// When auto is used, the model will automatically determine the best background for the image.
	// If transparent, the output format needs to support transparency, so it should be set to either png (default value) or webp.
	Background string

	// N is the number of images to generate. Must be 1 for DALL·E 3.
	// Although the API docs mention 'n', DALL-E 3 currently only supports n=1.
	// We keep it for potential future compatibility but default/validate to 1.
	N int

	// Size of the generated images. Defaults to "1024x1024".
	Size string

	// Quality of the image that will be generated. Defaults to "auto".
	Quality string

	// The compression level (0-100%) for the generated images.
	// This parameter is only supported the webp or jpeg output formats, and defaults to 100.
	OutputCompression string

	// The format in which the generated images are returned.
	// Must be one of png, jpeg, or webp.
	OutputFormat string

	// Must be either low for less restrictive filtering or auto (default value).
	Moderation string

	// User is an optional unique identifier representing your end-user,
	// which can help OpenAI monitor and detect abuse.
	User string
}

// TODO
func (d GPTImage) EstimateCost(text string) float64 {
	return 0.0
}

func (d GPTImage) GetName() string {
	return ImageModelAlias
}

func (d GPTImage) GetProvider() string {
	return OpenaiProvider
}

var _ Model = new(GPTImage)
