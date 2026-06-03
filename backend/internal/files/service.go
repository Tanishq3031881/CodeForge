package files

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/Tanishq3031881/CodeForge/backend/internal/rooms"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrTooLarge     = errors.New("file content too large")
)

// maxContentBytes caps a single file's content. 1 MiB is far more than any
// reasonable source file and protects the DB from abuse.
const maxContentBytes = 1 << 20

// supported is the set of languages the editor knows how to highlight. Kept
// small for now; extended when multi-language sandboxes land in a later stage.
var supported = map[string]bool{
	"python":     true,
	"javascript": true,
	"typescript": true,
	"go":         true,
	"rust":       true,
	"plaintext":  true,
}

type Service struct {
	store *Store
	rooms *rooms.Service
}

func NewService(store *Store, roomsSvc *rooms.Service) *Service {
	return &Service{store: store, rooms: roomsSvc}
}

// CreateFile adds a file to a room the user owns.
func (s *Service) CreateFile(ctx context.Context, slug, userID, filePath, language string) (*File, error) {
	room, err := s.rooms.RequireOwner(ctx, slug, userID)
	if err != nil {
		return nil, err
	}

	filePath = cleanPath(filePath)
	language = strings.TrimSpace(strings.ToLower(language))
	if filePath == "" || len(filePath) > 255 || !supported[language] {
		return nil, ErrInvalidInput
	}

	return s.store.Create(ctx, room.ID, filePath, language)
}

// ListFiles returns the files in a room the user can view (owner or public).
func (s *Service) ListFiles(ctx context.Context, slug, userID string) ([]*File, error) {
	room, err := s.rooms.OpenRoom(ctx, slug, userID)
	if err != nil {
		return nil, err
	}
	return s.store.ByRoom(ctx, room.ID)
}

// DeleteFile removes a file from a room the user owns. It verifies the file
// actually belongs to that room so a file id can't be deleted via another room.
func (s *Service) DeleteFile(ctx context.Context, slug, userID, fileID string) error {
	room, err := s.rooms.RequireOwner(ctx, slug, userID)
	if err != nil {
		return err
	}
	f, err := s.store.ByID(ctx, fileID)
	if err != nil {
		return err
	}
	if f.RoomID != room.ID {
		return ErrNotFound
	}
	return s.store.Delete(ctx, fileID)
}

// GetContent returns a file's text content if the user can view the room.
func (s *Service) GetContent(ctx context.Context, slug, userID, fileID string) (string, error) {
	room, err := s.rooms.OpenRoom(ctx, slug, userID)
	if err != nil {
		return "", err
	}
	f, err := s.store.ByID(ctx, fileID)
	if err != nil {
		return "", err
	}
	if f.RoomID != room.ID {
		return "", ErrNotFound
	}
	return s.store.GetContent(ctx, fileID)
}

// SaveContent persists file content. Only the room owner may write.
func (s *Service) SaveContent(ctx context.Context, slug, userID, fileID, content string) error {
	if len(content) > maxContentBytes {
		return ErrTooLarge
	}
	room, err := s.rooms.RequireOwner(ctx, slug, userID)
	if err != nil {
		return err
	}
	f, err := s.store.ByID(ctx, fileID)
	if err != nil {
		return err
	}
	if f.RoomID != room.ID {
		return ErrNotFound
	}
	return s.store.SetContent(ctx, fileID, content)
}

// cleanPath normalises a user-supplied file path and strips any directory
// traversal, leaving a clean relative path like "src/main.py".
func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.ReplaceAll(p, "\\", "/")
	p = path.Clean("/" + p) // collapses .. and . against the root
	return strings.TrimPrefix(p, "/")
}
