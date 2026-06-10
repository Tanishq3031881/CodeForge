package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// YjsWS authenticates a WebSocket connection and forwards it to the Yjs
// sidecar. Browsers can't set headers on WebSocket requests, so the JWT
// arrives as a query parameter instead of an Authorization header.
func (d *Deps) YjsWS(w http.ResponseWriter, r *http.Request) {
	userID, err := d.Issuer.Parse(r.URL.Query().Get("token"))
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid token")
		return
	}
	slug := r.URL.Query().Get("slug")
	fileID := chi.URLParam(r, "id")

	f, err := d.Files.RequireViewable(r.Context(), slug, userID, fileID)
	if err != nil {
		writeFileErr(w, err)
		return
	}
	d.Yjs.Forward(w, r, f.ID)
}
