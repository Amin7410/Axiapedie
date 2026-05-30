package domain

import "context"

// Bookmark đại diện cho một liên kết lưu tài liệu của người dùng
type Bookmark struct {
	UserID     string `json:"user_id"`
	DocumentID string `json:"document_id"`
	CreatedAt  string `json:"created_at"`
}

// BookmarkRepository định nghĩa các thao tác dữ liệu liên quan đến Bookmark
type BookmarkRepository interface {
	Add(ctx context.Context, userID string, docID string) error
	Remove(ctx context.Context, userID string, docID string) error
	IsBookmarked(ctx context.Context, userID string, docID string) (bool, error)
	GetByUser(ctx context.Context, userID string) ([]*Document, error)
}

// BookmarkUsecase định nghĩa các nghiệp vụ xử lý Bookmark
type BookmarkUsecase interface {
	Add(ctx context.Context, userID string, docID string) error
	Remove(ctx context.Context, userID string, docID string) error
	IsBookmarked(ctx context.Context, userID string, docID string) (bool, error)
	GetByUser(ctx context.Context, userID string) ([]*Document, error)
}
