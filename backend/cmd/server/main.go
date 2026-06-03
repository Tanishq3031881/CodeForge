package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Tanishq3031881/CodeForge/backend/internal/api"
	"github.com/Tanishq3031881/CodeForge/backend/internal/auth"
	"github.com/Tanishq3031881/CodeForge/backend/internal/config"
	"github.com/Tanishq3031881/CodeForge/backend/internal/db"
	"github.com/Tanishq3031881/CodeForge/backend/internal/files"
	"github.com/Tanishq3031881/CodeForge/backend/internal/rooms"
	"github.com/Tanishq3031881/CodeForge/backend/internal/users"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := db.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	issuer := auth.NewIssuer(cfg.JWTSecret, 24*time.Hour)

	userStore := users.NewStore(pool)
	userService := users.NewService(userStore, issuer)

	roomService := rooms.NewService(rooms.NewStore(pool))
	fileService := files.NewService(files.NewStore(pool), roomService)

	deps := &api.Deps{
		Pool:   pool,
		Users:  userService,
		Store:  userStore,
		Rooms:  roomService,
		Files:  fileService,
		Issuer: issuer,
	}

	router := api.NewRouter(deps)

	log.Printf("listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
