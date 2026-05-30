package domain

import (
	"context"
	"time"
)

type GlossaryTerm struct {
	ID         string    `json:"id"`
	Term       string    `json:"term"`
	Definition string    `json:"definition"`
	DocumentID *string   `json:"document_id"`
	CreatedAt  time.Time `json:"created_at"`
	Aliases    []string  `json:"aliases"`
}

type GlossaryRepository interface {
	GetAllTerms(ctx context.Context) ([]*GlossaryTerm, error)
	GetTermByID(ctx context.Context, id string) (*GlossaryTerm, error)
}

type GlossaryUsecase interface {
	GetAllTerms(ctx context.Context) ([]*GlossaryTerm, error)
	GetTooltipInfo(ctx context.Context, id string) (*GlossaryTerm, error)
	// Trả về đối tượng Aho-Corasick hoặc map chứa các term để goldmark dùng
}
