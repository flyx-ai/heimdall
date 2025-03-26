package providers

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
	"google.golang.org/api/option"
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
		var systemMsg string
		var userMsg string
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				systemMsg = msg.Content
			}
			if msg.Role == "user" {
				userMsg = msg.Content
			}
		}

		req.Tags["request_type"] = "streaming"

		reqLog = &response.Logging{
			Events: []response.Event{
				{
					Timestamp:   time.Now(),
					Description: "start of call to StreamResponse",
				},
			},
			SystemMsg: systemMsg,
			UserMsg:   userMsg,
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

// StreamResponse implements LLMProvider.
func (v *VertexAI) StreamResponse(
	ctx context.Context,
	client http.Client,
	req request.Completion,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	return response.Completion{}, nil
}

func (v *VertexAI) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	return response.Completion{}, 0, nil
}

// tryWithBackup implements LLMProvider.
func (v *VertexAI) tryWithBackup(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	requestLog *response.Logging,
) (response.Completion, error) {
	googleModel, ok := req.Model.(models.GoogleModel)
	if !ok {
		return response.Completion{}, errors.New(
			"could not convert to google args",
		)
	}

	model := v.vertexAIClient.GenerativeModel(googleModel.GetName())

	var parts []genai.Part
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "user" {
			parts = append(parts, genai.Text(msg.Content))
		}
		if msg.Role == "file" {
			parts = append(parts, genai.FileData{
				MIMEType: string(msg.FileType),
				FileURI:  msg.Content,
			})
		}
	}
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
			res, err := model.GenerateContent(ctx, parts...)
			if err == nil {
				rb, err := json.MarshalIndent(res, "", "  ")
				if err != nil {
					return response.Completion{}, err
				}

				return response.Completion{
					Content: string(rb),
					Model:   req.Model.GetName(),
					Usage: response.Usage{
						PromptTokens: int(
							res.UsageMetadata.PromptTokenCount,
						),
						CompletionTokens: int(
							res.UsageMetadata.CandidatesTokenCount,
						),
						TotalTokens: int(
							res.UsageMetadata.TotalTokenCount,
						),
					},
				}, nil
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
		projectID,
		location,
		option.WithCredentialsJSON([]byte(credentialsJSON)),
	)
	if err != nil {
		return VertexAI{}, errors.New("could not setup new genai client")
	}

	return VertexAI{
		client,
	}, nil
}

var _ LLMProvider = new(VertexAI)
