package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"axia-wiki/internal/domain"
)

type documentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new SQLite-backed document repository.
func NewDocumentRepository(db *sql.DB) domain.DocumentRepository {
	return &documentRepository{
		db: db,
	}
}

func (r *documentRepository) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	doc := &domain.Document{}
	query := `SELECT id, title, subtitle, parent_id, is_folder, is_locked, published_revision_id, latest_revision_id, review_status, sort_order, created_at, updated_at FROM documents WHERE id = ?`
	
	var isFolderInt int
	var isLockedInt int
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.Title, &doc.Subtitle, &doc.ParentID, &isFolderInt, &isLockedInt, &doc.PublishedRevisionID, &doc.LatestRevisionID, &doc.ReviewStatus, &doc.SortOrder, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}
	doc.IsFolder = isFolderInt != 0
	doc.IsLocked = isLockedInt != 0
	return doc, nil
}

func (r *documentRepository) GetByTitle(ctx context.Context, title string) (*domain.Document, error) {
	doc := &domain.Document{}
	query := `SELECT id, title, subtitle, parent_id, is_folder, is_locked, published_revision_id, latest_revision_id, review_status, sort_order, created_at, updated_at FROM documents WHERE title = ?`
	
	var isFolderInt int
	var isLockedInt int
	err := r.db.QueryRowContext(ctx, query, title).Scan(
		&doc.ID, &doc.Title, &doc.Subtitle, &doc.ParentID, &isFolderInt, &isLockedInt, &doc.PublishedRevisionID, &doc.LatestRevisionID, &doc.ReviewStatus, &doc.SortOrder, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}
	doc.IsFolder = isFolderInt != 0
	doc.IsLocked = isLockedInt != 0
	return doc, nil
}

