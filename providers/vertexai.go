package providers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/flyx-ai/heimdall/models"
	"github.com/flyx-ai/heimdall/request"
	"github.com/flyx-ai/heimdall/response"
)

// convertSchemaToGenai converts a map[string]any schema to *genai.Schema
func convertSchemaToGenai(schema map[string]any) *genai.Schema {
	if schema == nil {
		return nil
	}

	result := &genai.Schema{}

	if typ, ok := schema["type"].(string); ok {
		result.Type = genai.Type(typ)
	}

	if desc, ok := schema["description"].(string); ok {
		result.Description = desc
	}

	if format, ok := schema["format"].(string); ok {
		result.Format = format
	}

	if enum, ok := schema["enum"].([]any); ok {
		for _, e := range enum {
			if s, ok := e.(string); ok {
				result.Enum = append(result.Enum, s)
			}
		}
	}

	if properties, ok := schema["properties"].(map[string]any); ok {
		result.Properties = make(map[string]*genai.Schema)
		for key, val := range properties {
			if propSchema, ok := val.(map[string]any); ok {
				result.Properties[key] = convertSchemaToGenai(propSchema)
			}
		}
	}

	if required, ok := schema["required"].([]any); ok {
		for _, r := range required {
			if s, ok := r.(string); ok {
				result.Required = append(result.Required, s)
			}
		}
	}

	if items, ok := schema["items"].(map[string]any); ok {
		result.Items = convertSchemaToGenai(items)
	}

	return result
}

// buildThinkingConfig creates a ThinkingConfig based on ThinkBudget or ThinkingLevel
// ThinkingLevel is for Gemini 3 models, ThinkBudget is for Gemini 2.5 models
// They are mutually exclusive - using both will cause an API error for Gemini 3 models
func buildThinkingConfig(budget models.ThinkBudget, level models.ThinkingLevel) *genai.ThinkingConfig {
	// ThinkingLevel takes precedence if set (for Gemini 3 models)
	if level != "" {
		config := &genai.ThinkingConfig{
			IncludeThoughts: true,
		}
		// Convert models.ThinkingLevel (lowercase) to genai.ThinkingLevel (uppercase)
		// Gemini 3 Pro supports: LOW, HIGH
		// Gemini 3 Flash supports: MINIMAL, LOW, MEDIUM, HIGH
		switch level {
		case models.HighThinkingLevel:
			config.ThinkingLevel = genai.ThinkingLevelHigh
		case models.LowThinkingLevel:
			config.ThinkingLevel = genai.ThinkingLevelLow
		default:
			// For any other values (medium, minimal), convert to uppercase
			config.ThinkingLevel = genai.ThinkingLevel(strings.ToUpper(string(level)))
		}
		return config
	}

	// ThinkBudget is for Gemini 2.5 models
	if budget == "" {
		return nil
	}

	config := &genai.ThinkingConfig{
		IncludeThoughts: true,
	}

	switch budget {
	case models.HighThinkBudget:
		budgetVal := int32(24576)
		config.ThinkingBudget = &budgetVal
	case models.MediumThinkBudget:
		budgetVal := int32(12288)
		config.ThinkingBudget = &budgetVal
	case models.LowThinkBudget:
		budgetVal := int32(4096)
		config.ThinkingBudget = &budgetVal
	}

	return config
}

// convertMediaResolution converts models.MediaResolution (lowercase) to genai.MediaResolution
// models uses "high", "medium", "low" while genai uses "MEDIA_RESOLUTION_HIGH", etc.
func convertMediaResolution(resolution models.MediaResolution) genai.MediaResolution {
	switch resolution {
	case models.HighMediaResolution:
		return genai.MediaResolutionHigh
	case models.MediumMediaResolution:
		return genai.MediaResolutionMedium
	case models.LowMediaResolution:
		return genai.MediaResolutionLow
	default:
		return genai.MediaResolutionUnspecified
	}
}

// convertToolsToGenai converts GoogleTool to []*genai.Tool
func convertToolsToGenai(tools models.GoogleTool) []*genai.Tool {
	if len(tools) == 0 {
		return nil
	}

	var result []*genai.Tool

	for _, toolMap := range tools {
		tool := &genai.Tool{}

		// Handle code_execution tool
		if _, hasCodeExec := toolMap["code_execution"]; hasCodeExec {
			tool.CodeExecution = &genai.ToolCodeExecution{}
			result = append(result, tool)
			continue
		}

		// Handle function_declarations - the value is map[string]any
		if funcDeclsMap, ok := toolMap["function_declarations"]; ok {
			// funcDeclsMap is map[string]any, try to get as slice
			if declsSlice, ok := funcDeclsMap["declarations"].([]any); ok {
				for _, decl := range declsSlice {
					if declMap, ok := decl.(map[string]any); ok {
						funcDecl := &genai.FunctionDeclaration{}
						if name, ok := declMap["name"].(string); ok {
							funcDecl.Name = name
						}
						if desc, ok := declMap["description"].(string); ok {
							funcDecl.Description = desc
						}
						if params, ok := declMap["parameters"].(map[string]any); ok {
							funcDecl.Parameters = convertSchemaToGenai(params)
						}
						tool.FunctionDeclarations = append(tool.FunctionDeclarations, funcDecl)
					}
				}
			}
		}

		if len(tool.FunctionDeclarations) > 0 || tool.CodeExecution != nil {
			result = append(result, tool)
		}
	}

	return result
}

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

