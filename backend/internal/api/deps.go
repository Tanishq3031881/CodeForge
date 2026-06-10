package api

import (
	"github.com/Tanishq3031881/CodeForge/backend/internal/auth"
	"github.com/Tanishq3031881/CodeForge/backend/internal/files"
	"github.com/Tanishq3031881/CodeForge/backend/internal/rooms"
	"github.com/Tanishq3031881/CodeForge/backend/internal/users"
	"github.com/Tanishq3031881/CodeForge/backend/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool   *pgxpool.Pool
	Users  *users.Service
	Store  *users.Store
	Rooms  *rooms.Service
	Files  *files.Service
	Issuer *auth.Issuer
	Yjs    *ws.Proxy
	// InternalKey is the shared secret the Yjs sidecar presents on /internal/*.
	InternalKey string
}
