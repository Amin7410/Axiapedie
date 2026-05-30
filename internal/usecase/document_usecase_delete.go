package usecase

import (
	"context"
	"errors"

	"axia-wiki/internal/domain"
)

// Delete deletes a document/folder.
func (u *documentUsecase) Delete(ctx context.Context, id string) error {
	doc, err := u.docRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if doc == nil {
		return ErrNotFound
	}

	userRole, _ := ctx.Value(domain.ContextUserRoleKey).(string)
	if doc.IsLocked && userRole != "admin" {
		return errors.New("this article is locked; only administrators can delete it")
	}

	// Nếu là thư mục, thực hiện xóa đệ quy tất cả trang/thư mục con
	if doc.IsFolder {
		allDocs, err := u.docRepo.GetAll(ctx)
		if err != nil {
			return err
		}

		// 1. Kiểm tra khóa của các tài liệu con trước khi xóa (nếu không phải admin)
		if userRole != "admin" {
			var checkLocked func(parentID string) error
			checkLocked = func(parentID string) error {
				for _, d := range allDocs {
					if d.ParentID != nil && *d.ParentID == parentID {
						if d.IsLocked {
							return errors.New("cannot delete folder: it contains a locked document: " + d.Title)
						}
						if d.IsFolder {
							if err := checkLocked(d.ID); err != nil {
								return err
							}
						}
					}
				}
				return nil
			}
			if err := checkLocked(id); err != nil {
				return err
			}
		}

		// 2. Thực hiện xóa đệ quy thực tế các tài liệu con
		var deleteChildren func(parentID string) error
		deleteChildren = func(parentID string) error {
			for _, d := range allDocs {
				if d.ParentID != nil && *d.ParentID == parentID {
					if d.IsFolder {
						if err := deleteChildren(d.ID); err != nil {
							return err
						}
					}
					if err := u.docRepo.Delete(ctx, d.ID); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := deleteChildren(id); err != nil {
			return err
		}
	}

	return u.docRepo.Delete(ctx, id)
}

// BulkDelete deletes multiple documents in a single operation.
func (u *documentUsecase) BulkDelete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		doc, err := u.docRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if doc == nil {
			continue // Đã bị xóa trước đó bởi xóa đệ quy của thư mục cha
		}

		if err := u.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
