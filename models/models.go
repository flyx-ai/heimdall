package models

type Model interface {
	GetProvider() string
	GetName() string
}
