package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Tanishq3031881/CodeForge/backend/internal/files"
	"github.com/Tanishq3031881/CodeForge/backend/internal/rooms"
	"github.com/go-chi/chi/v5"
)

type createFileReq struct {
	Path     string `json:"path"`
	Language string `json:"language"`
}

func (d *Deps) CreateFile(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())
	slug := chi.URLParam(r, "slug")

	var req createFileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}

	f, err := d.Files.CreateFile(r.Context(), slug, userID, req.Path, req.Language)
	if err != nil {
		writeFileErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (d *Deps) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	fileID := chi.URLParam(r, "id")

	if err := d.Files.DeleteFile(r.Context(), slug, userID, fileID); err != nil {
		writeFileErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type fileContentResp struct {
	Content string `json:"content"`
}

type saveContentReq struct {
	Content string `json:"content"`
}

func (d *Deps) GetFileContent(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	fileID := chi.URLParam(r, "id")

	content, err := d.Files.GetContent(r.Context(), slug, userID, fileID)
	if err != nil {
		writeFileErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, fileContentResp{Content: content})
}

func (d *Deps) SaveFileContent(w http.ResponseWriter, r *http.Request) {
	userID, _ := UserIDFrom(r.Context())
	slug := chi.URLParam(r, "slug")
	fileID := chi.URLParam(r, "id")

	var req saveContentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := d.Files.SaveContent(r.Context(), slug, userID, fileID, req.Content); err != nil {
		writeFileErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeFileErr maps file- and room-domain errors to HTTP responses. File
// operations are authorised through the room, so room errors surface here too.
func writeFileErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, files.ErrInvalidInput):
		writeErr(w, http.StatusBadRequest, "path required and language must be supported")
	case errors.Is(err, files.ErrPathTaken):
		writeErr(w, http.StatusConflict, "a file with that path already exists")
	case errors.Is(err, files.ErrNotFound):
		writeErr(w, http.StatusNotFound, "file not found")
	case errors.Is(err, files.ErrTooLarge):
		writeErr(w, http.StatusRequestEntityTooLarge, "file content too large (max 1 MiB)")
	case errors.Is(err, rooms.ErrNotFound):
		writeErr(w, http.StatusNotFound, "room not found")
	case errors.Is(err, rooms.ErrForbidden):
		writeErr(w, http.StatusForbidden, "you don't have access to this room")
	default:
		writeErr(w, http.StatusInternalServerError, "request failed")
	}
}
