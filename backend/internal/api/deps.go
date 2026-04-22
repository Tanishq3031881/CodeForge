package api

import (
	"github.com/Tanishq3031881/CodeForge/backend/internal/auth"
	"github.com/Tanishq3031881/CodeForge/backend/internal/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool   *pgxpool.Pool
	Users  *users.Service
	Store  *users.Store
	Issuer *auth.Issuer
}