// extractModelConfig extracts configuration from various Vertex model types
type vertexModelConfig struct {
	Tools            models.GoogleTool
	StructuredOutput map[string]any
	PdfFiles         []models.GooglePdf
	ImageFile        []models.GoogleImagePayload
	Files            []models.GoogleFilePayload
	Thinking         models.ThinkBudget
	ThinkingLevel    models.ThinkingLevel
	MediaResolution  models.MediaResolution
	// Image generation config
	NumberOfImages int
	AspectRatio    models.AspectRatio
	IsImageModel   bool
}

func extractVertexModelConfig(model models.Model) vertexModelConfig {
	config := vertexModelConfig{}

	switch m := model.(type) {
	case models.VertexGemini20Flash:
		config.Tools = m.Tools
		config.StructuredOutput = m.StructuredOutput
		config.PdfFiles = m.PdfFiles
		config.ImageFile = m.ImageFile
		config.Files = m.Files
		config.Thinking = m.Thinking
	case models.VertexGemini20FlashLite:
		config.Tools = m.Tools
		config.StructuredOutput = m.StructuredOutput
		config.PdfFiles = m.PdfFiles
		config.ImageFile = m.ImageFile
		config.Files = m.Files
		config.Thinking = m.Thinking
	case models.VertexGemini25Pro:
		config.Tools = m.Tools
		config.StructuredOutput = m.StructuredOutput
		config.PdfFiles = m.PdfFiles
		config.ImageFile = m.ImageFile
		config.Files = m.Files
		config.Thinking = m.Thinking
	case models.VertexGemini25Flash:
		config.Tools = m.Tools
		config.StructuredOutput = m.StructuredOutput
		config.PdfFiles = m.PdfFiles
		config.ImageFile = m.ImageFile
		config.Files = m.Files
		config.Thinking = m.Thinking
	case models.VertexGemini25FlashLite:
		config.Tools = m.Tools
		config.StructuredOutput = m.StructuredOutput
		config.PdfFiles = m.PdfFiles
		config.ImageFile = m.ImageFile
		config.Files = m.Files
		config.Thinking = m.Thinking
	case models.VertexGemini3ProPreview:
		config.Tools = m.Tools
		config.StructuredOutput = m.StructuredOutput
		config.PdfFiles = m.PdfFiles
		config.ImageFile = m.ImageFile
		config.Files = m.Files
		config.ThinkingLevel = m.ThinkingLevel
		config.MediaResolution = m.MediaResolution
	case models.VertexGemini3FlashPreview:
		config.Tools = m.Tools
		config.StructuredOutput = m.StructuredOutput
		config.PdfFiles = m.PdfFiles
		config.ImageFile = m.ImageFile
		config.Files = m.Files
		config.ThinkingLevel = m.ThinkingLevel
		config.MediaResolution = m.MediaResolution
	case models.VertexGemini25FlashImage:
		config.ImageFile = m.ImageFile
		config.PdfFiles = m.PdfFiles
		config.Files = m.Files
		config.NumberOfImages = m.NumberOfImages
		config.AspectRatio = m.AspectRatio
		config.IsImageModel = true
	case models.VertexGemini3ProImagePreview:
		config.ImageFile = m.ImageFile
		config.PdfFiles = m.PdfFiles
		config.Files = m.Files
		config.ThinkingLevel = m.ThinkingLevel
		config.MediaResolution = m.MediaResolution
		config.NumberOfImages = m.NumberOfImages
		config.AspectRatio = m.AspectRatio
		config.IsImageModel = true
	}

	return config
}

