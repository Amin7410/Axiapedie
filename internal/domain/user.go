package domain

import (
	"context"
	"time"
)

// User represents an account in the system.
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string // 'guest', 'reader', 'writer', 'admin'
	CreatedAt    time.Time
}

// UserRepository defines interface for user data operations.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
}

// AuthUsecase defines business logic for authentication.
type AuthUsecase interface {
	Login(ctx context.Context, username, password string) (*User, error)
}

