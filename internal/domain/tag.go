package domain

import "context"

// Tag đại diện cho một thẻ phân loại tài liệu
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TagRepository định nghĩa các thao tác dữ liệu liên quan đến Tag
type TagRepository interface {
	GetByDocumentID(ctx context.Context, docID string) ([]*Tag, error)
	AddTagToDocument(ctx context.Context, docID string, tagName string) error
	RemoveTagFromDocument(ctx context.Context, docID string, tagID string) error
	ClearDocumentTags(ctx context.Context, docID string) error
	GetAllTags(ctx context.Context) ([]*Tag, error)
	GetDocumentsByTag(ctx context.Context, tagName string) ([]*Document, error)
	CreateTag(ctx context.Context, name string) error
}

// TagUsecase định nghĩa các nghiệp vụ xử lý Tag
type TagUsecase interface {
	GetTagsByDocument(ctx context.Context, docID string) ([]*Tag, error)
	GetAllTags(ctx context.Context) ([]*Tag, error)
	CreateTag(ctx context.Context, name string) error
}
