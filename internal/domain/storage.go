package domain

import (
	"context"
	"io"
)

// StorageService defines the interface for file storage operations.
type StorageService interface {
	Save(ctx context.Context, path string, content io.Reader) error
	Delete(ctx context.Context, path string) error
}
