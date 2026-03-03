# Heimdall

Heimdall is a robust Go library for making LLM (Large Language Model) requests more consistent by providing automatic retries and fallback options. It acts as a router between your application and various LLM providers, ensuring reliable and efficient interaction with AI models.

## Features

- **Provider Abstraction**: Unified interface for multiple LLM providers (OpenAI, Anthropic, Google/Gemini, Grok, VertexAI, OpenRouter, Perplexity)
- **Request Retries**: Automatic retry mechanism for handling transient failures
- **Model Fallbacks**: Configurable fallback models if primary model fails
- **Streaming Support**: Fully supports streaming responses for real-time applications
- **Multimodal Inputs**: Support for PDFs, images, and other file types
- **Structured Output**: Format responses as JSON using provider-specific schema formats
- **Request Logging**: Built-in request logging for monitoring and debugging

## Installation

```bash
go get github.com/flyx-ai/heimdall
```

## Quick Start

### Basic Usage

Here's a simple example of how to use Heimdall with OpenAI:

```go
package main

import (
	"context"
	"os"
	"time"

	"github.com/flyx-ai/heimdall"
	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/providers"
	"github.com/flyx-ai/heimdall/request"
)

func main() {
	ctx := context.Background()
	
	// Create a provider with your API key
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	openAIProvider := providers.NewOpenAI([]string{openAIAPIKey})
	
	// Setup the router with a timeout and the provider
	timeout := 30 * time.Second
	router := heimdall.New(timeout, []heimdall.LLMProvider{openAIProvider})
	
	// Create a completion request
	req := request.Completion{
		Model: models.GPT4O{},
		SystemMessage: "You are a helpful assistant.",
		UserMessage:   "What's the capital of France?",
		Fallback: []models.Model{
			models.GPT4OMini{},
		},
		Temperature: 0.7,
		TopP:        1.0,
		Tags: map[string]string{
			"env":  "production",
			"type": "geography",
		},
	}
	
	// Get a completion
	response, err := router.Complete(ctx, req)
	if err != nil {
		panic(err)
	}
	
	// Use the response
	println(response.Content)
}
```

## Provider Setup

Heimdall supports multiple LLM providers. Here's how to set them up:

### OpenAI

```go
// Single API key
openAIProvider := providers.NewOpenAI([]string{"your-api-key"})

// Multiple API keys for load balancing
openAIProvider := providers.NewOpenAI([]string{"key1", "key2", "key3"})
```

### Anthropic

```go
anthropicProvider := providers.NewAnthropic([]string{"your-api-key"})
```

### Google/Gemini

```go
googleProvider := providers.NewGoogle([]string{"your-api-key"})
```

Heimdall also supports caching tokens for Google provider to reduce latency and improve performance for repeated requests. 

```go
func main() {
	ctx := context.Background()

	googleApiKey := os.Getenv("GOOGLE_API_KEY")
	g := providers.NewGoogle([]string{googleApiKey})
    
	key, err := g.CacheContent(
		ctx,
		models.Gemini20Flash{}.GetName(),
		providers.CacheContentPayload{
			FileData: map[string]string{
                // Add your file data here
				<mime_type_here>: <file_uri_here>,
			},
		},
		<system_prompt>,
		10*time.Minute,
	)
}

```

### Perplexity

```go
perplexityProvider := providers.NewPerplexity([]string{"your-api-key"})
```

### VertexAI

```go
vertexAIProvider := providers.NewVertexAI([]string{"your-api-key"})
```

### Grok

```go
grokProvider := providers.NewGrok([]string{"your-api-key"})
```

### OpenRouter

```go
openRouterProvider := providers.NewOpenRouter([]string{"your-api-key"})
```

## Working with PDF Files

### OpenAI with PDF Input

