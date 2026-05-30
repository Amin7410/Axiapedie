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
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE id = ?`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt,
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
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`
	
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt,
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
	query := `INSERT INTO users (id, username, password_hash, role, created_at) VALUES (?, ?, ?, ?, ?)`
	user.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt)
	return err
}
