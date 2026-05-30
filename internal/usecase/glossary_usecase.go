package usecase

import (
	"context"
	"axia-wiki/internal/domain"
)

type glossaryUsecase struct {
	repo domain.GlossaryRepository
}

func NewGlossaryUsecase(repo domain.GlossaryRepository) domain.GlossaryUsecase {
	return &glossaryUsecase{repo: repo}
}

func (u *glossaryUsecase) GetAllTerms(ctx context.Context) ([]*domain.GlossaryTerm, error) {
	return u.repo.GetAllTerms(ctx)
}

func (u *glossaryUsecase) GetTooltipInfo(ctx context.Context, id string) (*domain.GlossaryTerm, error) {
	return u.repo.GetTermByID(ctx, id)
}
