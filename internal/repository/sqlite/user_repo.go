package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"axia-wiki/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new SQLite-backed user repository.
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, username, password_hash, role, COALESCE(google_id, '') AS google_id, created_at FROM users WHERE id = ?`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.GoogleID, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, username, password_hash, role, COALESCE(google_id, '') AS google_id, created_at FROM users WHERE username = ?`
	
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.GoogleID, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	if googleID == "" {
		return nil, nil
	}
	user := &domain.User{}
	query := `SELECT id, username, password_hash, role, COALESCE(google_id, '') AS google_id, created_at FROM users WHERE google_id = ?`
	
	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.GoogleID, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, username, password_hash, role, google_id, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	user.CreatedAt = time.Now()
	var gID interface{} = nil
	if user.GoogleID != "" {
		gID = user.GoogleID
	}
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.Role, gID, user.CreatedAt)
	return err
}

func (r *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, username, password_hash, role, COALESCE(google_id, '') AS google_id, created_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.GoogleID, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET password_hash = ?, role = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, user.PasswordHash, user.Role, user.ID)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
