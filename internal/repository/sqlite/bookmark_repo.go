package sqlite

import (
	"context"
	"database/sql"

	"axia-wiki/internal/domain"
)

type bookmarkRepository struct {
	db *sql.DB
}

// NewBookmarkRepository tạo một SQLite-backed bookmark repository.
func NewBookmarkRepository(db *sql.DB) domain.BookmarkRepository {
	return &bookmarkRepository{db: db}
}

func (r *bookmarkRepository) Add(ctx context.Context, userID string, docID string) error {
	query := `INSERT OR IGNORE INTO user_bookmarks (user_id, document_id) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, query, userID, docID)
	return err
}

func (r *bookmarkRepository) Remove(ctx context.Context, userID string, docID string) error {
	query := `DELETE FROM user_bookmarks WHERE user_id = ? AND document_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID, docID)
	return err
}

func (r *bookmarkRepository) IsBookmarked(ctx context.Context, userID string, docID string) (bool, error) {
	query := `SELECT COUNT(1) FROM user_bookmarks WHERE user_id = ? AND document_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, docID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *bookmarkRepository) GetByUser(ctx context.Context, userID string) ([]*domain.Document, error) {
	query := `
		SELECT d.id, d.title, d.parent_id, d.is_folder, d.is_locked, d.published_revision_id, d.latest_revision_id, d.review_status, d.created_at, d.updated_at
		FROM documents d
		JOIN user_bookmarks ub ON d.id = ub.document_id
		WHERE ub.user_id = ?
		ORDER BY ub.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		doc := &domain.Document{}
		var isFolderInt, isLockedInt int
		err := rows.Scan(
			&doc.ID, &doc.Title, &doc.ParentID, &isFolderInt, &isLockedInt, 
			&doc.PublishedRevisionID, &doc.LatestRevisionID, &doc.ReviewStatus, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		doc.IsFolder = isFolderInt != 0
		doc.IsLocked = isLockedInt != 0
		docs = append(docs, doc)
	}
	return docs, nil
}
