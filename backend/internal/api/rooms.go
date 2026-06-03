package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Tanishq3031881/CodeForge/backend/internal/files"
	"github.com/Tanishq3031881/CodeForge/backend/internal/rooms"
	"github.com/go-chi/chi/v5"
)

type createRoomReq struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

// roomWithFiles is the payload for opening a single room.
type roomWithFiles struct {
	Room  *rooms.Room   `json:"room"`
	Files []*files.File `json:"files"`
}

func (d *Deps) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())

	var req createRoomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}

	room, err := d.Rooms.CreateRoom(r.Context(), userID, req.Name, req.IsPublic)
	if err != nil {
		if errors.Is(err, rooms.ErrInvalidInput) {
			writeErr(w, http.StatusBadRequest, "name is required (1-100 chars)")
			return
		}
		writeErr(w, http.StatusInternalServerError, "could not create room")
		return
	}
	writeJSON(w, http.StatusCreated, room)
}

func (d *Deps) ListRooms(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())

	list, err := d.Rooms.ListRooms(r.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not list rooms")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (d *Deps) GetRoom(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())
	slug := chi.URLParam(r, "slug")

	room, err := d.Rooms.OpenRoom(r.Context(), slug, userID)
	if err != nil {
		writeRoomErr(w, err)
		return
	}

	fs, err := d.Files.ListFiles(r.Context(), slug, userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not load files")
		return
	}
	writeJSON(w, http.StatusOK, roomWithFiles{Room: room, Files: fs})
}

func (d *Deps) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())
	slug := chi.URLParam(r, "slug")

	if err := d.Rooms.DeleteRoom(r.Context(), slug, userID); err != nil {
		writeRoomErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeRoomErr maps room-domain errors to HTTP responses.
func writeRoomErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, rooms.ErrNotFound):
		writeErr(w, http.StatusNotFound, "room not found")
	case errors.Is(err, rooms.ErrForbidden):
		writeErr(w, http.StatusForbidden, "you don't have access to this room")
	default:
		writeErr(w, http.StatusInternalServerError, "request failed")
	}
}
