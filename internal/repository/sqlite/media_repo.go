package sqlite

import (
	"context"
	"database/sql"
	"axia-wiki/internal/domain"
)

type mediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) domain.MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Save(ctx context.Context, media *domain.Media) error {
	query := `
		INSERT INTO media (id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		media.ID, media.Filename, media.OriginalName, media.MimeType, media.FileSize, media.FilePath, media.UploadedBy, media.CreatedAt,
	)
	return err
}

func (r *mediaRepository) GetByID(ctx context.Context, id string) (*domain.Media, error) {
	query := `SELECT id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at FROM media WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	
	m := &domain.Media{}
	err := row.Scan(&m.ID, &m.Filename, &m.OriginalName, &m.MimeType, &m.FileSize, &m.FilePath, &m.UploadedBy, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}
