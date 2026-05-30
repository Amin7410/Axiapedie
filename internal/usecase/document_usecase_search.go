package usecase

import (
	"context"
	"strings"

	"axia-wiki/internal/domain"
)

// Search queries the FTS engine and filters by tags if specified.
func (u *documentUsecase) Search(ctx context.Context, query string) ([]*domain.Document, error) {
	if query == "" {
		return []*domain.Document{}, nil
	}

	// Tách câu query bằng khoảng trắng để bóc tách tags (#tag)
	words := strings.Fields(query)
	var tags []string
	var textWords []string

	for _, w := range words {
		if strings.HasPrefix(w, "#") {
			tag := strings.TrimPrefix(w, "#")
			tag = strings.ToLower(strings.TrimSpace(tag))
			if tag != "" {
				tags = append(tags, tag)
			}
		} else {
			textWords = append(textWords, w)
		}
	}

	textQuery := strings.Join(textWords, " ")
	textQuery = strings.TrimSpace(textQuery)

	var safeQuery string
	if textQuery != "" {
		// Để đơn giản và an toàn với FTS5 query syntax, ta bao query trong cặp nháy kép
		safeQuery = `"` + textQuery + `*"`
	}

	return u.docRepo.SearchWithTags(ctx, safeQuery, tags)
}

// GetAll returns all documents.
func (u *documentUsecase) GetAll(ctx context.Context) ([]*domain.Document, error) {
	return u.docRepo.GetAll(ctx)
}