```go
func main() {
	ctx := context.Background()
	
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	
	// Read and encode the PDF file
	fileBytes, err := os.ReadFile("document.pdf")
	if err != nil {
		panic(err)
	}
	
	encodedString := base64.StdEncoding.EncodeToString(fileBytes)
	dataURL := "data:application/pdf;base64," + encodedString
	
	openAIProvider := providers.NewOpenAI([]string{openAIAPIKey})
	res, err := openAIProvider.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.GPT4O{
				PdfFile: map[string]string{
					"document.pdf": dataURL,
				},
			},
			UserMessage: "Summarize the contents of this PDF",
		},
		http.Client{},
		nil,
	)
	
	if err != nil {
		panic(err)
	}
	
	fmt.Println(res.Content)
}
```

### Google/Gemini with PDF Input

```go
func main() {
	ctx := context.Background()
	
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	
	// Read and encode a PDF file
	fileBytes, err := os.ReadFile("document.pdf")
	if err != nil {
		panic(err)
	}
	encodedPdf := base64.StdEncoding.EncodeToString(fileBytes)
	
	// Create a Google provider
	googleProvider := providers.NewGoogle([]string{googleAPIKey})
	
	// Example with base64 encoded PDF
	res, err := googleProvider.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Gemini20Flash{
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
	
	if err != nil {
		panic(err)
	}
	
	fmt.Println(res.Content)
}
```

### Anthropic with PDF Input

```go
func main() {
	ctx := context.Background()
	
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	
	// Read and encode a PDF file
	fileBytes, err := os.ReadFile("document.pdf")
	if err != nil {
		panic(err)
	}
	encodedPdf := base64.StdEncoding.EncodeToString(fileBytes)
	
	// Create an Anthropic provider
	anthropicProvider := providers.NewAnthropic([]string{anthropicAPIKey})
	
	// Request with PDF
	res, err := anthropicProvider.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Claude46Sonnet{
				PdfFiles: []models.AnthropicPdf{
					models.AnthropicPdf(encodedPdf),
				},
			},
			SystemMessage: "You are a helpful document analyst.",
			UserMessage:   "Analyze this PDF and provide key insights",
			Temperature:   0,
			TopP:          0,
			Tags:          map[string]string{},
		},
		http.Client{},
		nil,
	)
	
	if err != nil {
		panic(err)
	}
	
	fmt.Println(res.Content)
}
```

## Streaming Responses

For streaming responses, use the `Stream` method:

```go
func main() {
	ctx := context.Background()
	
	// Create provider and router
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	openAIProvider := providers.NewOpenAI([]string{openAIAPIKey})
	
	timeout := 30 * time.Second
	router := heimdall.New(timeout, []heimdall.LLMProvider{openAIProvider})
	
	// Create request
	req := request.Completion{
		Model: models.GPT4O{},
		SystemMessage: "You are a helpful assistant.",
		UserMessage:   "Write a short story about a space explorer",
		Temperature: 0.7,
		Tags: map[string]string{
			"env":  "production",
			"type": "creative",
		},
	}
	
	// Handle streaming chunks
	chunkHandler := func(chunk string) error {
		fmt.Print(chunk)
		return nil
	}
	
	// Get streaming response
	_, err := router.Stream(ctx, req, chunkHandler)
	if err != nil {
		panic(err)
	}
}
```

## Structured Output

You can request structured output from supported models:

### OpenAI Structured Output

```go
var schema = map[string]any{
	"name": "financial_analysis",
	"schema": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary":     map[string]any{"type": "string"},
			"marketCap":   map[string]any{"type": "number"},
			"advantages":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"risks":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"recommendation": map[string]any{"type": "string", "enum": []string{"buy", "hold", "sell"}},
		},
		"required": []string{"summary", "recommendation"},
	},
}

func main() {
	ctx := context.Background()
	
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	
	openAIProvider := providers.NewOpenAI([]string{openAIAPIKey})
	res, err := openAIProvider.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.GPT4O{
				StructuredOutput: schema,
			},
			UserMessage: "Create a financial analysis of Nvidia at its current valuation",
		},
		http.Client{},
		nil,
	)
	
	if err != nil {
		panic(err)
	}
	
	fmt.Println(res.Content)
}
```

### Google/Gemini Structured Output

