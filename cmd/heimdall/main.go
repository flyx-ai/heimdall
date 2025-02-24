package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/flyx-ai/heimdall"
)

func chunkHandler(chunk string) error {
	// slog.Info(
	// 	"#################### CHUNK ######################",
	// 	"value",
	// 	chunk,
	// )
	//
	return nil
}

func main() {
	ctx := context.Background()

	router := heimdall.New(heimdall.RouterConfig{
		ProviderAPIKeys: map[heimdall.Provider][]heimdall.APIKey{
			heimdall.ProviderOpenAI: {
				heimdall.APIKey{
					Name:             "ONE",
					Key:              os.Getenv("OPENAI_API_KEY"),
					RequestsLimit:    500,
					RequestRemaining: 500,
				},
				heimdall.APIKey{
					Name:             "TWO",
					Key:              os.Getenv("OPEN_API_KEY_PROD"),
					RequestsLimit:    10000,
					RequestRemaining: 10000,
				},
			},
		},
		Timeout: 0,
	})

	var wg sync.WaitGroup

	req := heimdall.CompletionRequest{
		Model: heimdall.ModelGPT4,
		Messages: []heimdall.Message{
			{
				Role:    "system",
				Content: "you are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "please make a detailed analysis of the NVIDIA's current valuation.",
			},
		},
		Fallback: []heimdall.Model{
			heimdall.ModelGPT4OMini,
			heimdall.ModelGPT4OModel,
		},
		Temperature: 1,
		Tags: map[string]string{
			"env":  "test",
			"type": "stream",
		},
		TopP: 0,
	}

	router.ReqsStats()
	for i := range 50 {
		wg.Add(1)
		go func(routineNum int) {
			defer wg.Done()

			if err := router.Stream(ctx, req, chunkHandler); err != nil {
				slog.Info("ERRRRRR", "e", err)
			}
		}(i)
	}

	wg.Wait()

	router.ReqsStats()
}
