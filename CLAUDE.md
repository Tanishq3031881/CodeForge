# CodeForge — Project Memory (CLAUDE.md)

CodeForge is a real-time collaborative code editor with a Docker-sandboxed
execution backend. It's Tanishq's portfolio project (3rd-year UNSW CS). The
build is structured as 16 stages in `BUILD_GUIDE.txt` — that file is the bible.
**Read `BUILD_GUIDE.txt` for the authoritative stage-by-stage plan.**

User preference: AI-assisted *learning*, not ghostwriting. Explain the *why*,
help him understand, don't just dump finished code. Each stage should end in a
runnable, demo-able, committed state. Don't skip ahead; don't add features.

## Tech Stack
- **Backend:** Go (chi router, pgx/pgxpool, golang-jwt/v5, bcrypt). Standard
  golang-standards layout: `cmd/server/main.go`, `internal/<domain>/`.
- **Frontend:** React + TypeScript + Vite. Zustand (auth state, persisted to
  localStorage), React Router. API wrapper in `src/lib/api.ts`.
- **DB:** Postgres 16 (Docker). Migrations as numbered SQL files (golang-migrate).
- **Infra (later):** Redis, Yjs Node sidecar, Docker sandbox, Caddy, Hetzner/Fly.

## Repo Layout (current)
```
backend/
  cmd/server/main.go
  internal/
    api/      router.go, deps.go, auth.go, health.go, middleware.go
    auth/     jwt.go, password.go
    config/   config.go
    db/       db.go, migrations/ (001_users, 002_rooms, 003_files up+down)
    users/    store.go, service.go
    rooms/    store.go   <-- STAGE 4, only a broken stub so far
frontend/
  src/
    App.tsx           (routes: /login /signup /dashboard)
    lib/  api.ts, store.ts (Zustand auth)
    pages/ Login.tsx, Signup.tsx, Dashboard.tsx (placeholder)
BUILD_GUIDE.txt
docker-compose.yml, Makefile
```

## Progress Tracker
- [x] **Stage 0** — Learn Go
- [x] **Stage 1** — Skeleton (Go server + React, /health). Commit `b9db1e1`.
- [x] **Stage 2** — Postgres + migrations (users, rooms, files tables). Commit `801fd81`.
- [x] **Stage 3** — Auth: signup/login/JWT/bcrypt + Zustand. Commit `a64b4f1`.
      ⚠️ Backend committed, but FRONTEND for Stage 3 is still UNCOMMITTED
      (pages/, lib/store.ts, App.tsx + api.ts edits are untracked/modified).
- [x] **Stage 4** — Rooms & Files CRUD. DONE, verified end-to-end against live
      Postgres (signup→room→file→delete + authz/validation edge cases). Awaiting
      user's manual commit.
- [x] **Stage 5** — Monaco editor (single-user). DONE, verified end-to-end:
      content GET/PUT, persistence, public-viewer read-only, 1 MiB cap.
      `content` column added via migration 004. Awaiting user's manual commit.
      NOTE: not yet driven in a real browser — backend proven via curl, frontend
      type-checks + builds clean. Monaco loads from its CDN loader (needs net).
- [x] **Stage 6** — Real-time Yjs sync. DONE, verified: two clients converge
      bidirectionally through the authenticated Go WS proxy; bad-JWT WS → 401.
      Node sidecar (`yjs-sidecar/`, doc name = file ID, localhost-only),
      `internal/ws/proxy.go`, `api/yjs.go` (`/ws/yjs/{id}?token&slug`),
      frontend `useYjsRoom` + `CollaborativeEditor` (MonacoBinding). The old
      single-user `Editor.tsx` + Stage-5 save flow were removed.
- [x] **Stage 7** — Persist Yjs to Postgres. DONE, verified: edit → kill+restart
      sidecar → fresh client loads persisted text from DB. Internal endpoints
      `GET/POST /internal/files/{id}/yjs-state` gated by `X-Internal-Key`
      (constant-time compare; 401 without it). Sidecar `persistence.js` loads on
      first connect, debounced save (PERSIST_DEBOUNCE_MS) + final save on close;
      stores yjs_state (base64) + decoded text into `files.content`.
