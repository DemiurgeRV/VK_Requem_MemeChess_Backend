package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"meme_chess/internal/user"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidUsername    = errors.New("username must be 3–32 characters, letters, digits, underscore")
	ErrDuplicateUser      = errors.New("username or email already taken")
	ErrMissingToken       = errors.New("missing bearer token")
	ErrInvalidToken       = errors.New("invalid token")
)

type Service struct {
	users *user.Repository
	jwt   *JWTManager
}

func NewService(users *user.Repository, jwt *JWTManager) *Service {
	return &Service{users: users, jwt: jwt}
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (token string, userID string, err error) {
	username := strings.TrimSpace(in.Username)
	if !validUsername(username) {
		return "", "", ErrInvalidUsername
	}
	if len(in.Password) < 8 {
		return "", "", ErrWeakPassword
	}

	email := strings.TrimSpace(in.Email)
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hash password: %w", err)
	}

	id, err := s.users.Create(ctx, username, emailPtr, string(hash))
	if err != nil {
		if isUniqueViolation(err) {
			return "", "", ErrDuplicateUser
		}
		return "", "", err
	}

	token, err = s.jwt.Generate(id)
	if err != nil {
		return "", "", err
	}
	return token, id, nil
}

type LoginInput struct {
	Login    string
	Password string
}

func (s *Service) Login(ctx context.Context, in LoginInput) (token string, u *user.User, err error) {
	login := strings.TrimSpace(in.Login)
	if login == "" || in.Password == "" {
		return "", nil, ErrInvalidCredentials
	}

	u, err = s.users.GetByLogin(ctx, login)
	if err != nil {
		return "", nil, err
	}
	if u == nil || u.PasswordHash == "" {
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err = s.jwt.Generate(u.ID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *Service) UserFromBearer(ctx context.Context, authorization string) (*user.User, error) {
	raw := strings.TrimSpace(authorization)
	if !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return nil, ErrMissingToken
	}
	token := strings.TrimSpace(raw[7:])
	if token == "" {
		return nil, ErrMissingToken
	}

	claims, err := s.jwt.Parse(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidToken
	}
	return u, nil
}

func validUsername(s string) bool {
	if len(s) < 3 || len(s) > 32 {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
