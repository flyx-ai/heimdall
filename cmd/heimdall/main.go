package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/flyx-ai/heimdall"
)

func main() {
	ctx := context.Background()

	router := heimdall.New(heimdall.RouterConfig{
		ProviderAPIKeys: map[heimdall.Provider][]heimdall.APIKey{
			heimdall.ProviderOpenAI: {
				heimdall.APIKey{
					Name:             "ONE",
					Secret:           os.Getenv("OPENAI_API_KEY"),
					RequestsLimit:    500,
					RequestRemaining: 500,
				},
				heimdall.APIKey{
					Name:             "ONE",
					Secret:           os.Getenv("OPENAI_API_KEY_TWO"),
					RequestsLimit:    10000,
					RequestRemaining: 10000,
				},
			},
			heimdall.ProviderGoogle: {
				heimdall.APIKey{
					Name:             "ONE",
					Secret:           os.Getenv("GOOGLE_API_KEY"),
					RequestsLimit:    10000,
					RequestRemaining: 10000,
				},
			},
		},
		Timeout: 0,
	})

	var wg sync.WaitGroup

	req := heimdall.CompletionRequest{
		Model: heimdall.ModelGPT4OMini,
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
			heimdall.ModelGemini15Pro,
			heimdall.ModelGPT4O,
		},
		Temperature: 1,
		Tags: map[string]string{
			"env":  "test",
			"type": "stream",
		},
		TopP: 0,
	}

	router.ReqsStats()
	for i := range 2 {
		wg.Add(1)
		go func(routineNum int) {
			defer wg.Done()

			if res, err := router.Complete(ctx, req); err != nil {
				slog.Info("ERRRRRR", "e", err)
			} else {
				slog.Info("ITERATION DONE", "i", i, "res", res)
			}
		}(i)
	}

	wg.Wait()

	router.ReqsStats()
}
