package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/flyx-ai/heimdall"
)

func main() {
	ctx := context.Background()

	gApiKey := os.Getenv("GOOGLE_API_KEY")

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
					Name:             "ONE",
					Key:              os.Getenv("OPENAI_API_KEY_TWO"),
					RequestsLimit:    10000,
					RequestRemaining: 10000,
				},
			},
			heimdall.ProviderGoogle: {
				heimdall.APIKey{
					Name:             "ONE",
					Key:              gApiKey,
					RequestsLimit:    10000,
					RequestRemaining: 10000,
				},
			},
		},
		Timeout: 0,
	})

	req := heimdall.CompletionRequest{
		Model: heimdall.ModelGemini20Flash,
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
		Fallback:    []heimdall.Model{},
		Temperature: 1,
		Tags: map[string]string{
			"env":  "test",
			"type": "stream",
		},
		TopP: 0,
	}

	if res, err := router.Complete(ctx, req); err != nil {
		slog.Info("############# ERR ##################", "err", err)
	} else {
		slog.Info("############# RES ##################", "res", res.Content)
	}
}
