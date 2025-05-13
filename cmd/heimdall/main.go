package main

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"strings"

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

	gApiKey := os.Getenv("GOOGLE_API_KEY")
	// oaApiKey := os.Getenv("OPENAI_API_KEY")

	// f, e := os.ReadFile("cmd/heimdall/doggo.jpeg")
	// if e != nil {
	// 	panic(e)
	// }
	//
	// aApiKey := os.Getenv("ANTHROPIC_API_KEY")
	//
	// // timeout := 1 * time.Minute
	g := providers.NewGoogle([]string{gApiKey})
	// a := providers.NewAnthropic([]string{aApiKey})
	res, err := g.CompleteResponse(
		ctx,
		request.Completion{
			Model: models.Gemini20Flash{
				// ImageFile: map[models.AnthropicImageType]string{
				// 	models.AnthropicImageJpeg: base64.StdEncoding.EncodeToString(
				// 		f,
				// 	),
				// },

				GenerateImage: true,
			},
			// SystemMessage: "you are a master designer",
			UserMessage: "please generate me an image of our ai overlords",
			Temperature: 0,
			TopP:        0,
			Tags:        map[string]string{},
		},
		http.Client{},
		nil,
	)

	// Remove the data:image/jpeg;base64, prefix if it exists
	base64Data := res.Content
	if strings.HasPrefix(base64Data, "data:image/jpeg;base64,") {
		base64Data = strings.TrimPrefix(base64Data, "data:image/jpeg;base64,")
	}

	// Decode base64 string to bytes
	imgBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		slog.Error("Failed to decode base64", "error", err)
		return
	}

	// Write bytes to file
	err = os.WriteFile("new_img_two.png", imgBytes, 0644)
	if err != nil {
		slog.Error("Failed to write image file", "error", err)
		return
	}
	slog.Info("Successfully saved image to new_img.jpeg")

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
