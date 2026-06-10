package files

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("file not found")
	ErrPathTaken = errors.New("a file with that path already exists in this room")
)

type File struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	Path      string    `json:"path"`
	Language  string    `json:"language"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const fileCols = `id, room_id, path, language, updated_at`

func (s *Store) Create(ctx context.Context, roomID, path, language string) (*File, error) {
	const q = `
		INSERT INTO files (room_id, path, language)
		VALUES ($1, $2, $3)
		RETURNING ` + fileCols
	f := &File{}
	err := s.pool.QueryRow(ctx, q, roomID, path, language).
		Scan(&f.ID, &f.RoomID, &f.Path, &f.Language, &f.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrPathTaken
		}
		return nil, err
	}
	return f, nil
}

func (s *Store) ByID(ctx context.Context, id string) (*File, error) {
	f := &File{}
	err := s.pool.QueryRow(ctx, `SELECT `+fileCols+` FROM files WHERE id = $1`, id).
		Scan(&f.ID, &f.RoomID, &f.Path, &f.Language, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *Store) ByRoom(ctx context.Context, roomID string) ([]*File, error) {
	const q = `SELECT ` + fileCols + ` FROM files WHERE room_id = $1 ORDER BY path ASC`
	rows, err := s.pool.Query(ctx, q, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*File{}
	for rows.Next() {
		f := &File{}
		if err := rows.Scan(&f.ID, &f.RoomID, &f.Path, &f.Language, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// GetContent returns the raw text content of a file. Content is fetched
// on demand (not in ByRoom listings) because it can be large.
func (s *Store) GetContent(ctx context.Context, id string) (string, error) {
	var content string
	err := s.pool.QueryRow(ctx, `SELECT content FROM files WHERE id = $1`, id).Scan(&content)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return content, nil
}

// SetContent overwrites a file's content and bumps updated_at.
func (s *Store) SetContent(ctx context.Context, id, content string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE files SET content = $2, updated_at = now() WHERE id = $1`, id, content)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetYjsState returns the persisted Yjs CRDT state for a file, or nil if the
// column is NULL (file never edited in realtime yet).
func (s *Store) GetYjsState(ctx context.Context, id string) ([]byte, error) {
	var state []byte
	err := s.pool.QueryRow(ctx, `SELECT yjs_state FROM files WHERE id = $1`, id).Scan(&state)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return state, nil
}

// SetYjsState persists the CRDT state and the decoded plain text together.
// content is kept in sync so the Stage 5 content endpoint and the sandbox
// always see the latest text without needing to decode Yjs in Go.
func (s *Store) SetYjsState(ctx context.Context, id string, state []byte, text string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE files SET yjs_state = $2, content = $3, updated_at = now() WHERE id = $1`,
		id, state, text)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM files WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
