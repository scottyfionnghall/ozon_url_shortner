package storage

import (
	"context"
)

// Интерфейс для баз данных, который определяет основные функции для работы API.
type Storage interface {
	AddURL(context.Context, *URL) error
	ReturnURL(context.Context, string) (*URL, error)
	CheckExists(string, string) (*URL, error)
	Init() error
	ClearTable() error
}
