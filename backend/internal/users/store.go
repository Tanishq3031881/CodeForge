package users

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrEmailTaken    = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already in use")
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Create(ctx context.Context, email, username, hash string) (*User, error) {
	const q = `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, password_hash, created_at`
	u := &User{}
	err := s.pool.QueryRow(ctx, q, email, username, hash).
		Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_email_key" {
				return nil, ErrEmailTaken
			}
			if pgErr.ConstraintName == "users_username_key" {
				return nil, ErrUsernameTaken
			}
		}
		return nil, err
	}
	return u, nil
}

func (s *Store) ByEmail(ctx context.Context, email string) (*User, error) {
	return s.one(ctx, `SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1`, email)
}

func (s *Store) ByID(ctx context.Context, id string) (*User, error) {
	return s.one(ctx, `SELECT id, email, username, password_hash, created_at FROM users WHERE id = $1`, id)
}

func (s *Store) one(ctx context.Context, q string, arg any) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, q, arg).
		Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}
