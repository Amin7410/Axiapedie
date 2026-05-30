package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"axia-wiki/internal/domain"
)

type localStorageService struct {
	uploadDir string
}

// NewLocalStorageService creates a local disk storage service.
func NewLocalStorageService(uploadDir string) domain.StorageService {
	// Ensure directory exists
	os.MkdirAll(uploadDir, os.ModePerm)
	return &localStorageService{uploadDir: uploadDir}
}

func (s *localStorageService) Save(ctx context.Context, path string, content io.Reader) error {
	fullPath := filepath.Join(s.uploadDir, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, content)
	return err
}

func (s *localStorageService) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.uploadDir, path)
	return os.Remove(fullPath)
}