```go
var schemaGoogle = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary":     map[string]any{"type": "string"},
		"marketCap":   map[string]any{"type": "number"},
		"advantages":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"risks":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"recommendation": map[string]any{"type": "string", "enum": []string{"buy", "hold", "sell"}},
	},
	"required": []string{"summary", "recommendation"},
}

func main() {
	ctx := context.Background()
	
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	
	googleProvider := providers.NewGoogle([]string{googleAPIKey})
	res, err := googleProvider.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Gemini20Flash{
				StructuredOutput: schemaGoogle,
			},
			SystemMessage: "You are a financial analyst.",
			UserMessage:   "Create a financial analysis of Nvidia at its current valuation",
			Temperature:   0,
			TopP:          0,
			Tags:          map[string]string{},
		},
		http.Client{},
		nil,
	)
	
	if err != nil {
		panic(err)
	}
	
	fmt.Println(res.Content)
}
```

## Advanced Router Configuration

You can configure Heimdall with multiple providers and fallback options:

```go
func main() {
	ctx := context.Background()
	
	// Configure providers
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	
	openAIProvider := providers.NewOpenAI([]string{openAIAPIKey})
	googleProvider := providers.NewGoogle([]string{googleAPIKey})
	anthropicProvider := providers.NewAnthropic([]string{anthropicAPIKey})
	
	// Create router with all providers
	timeout := 60 * time.Second
	router := heimdall.New(timeout, []heimdall.LLMProvider{
		openAIProvider,
		googleProvider,
		anthropicProvider,
	})
	
	// Create request with primary model and fallbacks
	req := request.Completion{
		// Primary model - will be tried first
		Model: models.GPT4O{},
		
		// Fallbacks - will be tried in order if primary fails
		Fallback: []models.Model{
			models.Claude46Sonnet{},
			models.Gemini20Flash{},
		},
		
		SystemMessage: "You are a helpful assistant.",
		UserMessage:   "Explain quantum computing in simple terms",
		Temperature: 0.3,
		TopP:        1.0,
		Tags: map[string]string{
			"env":     "production",
			"type":    "science",
			"purpose": "education",
		},
	}
	
	// Get completion, which will try fallbacks if needed
	response, err := router.Complete(ctx, req)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(response.Content)
	
	// Access request logs for debugging
	fmt.Printf("Request logs: %+v\n", response.RequestLog)
}
```

## Working with Images

### OpenAI with Image Input

```go
func main() {
	ctx := context.Background()
	
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	
	// Read and encode the image file
	fileBytes, err := os.ReadFile("image.jpg")
	if err != nil {
		panic(err)
	}
	
	encodedString := base64.StdEncoding.EncodeToString(fileBytes)
	dataURL := "data:image/jpeg;base64," + encodedString
	
	openAIProvider := providers.NewOpenAI([]string{openAIAPIKey})
	res, err := openAIProvider.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.GPT4O{
				ImageFile: []models.OpenaiImagePayload{
					{
						Url:    dataURL,
						Detail: "high",
					},
				},
			},
			UserMessage: "Describe what you see in this image",
		},
		http.Client{},
		nil,
	)
	
	if err != nil {
		panic(err)
	}
	
	fmt.Println(res.Content)
}
```

### Anthropic with Image Input

```go
func main() {
	ctx := context.Background()
	
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	
	// Read and encode the image file
	fileBytes, err := os.ReadFile("image.jpg")
	if err != nil {
		panic(err)
	}
	
	encodedString := base64.StdEncoding.EncodeToString(fileBytes)
	
	anthropicProvider := providers.NewAnthropic([]string{anthropicAPIKey})
	res, err := anthropicProvider.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Claude46Sonnet{
				ImageFile: map[models.AnthropicImageType]string{
					models.AnthropicImageJpeg: encodedString,
				},
			},
			SystemMessage: "You are a helpful visual assistant.",
			UserMessage:   "Describe what you see in this image",
			Temperature:   0,
			TopP:          0,
			Tags:          map[string]string{},
		},
		http.Client{},
		nil,
	)
	
	if err != nil {
		panic(err)
	}
	
	fmt.Println(res.Content)
}
```

## Error Handling

Heimdall provides comprehensive error handling. Here's an example of how to handle errors:

