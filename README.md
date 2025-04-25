# Heimdall

## Purpose

Heimdall is a router that aims to make LLMs request more consistent; providing you with automatic retries and fallback option.

## Examples

Below is a collection of examples on how to use heimdall.

### File input (openai)

```go
func main() {
	ctx := context.Background()

	oaApiKey := os.Getenv("OPENAI_API_KEY")

	fileBytes, err := os.ReadFile("path-to-file")
	if err != nil {
		panic(err)
	}

	encodedString := base64.StdEncoding.EncodeToString(fileBytes)

	dataURL := "data:application/pdf;base64," + encodedString

	oa := providers.NewOpenAI([]string{oaApiKey})
	res, err := oa.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.GPT4O{
				PdfFile: map[string]string{
					"file-name.png": dataURL,
				},
			},
			Messages: []request.Message{
				{
					Role:    "user",
					Content: "describe the content of the file",
				},
			},
		},
	)
}
```

### File input + structured output (openai)

```go
var schema = map[string]any{
	"name": "picture_description",
	"schema": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary":   map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
		},
	},
}

func main() {
	ctx := context.Background()

	oaApiKey := os.Getenv("OPENAI_API_KEY")

	fileBytes, err := os.ReadFile("path-to-file")
	if err != nil {
		panic(err)
	}

	encodedString := base64.StdEncoding.EncodeToString(fileBytes)

	dataURL := "data:application/pdf;base64," + encodedString

	oa := providers.NewOpenAI([]string{oaApiKey})
	res, err := oa.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.GPT4O{
				StructuredOutput: schema,
				PdfFile: map[string]string{
					"file-name.png": dataURL,
				},
			},
			Messages: []request.Message{
				{
					Role:    "user",
					Content: "describe the content of the file",
				},
			},
		},
	)
}
```

### Structured output (openai)

```go
var schema = map[string]any{
	"name": "picture_description",
	"schema": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary":   map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
		},
	},
}

func main() {
	ctx := context.Background()

	oaApiKey := os.Getenv("OPENAI_API_KEY")

	oa := providers.NewOpenAI([]string{oaApiKey})
	res, err := oa.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.GPT4O{
				StructuredOutput: schema,
			},
			Messages: []request.Message{
				{
					Role:    "user",
					Content: "create a thorough financial analysis of Nvidia at its current valuation",
				},
			},
		},
	)
}
```

### PDF Input (Google/Gemini)

```go
func main() {
	ctx := context.Background()

	gApiKey := os.Getenv("GOOGLE_API_KEY")

	// Read and encode a PDF file
	fileBytes, err := os.ReadFile("path-to-pdf-file.pdf")
	if err != nil {
		panic(err)
	}
	encodedPdf := base64.StdEncoding.EncodeToString(fileBytes)

	// Create a Google provider
	g := providers.NewGoogle([]string{gApiKey})
	
	// Example with base64 encoded PDF
	res, err := g.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Gemini15Pro{
				PdfFiles: []models.GooglePdf{
					models.GooglePdf(encodedPdf), // Base64 encoded PDF
				},
			},
			SystemMessage: "You are a helpful assistant that analyzes documents.",
			UserMessage:   "Summarize the main points from this PDF document",
			Temperature:   0,
			TopP:          0,
			Tags:          map[string]string{},
		},
		http.Client{},
		nil,
	)
	
	// Example with PDF URL
	res, err = g.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Gemini15Pro{
				PdfFiles: []models.GooglePdf{
					models.GooglePdf("https://example.com/sample.pdf"), // PDF URL
				},
			},
			SystemMessage: "You are a helpful assistant that analyzes documents.",
			UserMessage:   "Summarize the main points from this PDF document",
			Temperature:   0,
			TopP:          0,
			Tags:          map[string]string{},
		},
		http.Client{},
		nil,
	)
}
```

### Multiple PDF Input (Google/Gemini)

```go
func main() {
	ctx := context.Background()

	gApiKey := os.Getenv("GOOGLE_API_KEY")

	// Read and encode a PDF file
	pdf1Bytes, _ := os.ReadFile("first-document.pdf")
	encodedPdf1 := base64.StdEncoding.EncodeToString(pdf1Bytes)
	
	// Create a Google provider
	g := providers.NewGoogle([]string{gApiKey})
	
	// Example with multiple PDFs (mix of base64 and URLs)
	res, err := g.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Gemini15Pro{
				PdfFiles: []models.GooglePdf{
					models.GooglePdf(encodedPdf1), // Base64 encoded PDF
					models.GooglePdf("https://example.com/second-document.pdf"), // PDF URL
				},
			},
			SystemMessage: "You are a helpful assistant that analyzes documents.",
			UserMessage:   "Compare the content of these two PDF documents",
			Temperature:   0,
			TopP:          0,
			Tags:          map[string]string{},
		},
		http.Client{},
		nil,
	)
}
```