func (r *documentRepository) Create(ctx context.Context, doc *domain.Document) error {
	query := `INSERT INTO documents (id, title, subtitle, parent_id, is_folder, is_locked, published_revision_id, latest_revision_id, review_status, sort_order, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	isFolderInt := 0
	if doc.IsFolder {
		isFolderInt = 1
	}
	isLockedInt := 0
	if doc.IsLocked {
		isLockedInt = 1
	}
	_, err := r.db.ExecContext(ctx, query, doc.ID, doc.Title, doc.Subtitle, doc.ParentID, isFolderInt, isLockedInt, doc.PublishedRevisionID, doc.LatestRevisionID, doc.ReviewStatus, doc.SortOrder, doc.CreatedAt, doc.UpdatedAt)
	return err
}

func (r *documentRepository) Update(ctx context.Context, doc *domain.Document) error {
	query := `UPDATE documents 
	          SET title = ?, subtitle = ?, parent_id = ?, is_folder = ?, is_locked = ?, published_revision_id = ?, latest_revision_id = ?, review_status = ?, sort_order = ?, updated_at = ? 
	          WHERE id = ?`
	
	isFolderInt := 0
	if doc.IsFolder {
		isFolderInt = 1
	}
	isLockedInt := 0
	if doc.IsLocked {
		isLockedInt = 1
	}
	doc.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, doc.Title, doc.Subtitle, doc.ParentID, isFolderInt, isLockedInt, doc.PublishedRevisionID, doc.LatestRevisionID, doc.ReviewStatus, doc.SortOrder, doc.UpdatedAt, doc.ID)
	return err
}

func (r *documentRepository) GetAll(ctx context.Context) ([]*domain.Document, error) {
	query := `SELECT id, title, subtitle, parent_id, is_folder, is_locked, published_revision_id, latest_revision_id, review_status, sort_order, created_at, updated_at FROM documents ORDER BY is_folder DESC, sort_order ASC, title ASC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		doc := &domain.Document{}
		var isFolderInt int
		var isLockedInt int
		err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Subtitle, &doc.ParentID, &isFolderInt, &isLockedInt, &doc.PublishedRevisionID, &doc.LatestRevisionID, &doc.ReviewStatus, &doc.SortOrder, &doc.CreatedAt, &doc.UpdatedAt,
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

func (r *documentRepository) SaveRevision(ctx context.Context, rev *domain.Revision, content *domain.TextContent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Rollback in case of failure
	defer tx.Rollback()

	// 1. Insert into revisions
	revQuery := `INSERT INTO revisions (id, document_id, parent_id, author_id, comment, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, revQuery, rev.ID, rev.DocumentID, rev.ParentID, rev.AuthorID, rev.Comment, rev.CreatedAt)
	if err != nil {
		return err
	}

	// 2. Insert into text_contents
	txtQuery := `INSERT INTO text_contents (revision_id, content_type, data) VALUES (?, ?, ?)`
	_, err = tx.ExecContext(ctx, txtQuery, content.RevisionID, content.ContentType, content.Data)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

func (r *documentRepository) GetRevision(ctx context.Context, id string) (*domain.Revision, *domain.TextContent, error) {
	rev := &domain.Revision{}
	content := &domain.TextContent{}

	query := `
		SELECT r.id, r.document_id, r.parent_id, r.author_id, r.comment, r.created_at,
		       t.revision_id, t.content_type, t.data
		FROM revisions r
		JOIN text_contents t ON r.id = t.revision_id
		WHERE r.id = ?
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rev.ID, &rev.DocumentID, &rev.ParentID, &rev.AuthorID, &rev.Comment, &rev.CreatedAt,
		&content.RevisionID, &content.ContentType, &content.Data,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	return rev, content, nil
}

func (r *documentRepository) UpdateTextContent(ctx context.Context, tc *domain.TextContent) error {
	query := `UPDATE text_contents SET content_type = ?, data = ? WHERE revision_id = ?`
	_, err := r.db.ExecContext(ctx, query, tc.ContentType, tc.Data, tc.RevisionID)
	return err
}

func (r *documentRepository) Search(ctx context.Context, query string) ([]*domain.Document, error) {
	// Sử dụng cú pháp MATCH của FTS5 và lấy snippet của tiêu đề (nếu cần, nhưng ở đây ta chỉ trả về mảng Document)
	sqlQuery := `
		SELECT d.id, d.title, d.subtitle, d.parent_id, d.is_folder, d.is_locked, d.published_revision_id, d.latest_revision_id, d.review_status, d.created_at, d.updated_at
		FROM documents d
		JOIN documents_fts fts ON d.id = fts.document_id
		WHERE documents_fts MATCH ?
		GROUP BY d.id
		ORDER BY rank
		LIMIT 20
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		doc := &domain.Document{}
		var isFolderInt int
		var isLockedInt int
		err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Subtitle, &doc.ParentID, &isFolderInt, &isLockedInt, &doc.PublishedRevisionID, &doc.LatestRevisionID, &doc.ReviewStatus, &doc.CreatedAt, &doc.UpdatedAt,
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

func (r *documentRepository) SearchWithTags(ctx context.Context, textQuery string, tags []string) ([]*domain.Document, error) {
	if len(tags) == 0 {
		return r.Search(ctx, textQuery)
	}

	// Xây dựng câu truy vấn SQL động cho danh sách Tags
	placeholders := make([]string, len(tags))
	for i := range tags {
		placeholders[i] = "?"
	}

	var sqlQuery string
	var args []interface{}

	if textQuery != "" {
		sqlQuery = `
			SELECT d.id, d.title, d.subtitle, d.parent_id, d.is_folder, d.is_locked, d.published_revision_id, d.latest_revision_id, d.review_status, d.created_at, d.updated_at
			FROM documents d
			JOIN documents_fts fts ON d.id = fts.document_id
			WHERE documents_fts MATCH ?
			  AND d.id IN (
				SELECT dt.document_id 
				FROM document_tags dt
				JOIN tags t ON dt.tag_id = t.id
				WHERE t.name IN (` + strings.Join(placeholders, ",") + `)
				GROUP BY dt.document_id
				HAVING COUNT(DISTINCT t.name) = ?
			  )
			GROUP BY d.id
			ORDER BY rank
			LIMIT 20
		`
		args = append(args, textQuery)
		for _, tag := range tags {
			args = append(args, tag)
		}
		args = append(args, len(tags))
	} else {
		sqlQuery = `
			SELECT d.id, d.title, d.subtitle, d.parent_id, d.is_folder, d.is_locked, d.published_revision_id, d.latest_revision_id, d.review_status, d.created_at, d.updated_at
			FROM documents d
			WHERE d.id IN (
				SELECT dt.document_id 
				FROM document_tags dt
				JOIN tags t ON dt.tag_id = t.id
				WHERE t.name IN (` + strings.Join(placeholders, ",") + `)
				GROUP BY dt.document_id
				HAVING COUNT(DISTINCT t.name) = ?
			  )
			ORDER BY d.updated_at DESC
			LIMIT 20
		`
		for _, tag := range tags {
			args = append(args, tag)
		}
		args = append(args, len(tags))
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		doc := &domain.Document{}
		var isFolderInt int
		var isLockedInt int
		err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Subtitle, &doc.ParentID, &isFolderInt, &isLockedInt, &doc.PublishedRevisionID, &doc.LatestRevisionID, &doc.ReviewStatus, &doc.CreatedAt, &doc.UpdatedAt,
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

func (r *documentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM documents WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
