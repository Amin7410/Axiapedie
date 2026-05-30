package domain

import (
	"context"
	"io"
	"time"
)

// Media represents an uploaded file metadata
type Media struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	FilePath     string    `json:"file_path"`
	UploadedBy   string    `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// MediaRepository defines database operations
type MediaRepository interface {
	Save(ctx context.Context, media *Media) error
	GetByID(ctx context.Context, id string) (*Media, error)
}

// MediaUsecase defines business logic for uploads (now decoupled from HTTP models)
type MediaUsecase interface {
	UploadFile(ctx context.Context, file io.Reader, filename string, uploaderID string) (*Media, error)
}
