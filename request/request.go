package request

import "github.com/flyx-ai/heimdall/models"

type MimeType string

const (
	MimeTypeJSON       MimeType = "application/json"
	MimeTypeXML        MimeType = "application/xml"
	MimeTypePDF        MimeType = "application/pdf"
	MimeTypeZip        MimeType = "application/zip"
	MimeTypeForm       MimeType = "application/x-www-form-urlencoded"
	MimeTypeOctetStr   MimeType = "application/octet-stream"
	MimeTypeJavaScript MimeType = "application/javascript"

	MimeTypePlainText MimeType = "text/plain"
	MimeTypeHTML      MimeType = "text/html"
	MimeTypeCSS       MimeType = "text/css"
	MimeTypeCSV       MimeType = "text/csv"
	MimeTypeMarkdown  MimeType = "text/markdown"

	MimeTypeJPEG MimeType = "image/jpeg"
	MimeTypePNG  MimeType = "image/png"
	MimeTypeGIF  MimeType = "image/gif"
	MimeTypeSVG  MimeType = "image/svg+xml"
	MimeTypeWebP MimeType = "image/webp"
)

type Completion struct {
	Model         models.Model
	SystemMessage string
	UserMessage   string
	// Messages    []Message
	Fallback    []models.Model
	Temperature float32
	TopP        float32
	Tags        map[string]string `json:"tags"`
}

type Message struct {
	Role     string
	Content  string
	FileType MimeType
}
