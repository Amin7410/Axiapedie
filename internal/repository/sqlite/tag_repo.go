package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"axia-wiki/internal/domain"
	"github.com/google/uuid"
)

type tagRepository struct {
	db *sql.DB
}

// NewTagRepository tạo một SQLite-backed tag repository.
func NewTagRepository(db *sql.DB) domain.TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) GetByDocumentID(ctx context.Context, docID string) ([]*domain.Tag, error) {
	query := `
		SELECT t.id, t.name 
		FROM tags t
		JOIN document_tags dt ON t.id = dt.tag_id
		WHERE dt.document_id = ?
		ORDER BY t.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		t := &domain.Tag{}
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (r *tagRepository) AddTagToDocument(ctx context.Context, docID string, tagName string) error {
	// 1. Tìm hoặc tạo mới Tag để lấy ID
	var tagID string
	queryFind := `SELECT id FROM tags WHERE name = ?`
	err := r.db.QueryRowContext(ctx, queryFind, tagName).Scan(&tagID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Tạo tag mới
			tagID = uuid.New().String()
			queryCreate := `INSERT INTO tags (id, name) VALUES (?, ?)`
			if _, err := r.db.ExecContext(ctx, queryCreate, tagID, tagName); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 2. Liên kết Tag với Document trong bảng document_tags
	queryLink := `INSERT OR IGNORE INTO document_tags (document_id, tag_id) VALUES (?, ?)`
	_, err = r.db.ExecContext(ctx, queryLink, docID, tagID)
	return err
}

func (r *tagRepository) RemoveTagFromDocument(ctx context.Context, docID string, tagID string) error {
	query := `DELETE FROM document_tags WHERE document_id = ? AND tag_id = ?`
	_, err := r.db.ExecContext(ctx, query, docID, tagID)
	return err
}

func (r *tagRepository) ClearDocumentTags(ctx context.Context, docID string) error {
	query := `DELETE FROM document_tags WHERE document_id = ?`
	_, err := r.db.ExecContext(ctx, query, docID)
	return err
}

func (r *tagRepository) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	// Lấy tất cả tag kèm sắp xếp
	query := `SELECT id, name FROM tags ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		t := &domain.Tag{}
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (r *tagRepository) GetDocumentsByTag(ctx context.Context, tagName string) ([]*domain.Document, error) {
	query := `
		SELECT d.id, d.title, d.parent_id, d.is_folder, d.is_locked, d.published_revision_id, d.latest_revision_id, d.review_status, d.created_at, d.updated_at
		FROM documents d
		JOIN document_tags dt ON d.id = dt.document_id
		JOIN tags t ON dt.tag_id = t.id
		WHERE t.name = ?
		ORDER BY d.updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tagName)
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

func (r *tagRepository) CreateTag(ctx context.Context, name string) error {
	var tagID string
	queryFind := `SELECT id FROM tags WHERE name = ?`
	err := r.db.QueryRowContext(ctx, queryFind, name).Scan(&tagID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tagID = uuid.New().String()
			queryCreate := `INSERT INTO tags (id, name) VALUES (?, ?)`
			_, err := r.db.ExecContext(ctx, queryCreate, tagID, name)
			return err
		}
		return err
	}
	return nil
}