func (v *VertexAI) doRequest(
	ctx context.Context,
	req request.Completion,
	client http.Client,
	chunkHandler func(chunk string) error,
	key string,
) (response.Completion, int, error) {
	// Extract model configuration
	modelConfig := extractVertexModelConfig(req.Model)

	// Build content parts
	var parts []*genai.Content

	// Add history
	for _, his := range req.History {
		parts = append(
			parts,
			genai.NewContentFromText(his.Content, genai.Role(his.Role)),
		)
	}

	// Build user content with text and any files
	userParts := []*genai.Part{
		genai.NewPartFromText(req.UserMessage),
	}

	// Add image files
	for _, img := range modelConfig.ImageFile {
		if strings.HasPrefix(img.Data, "https://") || strings.HasPrefix(img.Data, "gs://") {
			// File URI
			userParts = append(userParts, genai.NewPartFromURI(img.Data, img.MimeType))
		} else {
			// Base64 data
			data, err := base64.StdEncoding.DecodeString(img.Data)
			if err != nil {
				// Try without decoding if it fails (might already be raw)
				data = []byte(img.Data)
			}
			userParts = append(userParts, genai.NewPartFromBytes(data, img.MimeType))
		}
	}

	// Add PDF files
	for _, pdf := range modelConfig.PdfFiles {
		pdfStr := string(pdf)
		if strings.HasPrefix(pdfStr, "https://") || strings.HasPrefix(pdfStr, "gs://") {
			// File URI
			userParts = append(userParts, genai.NewPartFromURI(pdfStr, "application/pdf"))
		} else {
			// Base64 data
			data, err := base64.StdEncoding.DecodeString(pdfStr)
			if err != nil {
				data = []byte(pdfStr)
			}
			userParts = append(userParts, genai.NewPartFromBytes(data, "application/pdf"))
		}
	}

	// Add generic files
	for _, file := range modelConfig.Files {
		if strings.HasPrefix(file.Data, "https://") || strings.HasPrefix(file.Data, "gs://") {
			userParts = append(userParts, genai.NewPartFromURI(file.Data, file.MimeType))
		} else {
			data, err := base64.StdEncoding.DecodeString(file.Data)
			if err != nil {
				data = []byte(file.Data)
			}
			userParts = append(userParts, genai.NewPartFromBytes(data, file.MimeType))
		}
	}

	// Create user content
	userContent := genai.NewContentFromParts(userParts, genai.RoleUser)
	parts = append(parts, userContent)

	// Build generation config
	genConfig := &genai.GenerateContentConfig{}

	// Add system instruction
	if req.SystemMessage != "" {
		genConfig.SystemInstruction = genai.NewContentFromText(req.SystemMessage, genai.RoleUser)
	}

	// Add structured output (response schema)
	if len(modelConfig.StructuredOutput) > 0 {
		genConfig.ResponseMIMEType = "application/json"
		genConfig.ResponseSchema = convertSchemaToGenai(modelConfig.StructuredOutput)
	}

	// Add tools
	if len(modelConfig.Tools) > 0 {
		genConfig.Tools = convertToolsToGenai(modelConfig.Tools)
	}

	// Add thinking configuration
	if thinkingConfig := buildThinkingConfig(modelConfig.Thinking, modelConfig.ThinkingLevel); thinkingConfig != nil {
		genConfig.ThinkingConfig = thinkingConfig
	}

	// Add media resolution
	if modelConfig.MediaResolution != "" {
		genConfig.MediaResolution = convertMediaResolution(modelConfig.MediaResolution)
	}

	// Add image generation configuration for image models
	if modelConfig.IsImageModel {
		genConfig.ResponseModalities = []string{"Text", "Image"}
		if modelConfig.AspectRatio != "" {
			genConfig.ImageConfig = &genai.ImageConfig{
				AspectRatio: string(modelConfig.AspectRatio),
			}
		}
	}

	stream := v.vertexAIClient.Models.GenerateContentStream(
		ctx,
		req.Model.GetName(),
		parts,
		genConfig,
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

			if len(streamPart.Candidates) > 0 &&
				len(streamPart.Candidates[0].Content.Parts) > 0 {
				text := streamPart.Candidates[0].Content.Parts[0].Text
				if text != "Analyzing" {
					_, err := fullContent.WriteString(text)
					if err != nil {
						return response.Completion{}, 0, err
					}

					if chunkHandler != nil {
						if err := chunkHandler(text); err != nil {
							return response.Completion{}, 0, err
						}
					}
				}

				if streamPart.Candidates[0].FinishReason == "STOP" {
					isAnalyzing = false

					if streamPart.UsageMetadata != nil {
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
	projectID string,
	location string,
) (VertexAI, error) {
	client, err := genai.NewClient(
		ctx,
		&genai.ClientConfig{
			Project:    projectID,
			Location:   location,
			Backend:    genai.BackendVertexAI,
			HTTPClient: &http.Client{},
		},
	)
	if err != nil {
		return VertexAI{}, fmt.Errorf("could not setup new genai client: %w", err)
	}

	return VertexAI{
		vertexAIClient: client,
	}, nil
}

var _ LLMProvider = new(VertexAI)
