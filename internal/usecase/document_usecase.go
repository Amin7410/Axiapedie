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
// Thuật toán: Đi ngược từ revision cần xem → về bản latest (full text),
// thu thập tất cả backward patches dọc đường, rồi áp dụng chúng lần lượt.
func (u *documentUsecase) reconstructContent(ctx context.Context, revID string) (string, error) {
	// Thu thập chuỗi revisions từ revID đến bản full text gần nhất
	var patches []string // Stack các backward patches cần áp dụng
	currentID := revID

	for {
		rev, content, err := u.docRepo.GetRevision(ctx, currentID)
		if err != nil {
			return "", err
		}
		if rev == nil || content == nil {
			return "", ErrNotFound
		}

		if content.ContentType == "full" {
			// Đã tìm thấy bản full text gốc
			fullText := string(content.Data)

			// Áp dụng các backward patches theo thứ tự ngược (từ bản mới → bản cũ)
			for i := len(patches) - 1; i >= 0; i-- {
				fullText, err = delta.ApplyPatch(fullText, patches[i])
				if err != nil {
					return "", err
				}
			}
			return fullText, nil
		}

		// Đây là bản delta, giải nén và lưu vào stack
		patchText, err := delta.DecompressGzip(content.Data)
		if err != nil {
			return "", err
		}
		patches = append(patches, patchText)

		// Đi tiếp về bản cha (bản cũ hơn nữa) — thực chất là tìm revision kế tiếp (mới hơn)
		// Trong Backward Delta, bản cha (ParentID) chỉ về bản cũ hơn.
		// Nhưng bản full nằm ở bản MỚI nhất. Ta cần đi theo hướng ngược lại.
		// => Truy vấn: tìm revision nào có parent_id = currentID (tức revision con, mới hơn)
		if rev.ParentID == nil {
			// Đã đến revision gốc mà vẫn chưa thấy full — fallback
			return string(content.Data), nil
		}
		break
	}

	// Chiến lược đúng cho Backward Delta:
	// 1. Lấy document chứa revision này
	// 2. Bắt đầu từ LatestRevisionID (full text)
	// 3. Đi ngược theo parent_id chain cho tới khi gặp revID
	// 4. Thu thập backward patches và áp dụng

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
	patches = nil
	currentID = *doc.LatestRevisionID

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
