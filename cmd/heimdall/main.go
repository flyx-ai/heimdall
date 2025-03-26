package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/flyx-ai/heimdall"
	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/providers"
	"github.com/flyx-ai/heimdall/request"
)

func main() {
	ctx := context.Background()

	gApiKey := os.Getenv("GOOGLE_API_KEY")

	timeout := 1 * time.Minute
	g := providers.NewGoogle([]string{gApiKey})
	router := heimdall.New(timeout, []providers.LLMProvider{g})

	req := request.Completion{
		Model: models.Gemini15FlashThink{},
		Messages: []request.Message{
			{
				Role:    "system",
				Content: "you are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "please make a detailed analysis of the NVIDIA's current valuation.",
			},
		},
		Fallback: []models.Model{
			models.Gemini15FlashThink{},
		},
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
