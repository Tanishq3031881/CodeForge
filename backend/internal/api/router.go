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

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/signup", d.Signup)
		r.Post("/auth/login", d.Login)

		r.Group(func(r chi.Router) {
			r.Use(d.RequireAuth)
			r.Get("/me", d.Me)
		})
	})

	return r
}
