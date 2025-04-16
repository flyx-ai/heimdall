package providers

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/auth"
	"google.golang.org/genai"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

type VertexAI struct {
	vertexAIClient *genai.Client
}

// CompleteResponse implements LLMProvider.
func (v *VertexAI) CompleteResponse(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	requestLog *response.Logging,
) (response.Completion, error) {
	reqLog := &response.Logging{}
	if requestLog == nil {

		req.Tags["request_type"] = "streaming"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to StreamResponse",
				},
			},
			SystemMsg: req.SystemMessage,
			UserMsg:   req.UserMessage,
			Start:     time.Now(),
		}
	}
	if requestLog != nil {
		reqLog = requestLog
	}

	return v.tryWithBackup(ctx, req, http.Client{}, nil, reqLog)
}

func (v *VertexAI) Name() string {
	return models.VertexProvider
}

func (v *VertexAI) StreamResponse(
	ctx context.Context,
	client http.Client,
	req request.Completion,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	reqLog := &response.Logging{}
	if requestLog == nil {
		req.Tags["request_type"] = "streaming"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to StreamResponse",
				},
			},
			SystemMsg: req.SystemMessage,
			UserMsg:   req.UserMessage,
			Start:     time.Now(),
		}
	}
	if requestLog != nil {
		reqLog = requestLog
	}

	reqLog.Events = append(reqLog.Events, response.Event{
		Timestamp: time.Now(),
		Description: fmt.Sprintf(
			"attempting to complete request with key_number: %v",
			1,
		),
	})
	res, _, err := v.doRequest(ctx, req, client, chunkHandler, "")
	if err == nil {
		return res, nil
	}

	reqLog.Events = append(reqLog.Events, response.Event{
		Timestamp: time.Now(),
		Description: fmt.Sprintf(
			"request could not be completed, err: %v",
			err,
		),
	})

	return v.tryWithBackup(ctx, req, client, chunkHandler, requestLog)
}

func (v *VertexAI) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	// TODO: system instructions seems to not work with current SDK version
	// systemInstructions := ""
	var parts []*genai.Content
	parts = append(
		parts,
		genai.NewContentFromText(req.UserMessage, genai.RoleUser),
	)
	// if msg.Role == "file" {
	// 	parts = append(
	// 		parts,
	// 		genai.NewContentFromURI(
	// 			msg.Content,
	// 			string(msg.FileType),
	// 			genai.RoleUser,
	// 		),
	// 	)
	// }
	// if msg.Role == "bytes" {
	// 	parts = append(
	// 		parts,
	// 		genai.NewContentFromBytes(
	// 			[]byte(msg.Content),
	// 			string(msg.FileType),
	// 			genai.RoleUser,
	// 		),
	// 	)
	// }

	stream := v.vertexAIClient.Models.GenerateContentStream(
		ctx,
		req.Model.GetName(),
		parts,
		nil,
	)

	var fullContent strings.Builder
	var usage response.Usage

	now := time.Now()
	isAnalyzing := true

	for isAnalyzing {
		for streamPart, err := range stream {
			if err != nil {
				return response.Completion{}, 0, err
			}
			if len(streamPart.Candidates) == 0 &&
				time.Since(now).Seconds() > 3.0 {
				return response.Completion{}, 0, context.Canceled
			}

			if streamPart.Candidates[0].Content.Parts[0].Text != "Analyzing" {
				_, err := fullContent.WriteString(
					streamPart.Candidates[0].Content.Parts[0].Text,
				)
				if err != nil {
					return response.Completion{}, 0, err
				}

				if chunkHandler != nil {
					if err := chunkHandler(streamPart.Candidates[0].Content.Parts[0].Text); err != nil {
						return response.Completion{}, 0, err
					}
				}
			}

			if streamPart.Candidates[0].FinishReason == "STOP" {
				isAnalyzing = false

				usage = response.Usage{
					PromptTokens: int(
						streamPart.UsageMetadata.PromptTokenCount,
					),
					CompletionTokens: int(
						streamPart.UsageMetadata.CandidatesTokenCount,
					),
					TotalTokens: int(
						streamPart.UsageMetadata.TotalTokenCount,
					),
				}
			}

		}
	}

	return response.Completion{
		Content: fullContent.String(),
		Model:   req.Model.GetName(),
		Usage:   usage,
	}, 0, nil
}

func (v *VertexAI) tryWithBackup(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	maxRetries := 5
	initialBackoff := 100 * time.Millisecond
	maxBackoff := 10 * time.Second

	var lastErr error
	for attempt := range maxRetries {
		requestLog.Events = append(requestLog.Events, response.Event{
			Timestamp: time.Now(),
			Description: fmt.Sprintf(
				"attempting to complete request with expoential backoff. attempt: %v",
				attempt,
			),
		})

		select {
		case <-ctx.Done():
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"context was called with error: %v",
					ctx.Err(),
				),
			})
			return response.Completion{}, ctx.Err()
		default:
			res, resCode, err := v.doRequest(
				ctx,
				req,
				client,
				chunkHandler,
				"",
			)
			if err == nil {
				return res, nil
			}
			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"request could not be completed, err: %v",
					err,
				),
			})

			if !isRetryableError(resCode) {
				requestLog.Events = append(requestLog.Events, response.Event{
					Timestamp: time.Now(),
					Description: fmt.Sprintf(
						"request was not retryable due to err: %v",
						err,
					),
				})
				return response.Completion{}, err
			}

			requestLog.Events = append(requestLog.Events, response.Event{
				Timestamp: time.Now(),
				Description: fmt.Sprintf(
					"request could not be completed, err: %v",
					err,
				),
			})

			lastErr = err

			backoff := min(initialBackoff*time.Duration(
				1<<attempt,
			), maxBackoff)

			var randomBytes [8]byte
			var jitter time.Duration
			if _, err := rand.Read(randomBytes[:]); err != nil {
				jitter = backoff
			} else {
				randFloat := float64(binary.LittleEndian.Uint64(randomBytes[:])) / (1 << 64)
				jitter = time.Duration(float64(backoff) * (0.8 + 0.4*randFloat))
			}

			timer := time.NewTimer(jitter)
			select {
			case <-ctx.Done():
				timer.Stop()
				return response.Completion{}, ctx.Err()
			case <-timer.C:
				continue
			}
		}
	}

	return response.Completion{}, fmt.Errorf(
		"max retries exceeded: %w",
		lastErr,
	)
}

func NewVertexAI(
	ctx context.Context,
	projectID,
	location,
	credentialsJSON string,
) (VertexAI, error) {
	client, err := genai.NewClient(
		ctx,
		&genai.ClientConfig{
			Project:  projectID,
			Location: location,
			Credentials: auth.NewCredentials(&auth.CredentialsOptions{
				JSON: []byte(credentialsJSON),
			}),
			HTTPClient:  &http.Client{},
			HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
		},
	)
	if err != nil {
		return VertexAI{}, errors.New("could not setup new genai client")
	}

	return VertexAI{
		client,
	}, nil
}

var _ LLMProvider = new(VertexAI)
