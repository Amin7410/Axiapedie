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
	GoogleID     string // Optional Google OAuth ID
	CreatedAt    time.Time
}

// UserRepository defines interface for user data operations.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
	Create(ctx context.Context, user *User) error
	GetAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// AuthUsecase defines business logic for authentication.
type AuthUsecase interface {
	Login(ctx context.Context, username, password string) (*User, error)
	Register(ctx context.Context, username, password string) (*User, error)
	LoginOrCreateWithGoogle(ctx context.Context, googleID, email, name string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	GetAllUsers(ctx context.Context) ([]*User, error)
	UpdateUserRole(ctx context.Context, id string, role string) error
	DeleteUser(ctx context.Context, id string) error
}

