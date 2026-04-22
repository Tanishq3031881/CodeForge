package users

import (
	"context"
	"errors"
	"strings"

	"github.com/Tanishq3031881/CodeForge/backend/internal/auth"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrBadLogin     = errors.New("invalid email or password")
)

type Service struct {
	store  *Store
	issuer *auth.Issuer
}

func NewService(store *Store, issuer *auth.Issuer) *Service {
	return &Service{store: store, issuer: issuer}
}

type SignupInput struct {
	Email    string
	Username string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *Service) Signup(ctx context.Context, in SignupInput) (string, *User, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	username := strings.TrimSpace(in.Username)
	if email == "" || username == "" || len(in.Password) < 8 {
		return "", nil, ErrInvalidInput
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return "", nil, err
	}

	u, err := s.store.Create(ctx, email, username, hash)
	if err != nil {
		return "", nil, err
	}

	token, err := s.issuer.Issue(u.ID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (string, *User, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	u, err := s.store.ByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", nil, ErrBadLogin
		}
		return "", nil, err
	}
	if !auth.CheckPassword(u.PasswordHash, in.Password) {
		return "", nil, ErrBadLogin
	}
	token, err := s.issuer.Issue(u.ID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}
