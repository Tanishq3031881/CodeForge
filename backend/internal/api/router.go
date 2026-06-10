package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(d *Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", HealthHandler)
	r.Get("/health/db", HealthDBHandler(d.Pool))

	// WebSocket entry point for Yjs sync. Lives outside /api because the JWT
	// arrives as a query param (browsers can't set WS headers), so the normal
	// RequireAuth middleware doesn't apply — YjsWS authenticates itself. The
	// doc name is the file ID; the room slug rides along as a query param so we
	// keep a single, slash-free path segment (avoids WS roomname URL-encoding).
	r.Get("/ws/yjs/{id}", d.YjsWS)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/signup", d.Signup)
		r.Post("/auth/login", d.Login)

		r.Group(func(r chi.Router) {
			r.Use(d.RequireAuth)
			r.Get("/me", d.Me)

			r.Post("/rooms", d.CreateRoom)
			r.Get("/rooms", d.ListRooms)
			r.Get("/rooms/{slug}", d.GetRoom)
			r.Delete("/rooms/{slug}", d.DeleteRoom)

			r.Post("/rooms/{slug}/files", d.CreateFile)
			r.Delete("/rooms/{slug}/files/{id}", d.DeleteFile)
			r.Get("/rooms/{slug}/files/{id}/content", d.GetFileContent)
			r.Put("/rooms/{slug}/files/{id}/content", d.SaveFileContent)
		})
	})

	return r
}
