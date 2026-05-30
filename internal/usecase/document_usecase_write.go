package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"axia-wiki/internal/domain"
	"axia-wiki/pkg/delta"
	"github.com/google/uuid"
)

// SaveDraft handles optimistic locking and saves a new draft revision.
// Implements Backward Delta Compression:
//   - Bản mới (R_new) luôn lưu dạng Full Text.
//   - Bản cũ (R_old, nếu đang là full) được chuyển thành Backward Delta + Gzip.
//   - Backward Delta = Patch để tái dựng R_old từ R_new.
func (u *documentUsecase) SaveDraft(ctx context.Context, title, subtitle, contentStr, authorID, baseRevID, comment string, parentID *string, tags []string) (*domain.Document, error) {
	doc, err := u.docRepo.GetByTitle(ctx, title)
	if err != nil {
		return nil, err
	}

	if doc != nil && doc.IsLocked {
		userRole, _ := ctx.Value(domain.ContextUserRoleKey).(string)
		if userRole != "admin" {
			return nil, errors.New("this article is locked; only administrators can edit it")
		}
	}

	isNewDoc := false
	if doc == nil {
		// Create new document
		isNewDoc = true
		doc = &domain.Document{
			ID:           uuid.New().String(),
			Title:        title,
			Subtitle:     subtitle,
			ParentID:     parentID,
			IsFolder:     false,
			ReviewStatus: "draft",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := u.docRepo.Create(ctx, doc); err != nil {
			return nil, err
		}
	} else {
		// Optimistic locking check
		if doc.LatestRevisionID != nil && *doc.LatestRevisionID != baseRevID {
			return nil, ErrConflict
		}
	}

	// Create new revision
	newRevID := uuid.New().String()
	rev := &domain.Revision{
		ID:         newRevID,
		DocumentID: doc.ID,
		AuthorID:   authorID,
		Comment:    comment,
		CreatedAt:  time.Now(),
	}
	if !isNewDoc {
		rev.ParentID = doc.LatestRevisionID
	}

	// === BACKWARD DELTA COMPRESSION ===
	// Bước 1: Nếu đây KHÔNG phải bài viết mới, lấy nội dung full hiện tại của bản cũ
	if !isNewDoc && doc.LatestRevisionID != nil {
		_, oldContent, err := u.docRepo.GetRevision(ctx, *doc.LatestRevisionID)
		if err != nil {
			return nil, err
		}

		if oldContent != nil && oldContent.ContentType == "full" {
			oldText := string(oldContent.Data)

			// Check if text, subtitle AND tags are identical to prevent duplicate revisions
			oldTags, err := u.tagRepo.GetByDocumentID(ctx, doc.ID)
			if err == nil {
				var oldNormalizedTags []string
				for _, t := range oldTags {
					oldNormalizedTags = append(oldNormalizedTags, t.Name)
				}
				
				var newNormalizedTags []string
				for _, t := range tags {
					trimmed := strings.TrimSpace(strings.ToLower(t))
					trimmed = strings.TrimPrefix(trimmed, "#")
					if trimmed != "" {
						newNormalizedTags = append(newNormalizedTags, trimmed)
					}
				}
				
				if oldText == contentStr && doc.Subtitle == subtitle && slicesEqual(oldNormalizedTags, newNormalizedTags) {
					return nil, ErrNoChanges
				}
			}

			// Bước 2: Tạo Backward Patch (Patch để tái dựng oldText từ newText)
			backwardPatch := delta.GenerateBackwardPatch(oldText, contentStr)

			// Bước 3: Nén bằng Gzip
			gzippedPatch, err := delta.CompressGzip(backwardPatch)
			if err != nil {
				return nil, err
			}

			// Bước 4: Chỉ chuyển sang delta nếu dung lượng nén nhỏ hơn bản gốc
			if len(gzippedPatch) < len(oldText) {
				oldContent.ContentType = "delta"
				oldContent.Data = gzippedPatch
				if err := u.docRepo.UpdateTextContent(ctx, oldContent); err != nil {
					return nil, err
				}
			}
			// Nếu gzippedPatch >= len(oldText), giữ nguyên bản cũ ở dạng full (trường hợp viết lại hoàn toàn)
		}
	}

	// Bước 5: Lưu bản mới luôn là FULL TEXT
	textContent := &domain.TextContent{
		RevisionID:  newRevID,
		ContentType: "full",
		Data:        []byte(contentStr),
	}

	if err := u.docRepo.SaveRevision(ctx, rev, textContent); err != nil {
		return nil, err
	}

	// Update document pointers
	doc.LatestRevisionID = &newRevID
	doc.ReviewStatus = "draft"
	doc.Subtitle = subtitle
	if err := u.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	// === CẬP NHẬT TAGS ===
	if err := u.tagRepo.ClearDocumentTags(ctx, doc.ID); err != nil {
		return nil, err
	}
	for _, tagName := range tags {
		normalized := strings.TrimSpace(strings.ToLower(tagName))
		normalized = strings.TrimPrefix(normalized, "#")
		if normalized != "" {
			if err := u.tagRepo.AddTagToDocument(ctx, doc.ID, normalized); err != nil {
				return nil, err
			}
		}
	}

	return doc, nil
}

// CreateFolder creates a virtual folder.
func (u *documentUsecase) CreateFolder(ctx context.Context, title string, parentID *string) (*domain.Document, error) {
	// Kiểm tra xem tên folder đã tồn tại chưa
	existing, err := u.docRepo.GetByTitle(ctx, title)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("a document or folder with this title already exists")
	}

	folder := &domain.Document{
		ID:           uuid.New().String(),
		Title:        title,
		ParentID:     parentID,
		IsFolder:     true,
		ReviewStatus: "published", // Folder is automatically published
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := u.docRepo.Create(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

// Rename renames a document/folder title.
func (u *documentUsecase) Rename(ctx context.Context, id string, newTitle string) (*domain.Document, error) {
	doc, err := u.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, ErrNotFound
	}

	userRole, _ := ctx.Value(domain.ContextUserRoleKey).(string)
	if doc.IsLocked && userRole != "admin" {
		return nil, errors.New("this article is locked; only administrators can rename it")
	}

	// Kiểm tra xem tên mới đã tồn tại chưa
	existing, err := u.docRepo.GetByTitle(ctx, newTitle)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, errors.New("a document or folder with this name already exists")
	}

	doc.Title = newTitle
	if err := u.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int)
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		if m[v] == 0 {
			return false
		}
		m[v]--
	}
	return true
}
