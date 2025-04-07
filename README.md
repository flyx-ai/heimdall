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
