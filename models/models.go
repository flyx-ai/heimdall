package models

type Model interface {
	GetProvider() string
	GetName() string
}

type StructuredOutput interface {
	GetStructuredOutput() map[string]any
}

type FileReader interface {
	GetFileData() map[string][]byte
}
