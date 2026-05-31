package usecase

import (
	"context"
	"errors"

	"axia-wiki/internal/domain"
	"axia-wiki/pkg/delta"
)

var (
	ErrConflict = errors.New("edit conflict: base revision does not match latest revision")
	ErrNotFound = errors.New("document not found")
	ErrNoChanges = errors.New("no changes detected")
)

type documentUsecase struct {
	docRepo domain.DocumentRepository
	tagRepo domain.TagRepository
}

// NewDocumentUsecase creates a new document usecase.
func NewDocumentUsecase(docRepo domain.DocumentRepository, tagRepo domain.TagRepository) domain.DocumentUsecase {
	return &documentUsecase{
		docRepo: docRepo,
		tagRepo: tagRepo,
	}
}

// GetDocument retrieves a document and its content by title.
func (u *documentUsecase) GetDocument(ctx context.Context, title string) (*domain.Document, string, error) {
	doc, err := u.docRepo.GetByTitle(ctx, title)
	if err != nil {
		return nil, "", err
	}
	if doc == nil {
		return nil, "", ErrNotFound
	}

	// Hạn chế truy cập tài liệu bị ẩn đối với non-admin
	if doc.IsHidden {
		userRole, _ := ctx.Value(domain.ContextUserRoleKey).(string)
		if userRole != "admin" {
			return nil, "", ErrNotFound
		}
	}

	// For standard users, we should read PublishedRevisionID.
	// For writers, LatestRevisionID. We'll default to Published for this base method.
	targetRevID := doc.PublishedRevisionID
	if targetRevID == nil {
		targetRevID = doc.LatestRevisionID // Fallback to draft if no published version (for admin viewing)
	}
	
	if targetRevID == nil {
		return doc, "", nil
	}

	_, content, err := u.docRepo.GetRevision(ctx, *targetRevID)
	if err != nil {
		return nil, "", err
	}

	// Nếu là delta, cần đệ quy giải nén
	contentStr := ""
	if content != nil {
		if content.ContentType == "full" {
			contentStr = string(content.Data)
		} else if content.ContentType == "delta" {
			contentStr, err = u.reconstructContent(ctx, content.RevisionID)
			if err != nil {
				return nil, "", err
			}
		}
	}

	return doc, contentStr, nil
}

// reconstructContent tái dựng nội dung đầy đủ cho một revision cũ bằng Backward Delta.
// Thuật toán: Đi ngược từ bản latest (full text) về revision cần xem,
// thu thập tất cả backward patches dọc đường, rồi áp dụng chúng lần lượt.
func (u *documentUsecase) reconstructContent(ctx context.Context, revID string) (string, error) {
	// Lấy revision đầu tiên để biết document_id
	firstRev, _, err := u.docRepo.GetRevision(ctx, revID)
	if err != nil || firstRev == nil {
		return "", ErrNotFound
	}

	doc, err := u.docRepo.GetByID(ctx, firstRev.DocumentID)
	if err != nil || doc == nil || doc.LatestRevisionID == nil {
		return "", ErrNotFound
	}

	// Bắt đầu từ bản latest (full text), đi ngược theo parent_id chain
	var patches []string
	currentID := *doc.LatestRevisionID

	for currentID != revID {
		rev, content, err := u.docRepo.GetRevision(ctx, currentID)
		if err != nil {
			return "", err
		}
		if rev == nil || content == nil {
			return "", ErrNotFound
		}

		if rev.ParentID == nil {
			// Đã đến gốc mà chưa tìm thấy target revision
			return "", ErrNotFound
		}

		currentID = *rev.ParentID

		// Lấy content của parent
		_, parentContent, err := u.docRepo.GetRevision(ctx, currentID)
		if err != nil {
			return "", err
		}
		if parentContent == nil {
			return "", ErrNotFound
		}

		if parentContent.ContentType == "delta" {
			patchText, err := delta.DecompressGzip(parentContent.Data)
			if err != nil {
				return "", err
			}
			patches = append(patches, patchText)
		}

		if currentID == revID {
			break
		}
	}

	// Lấy full text của latest
	_, latestContent, err := u.docRepo.GetRevision(ctx, *doc.LatestRevisionID)
	if err != nil || latestContent == nil {
		return "", ErrNotFound
	}

	fullText := string(latestContent.Data)

	// Áp dụng backward patches theo thứ tự (từ latest → revID)
	for _, patch := range patches {
		fullText, err = delta.ApplyPatch(fullText, patch)
		if err != nil {
			return "", err
		}
	}

	return fullText, nil
}
