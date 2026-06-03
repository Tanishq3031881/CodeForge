package rooms

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("room not found")
	ErrSlugTaken = errors.New("slug already in use")
)

type Room struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const roomCols = `id, slug, name, owner_id, is_public, created_at`

func (s *Store) Create(ctx context.Context, slug, name, ownerID string, isPublic bool) (*Room, error) {
	const q = `
		INSERT INTO rooms (slug, name, owner_id, is_public)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + roomCols
	r := &Room{}
	err := s.pool.QueryRow(ctx, q, slug, name, ownerID, isPublic).
		Scan(&r.ID, &r.Slug, &r.Name, &r.OwnerID, &r.IsPublic, &r.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return r, nil
}

func (s *Store) BySlug(ctx context.Context, slug string) (*Room, error) {
	return s.one(ctx, `SELECT `+roomCols+` FROM rooms WHERE slug = $1`, slug)
}

func (s *Store) ByID(ctx context.Context, id string) (*Room, error) {
	return s.one(ctx, `SELECT `+roomCols+` FROM rooms WHERE id = $1`, id)
}

func (s *Store) ListByOwner(ctx context.Context, ownerID string) ([]*Room, error) {
	const q = `SELECT ` + roomCols + ` FROM rooms WHERE owner_id = $1 ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := []*Room{}
	for rows.Next() {
		r := &Room{}
		if err := rows.Scan(&r.ID, &r.Slug, &r.Name, &r.OwnerID, &r.IsPublic, &r.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, rows.Err()
}

func (s *Store) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) one(ctx context.Context, q string, arg any) (*Room, error) {
	r := &Room{}
	err := s.pool.QueryRow(ctx, q, arg).
		Scan(&r.ID, &r.Slug, &r.Name, &r.OwnerID, &r.IsPublic, &r.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r, nil
}
