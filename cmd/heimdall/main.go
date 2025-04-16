package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/providers"
	"github.com/flyx-ai/heimdall/request"
)

type Schema struct {
	Revenue                  int      `json:"revenue"`
	CompetitiveAdvantages    []string `json:"competitive_advantages"`
	CompetitiveDisadvantages []string `json:"competitive_disadvantages"`
}

func main() {
	ctx := context.Background()

	// gApiKey := os.Getenv("GOOGLE_API_KEY")
	oaApiKey := os.Getenv("OPENAI_API_KEY")

	// timeout := 1 * time.Minute
	// g := providers.NewGoogle([]string{gApiKey})
	oa := providers.NewOpenAI([]string{oaApiKey})
	res, err := oa.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.O3Mini{
				StructuredOutput: nil,
			},
			UserMessage: "analyze the current performance of anthropic",
			Temperature: 0,
			TopP:        0,
			Tags:        map[string]string{},
		},
		http.Client{},
		nil,
	)
	slog.Info("############# ERR ##################", "err", err)
	slog.Info("############# RES ##################", "res", res.Content)

	// router := heimdall.New(timeout, []providers.LLMProvider{g, oa})
	//
	// req := request.Completion{
	// 	Model: models.Gemini15Flash{},
	// 	Messages: []request.Message{
	// 		{
	// 			Role:    "system",
	// 			Content: "you are a helpful assistant.",
	// 		},
	// 		{
	// 			Role:    "user",
	// 			Content: "please make a detailed analysis of the NVIDIA's current valuation.",
	// 		},
	// 	},
	// 	Fallback: []models.Model{
	// 		models.Gemini15Flash{},
	// 	},
	// 	Temperature: 1,
	// 	Tags: map[string]string{
	// 		"env":  "test",
	// 		"type": "stream",
	// 	},
	// 	TopP: 0,
	// }
	//
	// if res, err := router.Complete(ctx, req); err != nil {
	// 	slog.Info("############# ERR ##################", "err", err)
	// } else {
	// 	slog.Info("############# RES ##################", "res", res.Content)
	// }
}
