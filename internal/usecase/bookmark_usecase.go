package usecase

import (
	"context"
	"errors"

	"axia-wiki/internal/domain"
)

type bookmarkUsecase struct {
	repo domain.BookmarkRepository
}

// NewBookmarkUsecase tạo một thực thể xử lý nghiệp vụ Bookmark
func NewBookmarkUsecase(repo domain.BookmarkRepository) domain.BookmarkUsecase {
	return &bookmarkUsecase{repo: repo}
}

func (u *bookmarkUsecase) Add(ctx context.Context, userID string, docID string) error {
	if userID == "" || docID == "" {
		return errors.New("invalid user ID or document ID")
	}
	return u.repo.Add(ctx, userID, docID)
}

func (u *bookmarkUsecase) Remove(ctx context.Context, userID string, docID string) error {
	if userID == "" || docID == "" {
		return errors.New("invalid user ID or document ID")
	}
	return u.repo.Remove(ctx, userID, docID)
}

func (u *bookmarkUsecase) IsBookmarked(ctx context.Context, userID string, docID string) (bool, error) {
	if userID == "" || docID == "" {
		return false, nil
	}
	return u.repo.IsBookmarked(ctx, userID, docID)
}

func (u *bookmarkUsecase) GetByUser(ctx context.Context, userID string) ([]*domain.Document, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return u.repo.GetByUser(ctx, userID)
}
