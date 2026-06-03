package rooms

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

// CreateRoom validates the name, generates a unique slug, and persists the room.
// Slug collisions are astronomically rare but we retry a few times to be safe.
func (s *Service) CreateRoom(ctx context.Context, ownerID, name string, isPublic bool) (*Room, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 100 {
		return nil, ErrInvalidInput
	}

	const maxAttempts = 5
	for range maxAttempts {
		slug, err := GenerateSlug()
		if err != nil {
			return nil, err
		}
		room, err := s.store.Create(ctx, slug, name, ownerID, isPublic)
		if errors.Is(err, ErrSlugTaken) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return room, nil
	}
	return nil, ErrSlugTaken
}

// OpenRoom returns a room the user is allowed to view: their own, or any public room.
func (s *Service) OpenRoom(ctx context.Context, slug, userID string) (*Room, error) {
	room, err := s.store.BySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if room.OwnerID != userID && !room.IsPublic {
		return nil, ErrForbidden
	}
	return room, nil
}

func (s *Service) ListRooms(ctx context.Context, ownerID string) ([]*Room, error) {
	return s.store.ListByOwner(ctx, ownerID)
}

// RequireOwner returns the room only if the requester owns it. Used by callers
// that mutate a room's contents (e.g. creating or deleting files).
func (s *Service) RequireOwner(ctx context.Context, slug, userID string) (*Room, error) {
	room, err := s.store.BySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if room.OwnerID != userID {
		return nil, ErrForbidden
	}
	return room, nil
}

// DeleteRoom removes a room only if the requester owns it.
func (s *Service) DeleteRoom(ctx context.Context, slug, userID string) error {
	room, err := s.RequireOwner(ctx, slug, userID)
	if err != nil {
		return err
	}
	return s.store.Delete(ctx, room.ID)
}
