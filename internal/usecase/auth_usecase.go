package usecase

import (
	"context"
	"errors"

	"axia-wiki/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type authUsecase struct {
	userRepo domain.UserRepository
}

// NewAuthUsecase creates a new auth usecase.
func NewAuthUsecase(repo domain.UserRepository) domain.AuthUsecase {
	return &authUsecase{
		userRepo: repo,
	}
}

func (u *authUsecase) Login(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