```go
response, err := router.Complete(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, heimdall.ErrNoChunkHandler):
        // Handle missing chunk handler for streaming
        fmt.Println("Streaming requires a chunk handler")
    case errors.Is(err, heimdall.ErrUnsupportedProvider):
        // Handle case where provider for model is not registered
        fmt.Println("No provider registered for this model")
    default:
        // Handle other errors
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## Supported Models

Heimdall supports various models from different providers:

### OpenAI Models
- GPT-4 (gpt-4-0613)
- GPT-4 Turbo (gpt-4-turbo)
- GPT-4o (gpt-4o-2024-11-20)
- GPT-4o Mini (gpt-4o-mini-2024-07-18)
- GPT-4.1 (gpt-4.1-2025-04-14)
- GPT-4.1 Mini (gpt-4.1-mini-2025-04-14)
- GPT-4.1 Nano (gpt-4.1-nano-2025-04-14)
- GPT-5 (gpt-5-2025-08-07)
- GPT-5 Mini (gpt-5-mini-2025-08-07)
- GPT-5 Nano (gpt-5-nano-2025-08-07)
- GPT-5 Chat (gpt-5-chat-latest)
- GPT-5.1 (gpt-5.1)
- GPT-5.1 Chat (gpt-5.1-chat-latest)
- GPT-5.1 Codex (gpt-5.1-codex)
- GPT-5.1 Codex Mini (gpt-5.1-codex-mini)
- GPT-5.2 (gpt-5.2)
- O1 (o1-2024-12-17)
- O3 Mini (o3-mini-2025-01-31)
- O3 (o3)
- O4 Mini (o4-mini)
- GPT Image (gpt-image-1)

### Anthropic Models
- Claude 4.6 Opus (claude-opus-4-6)
- Claude 4.6 Sonnet (claude-sonnet-4-6)
- Claude 4.5 Opus (claude-opus-4-5-20251101)
- Claude 4.5 Sonnet (claude-sonnet-4-5-20250929)
- Claude 4.5 Haiku (claude-haiku-4-5)
- Claude 4 Opus (claude-opus-4-20250514)
- Claude 4 Sonnet (claude-sonnet-4-20250514)
- ~~Claude 3.7 Sonnet~~ (retired Feb 19, 2026)
- ~~Claude 3.5 Sonnet~~ (retired Oct 28, 2025)
- ~~Claude 3.5 Haiku~~ (retired Feb 19, 2026)
- ~~Claude 3 Opus~~ (retired Jan 5, 2026)

### Google/Gemini Models
- Gemini 3 Flash Preview (gemini-3-flash-preview)
- Gemini 3 Pro Preview (gemini-3-pro-preview) — _deprecated, shuts down March 9, 2026; use gemini-3.1-pro-preview_
- Gemini 3 Pro Image Preview (gemini-3-pro-image-preview)
- Gemini 2.5 Pro (gemini-2.5-pro)
- Gemini 2.5 Flash (gemini-2.5-flash)
- Gemini 2.5 Flash Lite (gemini-2.5-flash-lite)
- Gemini 2.5 Flash Image (gemini-2.5-flash-image)
- Gemini 2.0 Flash (gemini-2.0-flash-001) — _deprecated, retires June 1, 2026_
- Gemini 2.0 Flash Lite (gemini-2.0-flash-lite-001) — _deprecated, retires June 1, 2026_

### Grok Models
- Grok 4 (grok-4)
- Grok 4 Fast (grok-4-fast)
- Grok 3 (grok-3)
- Grok 3 Mini (grok-3-mini)
- Grok 3 Fast (grok-3-fast)
- Grok 3 Mini Fast (grok-3-mini-fast)
- Grok 2 Vision (grok-2-vision-1212)

### Perplexity Models (build tag: `perplexity`)
- Sonar Reasoning Pro (sonar-reasoning-pro)
- Sonar Reasoning (sonar-reasoning)
- Sonar Pro (sonar-pro)
- Sonar (sonar)

### OpenRouter
- Supports any model available on OpenRouter via dynamic model names

## License

This project is licensed under the terms found in the LICENSE file.
