package usecase

import (
	"context"
	"errors"
	"time"

	"axia-wiki/internal/domain"
)

// Move moves a document or folder to a new parent folder.
func (u *documentUsecase) Move(ctx context.Context, id string, parentID *string, targetID *string, position string) (*domain.Document, error) {
	doc, err := u.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, ErrNotFound
	}

	userRole, _ := ctx.Value(domain.ContextUserRoleKey).(string)
	if doc.IsLocked && userRole != "admin" {
		return nil, errors.New("this item is locked and cannot be moved")
	}

	// Nếu di chuyển xếp trước/sau (before/after) một node đích, ta lấy parent_id của node đích làm parent_id mới
	if targetID != nil && *targetID != "" && (position == "before" || position == "after") {
		targetDoc, err := u.docRepo.GetByID(ctx, *targetID)
		if err != nil {
			return nil, err
		}
		if targetDoc != nil {
			parentID = targetDoc.ParentID
		}
	}

	if parentID != nil && *parentID != "" {
		if *parentID == id {
			return nil, errors.New("cannot move a folder or page into itself")
		}

		// Kiểm tra vòng lặp vô hạn (cyclic check)
		isChild, err := u.isSubfolder(ctx, id, *parentID)
		if err != nil {
			return nil, err
		}
		if isChild {
			return nil, errors.New("cannot move a parent folder into its own subfolder")
		}

		// Lấy thông tin thư mục cha mới
		parentDoc, err := u.docRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parentDoc == nil {
			return nil, errors.New("destination folder does not exist")
		}
		if !parentDoc.IsFolder {
			return nil, errors.New("destination must be a folder")
		}
		if parentDoc.IsLocked && userRole != "admin" && parentDoc.ID != "unsorted_bin_folder" {
			return nil, errors.New("destination folder is locked; cannot move here")
		}
	}

	var newParent *string
	if parentID != nil && *parentID != "" {
		newParent = parentID
	} else {
		newParent = nil
	}

	doc.ParentID = newParent
	doc.UpdatedAt = time.Now()

	if err := u.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	// Sắp xếp lại sort_order của tất cả anh em (siblings) cùng cấp
	if err := u.reorderSiblings(ctx, newParent, id, targetID, position); err != nil {
		return nil, err
	}

	return doc, nil
}

// reorderSiblings sắp xếp lại sort_order cho các tài liệu cùng cấp
func (u *documentUsecase) reorderSiblings(ctx context.Context, parentID *string, movedID string, targetID *string, position string) error {
	allDocs, err := u.docRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	var siblings []*domain.Document
	for _, d := range allDocs {
		if d.ID == movedID {
			continue
		}

		isSibling := false
		if parentID == nil || *parentID == "" {
			isSibling = (d.ParentID == nil || *d.ParentID == "")
		} else {
			isSibling = (d.ParentID != nil && *d.ParentID == *parentID)
		}

		if isSibling {
			siblings = append(siblings, d)
		}
	}

	movedDoc, err := u.docRepo.GetByID(ctx, movedID)
	if err != nil || movedDoc == nil {
		return err
	}

	var ordered []*domain.Document
	inserted := false

	if targetID == nil || *targetID == "" || (position != "before" && position != "after") {
		ordered = append(siblings, movedDoc)
		inserted = true
	} else {
		for _, s := range siblings {
			if s.ID == *targetID {
				if position == "before" {
					ordered = append(ordered, movedDoc)
					ordered = append(ordered, s)
					inserted = true
				} else if position == "after" {
					ordered = append(ordered, s)
					ordered = append(ordered, movedDoc)
					inserted = true
				}
			} else {
				ordered = append(ordered, s)
			}
		}
	}

	if !inserted {
		ordered = append(ordered, movedDoc)
	}

	for i, d := range ordered {
		d.SortOrder = (i + 1) * 10
		if err := u.docRepo.Update(ctx, d); err != nil {
			return err
		}
	}

	return nil
}

// isSubfolder kiểm tra xem parentID có phải là thư mục con nằm trong childID không
func (u *documentUsecase) isSubfolder(ctx context.Context, childID string, parentID string) (bool, error) {
	currParent := parentID
	for currParent != "" {
		if currParent == childID {
			return true, nil
		}
		pDoc, err := u.docRepo.GetByID(ctx, currParent)
		if err != nil {
			return false, err
		}
		if pDoc == nil {
			break
		}
		if pDoc.ParentID == nil {
			break
		}
		currParent = *pDoc.ParentID
	}
	return false, nil
}
