package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Tanishq3031881/CodeForge/backend/internal/api"
	"github.com/Tanishq3031881/CodeForge/backend/internal/config"
	"github.com/Tanishq3031881/CodeForge/backend/internal/db"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := db.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("/health/db", api.HealthDBHandler(pool))

	log.Printf("listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
