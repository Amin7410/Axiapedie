package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"axia-wiki/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username is already taken")
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

func (u *authUsecase) Register(ctx context.Context, username, password string) (*domain.User, error) {
	existing, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: string(passwordHash),
		Role:         "reader", // Default role set to reader for security
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func sanitizeGoogleUsername(email, name string) string {
	var input string
	if parts := strings.Split(email, "@"); len(parts) > 0 && parts[0] != "" {
		input = parts[0]
	} else {
		input = name
	}
	
	// Keep only alphanumeric and underscore characters
	var sb strings.Builder
	for _, ch := range input {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		} else if ch == '.' || ch == '-' || ch == ' ' {
			sb.WriteRune('_')
		}
	}
	
	username := sb.String()
	// Length limits
	if len(username) < 3 {
		username = username + "_user"
	}
	if len(username) > 15 {
		username = username[:15]
	}
	return username
}

func (u *authUsecase) LoginOrCreateWithGoogle(ctx context.Context, googleID, email, name string) (*domain.User, error) {
	if googleID == "" {
		return nil, errors.New("google ID cannot be empty")
	}

	// 1. Check if user already exists by Google ID
	user, err := u.userRepo.GetByGoogleID(ctx, googleID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	// 2. Otherwise, create a new user. Generate username from email/name
	baseUsername := sanitizeGoogleUsername(email, name)
	username := baseUsername

	// 3. Ensure uniqueness of the generated username
	counter := 1
	for {
		existing, err := u.userRepo.GetByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			break
		}
		// If taken, append counter or random chars
		suffix := fmt.Sprintf("_%d", counter)
		if len(baseUsername)+len(suffix) > 20 {
			username = baseUsername[:20-len(suffix)] + suffix
		} else {
			username = baseUsername + suffix
		}
		counter++
		if counter > 100 {
			// Fail-safe to prevent infinite loop
			username = baseUsername[:10] + "_" + uuid.New().String()[:8]
			break
		}
	}

	newUser := &domain.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: "", // No local password
		Role:         "reader", // Default role
		GoogleID:     googleID,
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (u *authUsecase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *authUsecase) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// For Google OAuth users who don't have local password yet
	if user.PasswordHash == "" {
		newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.PasswordHash = string(newHash)
		return u.userRepo.Update(ctx, user)
	}

	// Xác nhận mật khẩu cũ
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	// Băm mật khẩu mới
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(newHash)
	return u.userRepo.Update(ctx, user)
}

func (u *authUsecase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return u.userRepo.GetAll(ctx)
}

func (u *authUsecase) UpdateUserRole(ctx context.Context, id string, role string) error {
	// Kiểm tra vai trò hợp lệ
	if role != "reader" && role != "writer" && role != "admin" {
		return errors.New("invalid role value")
	}

	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.Role = role
	return u.userRepo.Update(ctx, user)
}

func (u *authUsecase) DeleteUser(ctx context.Context, id string) error {
	return u.userRepo.Delete(ctx, id)
}
