package usecase

import (
	"context"
	"errors"

	"axia-wiki/internal/domain"
)

// SetLock locks or unlocks a document.
func (u *documentUsecase) SetLock(ctx context.Context, id string, locked bool) (*domain.Document, error) {
	doc, err := u.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, ErrNotFound
	}

	userRole, _ := ctx.Value(domain.ContextUserRoleKey).(string)
	if userRole != "admin" {
		return nil, errors.New("only administrators can lock or unlock articles")
	}

	doc.IsLocked = locked
	if err := u.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// BulkSetLock locks or unlocks multiple documents.
func (u *documentUsecase) BulkSetLock(ctx context.Context, ids []string, locked bool) error {
	for _, id := range ids {
		doc, err := u.docRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if doc == nil {
			continue
		}

		if _, err := u.SetLock(ctx, id, locked); err != nil {
			return err
		}
	}
	return nil
}
