package usecase

import (
	"context"
	"strings"

	"axia-wiki/internal/domain"
)

type tagUsecase struct {
	tagRepo domain.TagRepository
}

// NewTagUsecase tạo một thực thể xử lý nghiệp vụ Tag
func NewTagUsecase(repo domain.TagRepository) domain.TagUsecase {
	return &tagUsecase{tagRepo: repo}
}

func (u *tagUsecase) GetTagsByDocument(ctx context.Context, docID string) ([]*domain.Tag, error) {
	return u.tagRepo.GetByDocumentID(ctx, docID)
}

func (u *tagUsecase) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	return u.tagRepo.GetAllTags(ctx)
}

func (u *tagUsecase) CreateTag(ctx context.Context, name string) error {
	normalized := NormalizeTag(name)
	if normalized == "" {
		return nil
	}
	return u.tagRepo.CreateTag(ctx, normalized)
}

// NormalizeTag chuẩn hóa chuỗi tag
func NormalizeTag(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	// Bỏ dấu thăng nếu có ở đầu
	name = strings.TrimPrefix(name, "#")
	return name
}
