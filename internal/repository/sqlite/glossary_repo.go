package sqlite

import (
	"context"
	"database/sql"
	"axia-wiki/internal/domain"
)

type glossaryRepo struct {
	db *sql.DB
}

func NewGlossaryRepository(db *sql.DB) domain.GlossaryRepository {
	return &glossaryRepo{db: db}
}

func (r *glossaryRepo) GetAllTerms(ctx context.Context) ([]*domain.GlossaryTerm, error) {
	// Lấy tất cả thuật ngữ (Tạm thời MVP chưa có Aliases)
	query := `SELECT id, term, definition, document_id, created_at FROM glossary_terms`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []*domain.GlossaryTerm
	for rows.Next() {
		t := &domain.GlossaryTerm{}
		if err := rows.Scan(&t.ID, &t.Term, &t.Definition, &t.DocumentID, &t.CreatedAt); err != nil {
			return nil, err
		}
		terms = append(terms, t)
	}
	return terms, nil
}

func (r *glossaryRepo) GetTermByID(ctx context.Context, id string) (*domain.GlossaryTerm, error) {
	query := `SELECT id, term, definition, document_id, created_at FROM glossary_terms WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	
	t := &domain.GlossaryTerm{}
	err := row.Scan(&t.ID, &t.Term, &t.Definition, &t.DocumentID, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}
