package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Internal endpoints carry sidecar↔backend traffic and are gated by the shared
// X-Internal-Key secret (see RequireInternal), not by user JWTs.

// GetYjsState returns a file's persisted CRDT state as raw bytes, or 204 if the
// file has never been edited in realtime (no state yet).
func (d *Deps) GetYjsState(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	state, err := d.Files.LoadYjsState(r.Context(), id)
	if err != nil {
		writeFileErr(w, err)
		return
	}
	if len(state) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(state)
}

type saveYjsReq struct {
	// State is the base64-encoded Yjs update (Y.encodeStateAsUpdate).
	State string `json:"state"`
	// Text is the decoded plain text, kept in sync so the content endpoint and
	// the sandbox don't need to decode Yjs.
	Text string `json:"text"`
}

// SaveYjsState persists a file's CRDT state and decoded text. Called by the
// sidecar on a debounce while editing and once more when the last client
// disconnects.
func (d *Deps) SaveYjsState(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req saveYjsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	state, err := base64.StdEncoding.DecodeString(req.State)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "state must be base64")
		return
	}
	if err := d.Files.SaveYjsState(r.Context(), id, state, req.Text); err != nil {
		writeFileErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
