package usecase

import (
	"context"
	"errors"

	"axia-wiki/internal/domain"
)

// SetHidden hides or unhides a document.
func (u *documentUsecase) SetHidden(ctx context.Context, id string, hidden bool) (*domain.Document, error) {
	doc, err := u.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, ErrNotFound
	}

	// We check the role from the context
	userRole, _ := ctx.Value(domain.ContextUserRoleKey).(string)
	if userRole != "" && userRole != "admin" {
		return nil, errors.New("only administrators can hide or unhide articles")
	}

	doc.IsHidden = hidden
	if err := u.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// BulkSetHidden hides or unhides multiple documents.
func (u *documentUsecase) BulkSetHidden(ctx context.Context, ids []string, hidden bool) error {
	for _, id := range ids {
		if _, err := u.SetHidden(ctx, id, hidden); err != nil {
			return err
		}
	}
	return nil
}