- [x] **Stage 8** — Docker sandbox execution. DONE, verified 8/8 end-to-end
      (incl. through the Vite proxy, the browser path). `sandbox/python.Dockerfile`
      + `entrypoint.sh` (reads stdin → /tmp/code.py → exec). `internal/sandbox/`:
      `runner.go` (Docker SDK, hardened flags), `pool.go` (pre-warmed pool),
      embedded `seccomp.json` (default-deny allowlist). `api/exec.go`:
      `POST /api/rooms/{slug}/run` → execution_id, `GET /ws/exec/{id}?token`
      streams stdout/stderr/exit (coder/websocket). Frontend: Run button +
      Ctrl/Cmd+Enter + `OutputPanel`. Vite now proxies `/ws/*` (ws:true).
      Hardening: --network none, --read-only + tmpfs, --memory/--memory-swap
      128m, --cpus 0.5, --pids-limit 64, --cap-drop ALL, --no-new-privileges,
      seccomp, --user 1000, 5s wall-clock. Dropped `--ulimit nproc` (host-wide
      per-UID → EAGAIN; pids-limit is the container-scoped control).
      VERIFIED hostile: memory bomb→OOM 137; fork bomb→EAGAIN (pids-limit);
      network→name-resolution fail (no netns); fs→read-only error; infinite
      loop→timeout kill. Pooled time-to-first-byte ~15-25ms (target <150ms).
      Deps: docker SDK v27.5.1 (go-connections pinned to v0.5.0 via `replace`
      for Windows npipe), coder/websocket. ⚠️ Awaiting user's manual commit.
- [ ] Stage 9 — Cursor presence + polish
- [ ] Stage 10 — Deployment
- [ ] Stage 11 — Hardening
- [ ] Stage 12 — Launch (README, blog, Show HN)

## STAGE 4 — Exact State (where we left off)
Goal: logged-in users create rooms, see them on dashboard, open a room,
create/delete files inside it.

What exists:
- `backend/internal/rooms/store.go` — ONLY the `Room` struct, and it's BROKEN:
  no `package rooms` declaration, no `import "time"`, and it's indented as if
  pasted into a function. Won't compile. Needs rewrite.

What's NOT done yet (Stage 4 checklist):
- Backend `rooms/store.go`: Create, ByID, BySlug, ListByOwner, Delete + slug gen.
- Backend `rooms/service.go`: CreateRoom(ownerID,name), OpenRoom(slug,userID).
- Backend `files/store.go` + `files/service.go`: Create, ByRoom, Delete.
- Backend `api/rooms.go`: POST/GET /api/rooms, GET/DELETE /api/rooms/:slug.
- Backend `api/files.go`: POST /api/rooms/:slug/files, DELETE .../files/:id.
- Wire all new routes into `api/router.go` (currently only auth + /me).
- Frontend `Dashboard.tsx`: UPDATE from placeholder → fetch+show room cards.
- Frontend `pages/Room.tsx` (NEW, routed /r/:slug): file tree + editor placeholder.
- Frontend `components/CreateRoomModal.tsx`, `components/FileTree.tsx` (NEW;
  no components/ dir exists yet).
- Add /r/:slug route to App.tsx.

Stage 4 "Done when": create room → appears on dashboard → open → file tree →
add/delete file → refresh persists.

## Immediate Housekeeping (recommend before/at start of Stage 4)
1. Commit the uncommitted Stage 3 frontend (pages/, store.ts, App.tsx, api.ts).
   It's currently untracked — Stage 3 isn't fully saved.
2. Fix/replace `rooms/store.go` (add package + imports + real methods).

## Conventions
- Repository pattern: `store.go` = SQL only; `service.go` = business logic;
  `api/*.go` = thin HTTP handlers. Keep SQL out of handlers.
- New domains get their own `internal/<domain>/` package.
- Commit + screen-recap at the end of each stage. Push to public GitHub.
- Windows dev box, PowerShell. Go backend, npm frontend (Vite on :5173,
  backend on :8080, /api proxied).
