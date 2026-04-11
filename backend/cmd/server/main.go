package main

import (
	"log"
	"net/http"
	"github.com/Tanishq3031881/CodeForge/backend/internal/config"
	"github.com/Tanishq3031881/CodeForge/backend/internal/api"
)

func main() {
	cfg := config.LoadConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HealthHandler)

	log.Printf("listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}