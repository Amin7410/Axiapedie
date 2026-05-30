package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"strings"
	"time"

	"axia-wiki/internal/domain"

	"github.com/chai2010/webp"
	"github.com/google/uuid"
)

type mediaUsecase struct {
	mediaRepo domain.MediaRepository
	storage   domain.StorageService
}

func NewMediaUsecase(repo domain.MediaRepository, storage domain.StorageService) domain.MediaUsecase {
	return &mediaUsecase{
		mediaRepo: repo,
		storage:   storage,
	}
}

func (u *mediaUsecase) UploadFile(ctx context.Context, file io.Reader, filename string, uploaderID string) (*domain.Media, error) {
	// 1. Kiểm tra an toàn
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return nil, errors.New("unsupported file format. Only JPG and PNG are allowed")
	}

	// 2. Tối ưu ảnh sang WebP (CGO)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 3. Chuẩn bị thư mục và tên file
	newID := uuid.New().String()
	newFilename := newID + ".webp"
	dateDir := time.Now().Format("2006/01")
	relativeSavePath := filepath.Join(dateDir, newFilename)

	// 4. Lưu dưới dạng WebP vào Buffer
	var buf bytes.Buffer
	err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("failed to encode webp: %w", err)
	}
	finalSize := int64(buf.Len())

	// 5. Lưu vào Storage (hạ tầng lưu trữ)
	err = u.storage.Save(ctx, relativeSavePath, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to save file to storage: %w", err)
	}

	// 6. Lưu Metadata vào DB
	media := &domain.Media{
		ID:           newID,
		Filename:     newFilename,
		OriginalName: filename,
		MimeType:     "image/webp",
		FileSize:     finalSize,
		FilePath:     fmt.Sprintf("/uploads/%s/%s", dateDir, newFilename),
		UploadedBy:   uploaderID,
		CreatedAt:    time.Now(),
	}

	err = u.mediaRepo.Save(ctx, media)
	if err != nil {
		// Rollback file từ storage
		u.storage.Delete(ctx, relativeSavePath)
		return nil, err
	}

	return media, nil
}
