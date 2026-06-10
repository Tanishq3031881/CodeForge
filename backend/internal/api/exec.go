package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
)

// maxRunBytes caps the size of code accepted for execution.
const maxRunBytes = 1 << 20 // 1 MiB

// pendingExec is a run that has been authorised but not yet streamed. The
// two-step flow (POST to register, WS to stream) means output can't start
// before the client is listening, so nothing is lost. Entries are single-use.
type pendingExec struct {
	ownerID   string
	code      string
	createdAt time.Time
}

// execRegistry holds pending executions between the POST and the WS connect.
type execRegistry struct {
	mu sync.Mutex
	m  map[string]pendingExec
}

// NewExecRegistry creates the in-memory store of pending executions.
func NewExecRegistry() *execRegistry { return &execRegistry{m: map[string]pendingExec{}} }

func (e *execRegistry) put(id string, p pendingExec) {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Opportunistically evict anything stale (client never connected the WS).
	for k, v := range e.m {
		if time.Since(v.createdAt) > time.Minute {
			delete(e.m, k)
		}
	}
	e.m[id] = p
}

// take returns and removes a pending execution (single-use).
func (e *execRegistry) take(id string) (pendingExec, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.m[id]
	if ok {
		delete(e.m, id)
	}
	return p, ok
}

type runReq struct {
	FileID string `json:"file_id"`
	// Code is the current editor buffer. Sending it (rather than re-reading the
	// persisted file) means we run exactly what's on screen, with no debounce lag.
	Code string `json:"code"`
}

type runResp struct {
	ExecutionID string `json:"execution_id"`
}

// Run authorises an execution and returns an id the client then streams over
// the WebSocket. Requires view access to the file's room.
func (d *Deps) Run(w http.ResponseWriter, r *http.Request) {
	if d.Sandbox == nil {
		writeErr(w, http.StatusServiceUnavailable, "code execution is not available on this server")
		return
	}
	userID, _ := UserIDFrom(r.Context())
	slug := chi.URLParam(r, "slug")

	var req runReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.Code) > maxRunBytes {
		writeErr(w, http.StatusRequestEntityTooLarge, "code too large (max 1 MiB)")
		return
	}

	f, err := d.Files.RequireViewable(r.Context(), slug, userID, req.FileID)
	if err != nil {
		writeFileErr(w, err)
		return
	}
	if f.Language != "python" {
		writeErr(w, http.StatusBadRequest, "only python execution is supported")
		return
	}

	id := randomID()
	d.Exec.put(id, pendingExec{ownerID: userID, code: req.Code, createdAt: time.Now()})
	writeJSON(w, http.StatusOK, runResp{ExecutionID: id})
}

// execMsg is one frame on the exec WebSocket.
type execMsg struct {
	Type     string `json:"type"`           // "stdout" | "stderr" | "exit" | "error"
	Data     string `json:"data,omitempty"` // output text
	Code     int    `json:"code"`           // exit code (type=="exit"); 0 must not be omitted
	TimedOut bool   `json:"timed_out,omitempty"`
}

// RunStream upgrades to a WebSocket, executes the pending run, and streams its
// output. Auth is via the token query param (browsers can't set WS headers).
func (d *Deps) RunStream(w http.ResponseWriter, r *http.Request) {
	userID, err := d.Issuer.Parse(r.URL.Query().Get("token"))
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid token")
		return
	}
	id := chi.URLParam(r, "id")
	pending, ok := d.Exec.take(id)
	if !ok || pending.ownerID != userID {
		writeErr(w, http.StatusNotFound, "execution not found")
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	// From here errors are reported to the client over the socket, not HTTP.
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := r.Context()
	stdout := &wsWriter{ctx: ctx, conn: conn, stream: "stdout"}
	stderr := &wsWriter{ctx: ctx, conn: conn, stream: "stderr"}

	res, err := d.Sandbox.Run(ctx, pending.code, stdout, stderr)
	if err != nil {
		writeWSMsg(ctx, conn, execMsg{Type: "error", Data: "execution failed: " + err.Error()})
		return
	}
	writeWSMsg(ctx, conn, execMsg{Type: "exit", Code: res.ExitCode, TimedOut: res.TimedOut})
}

// wsWriter turns each Write into one stdout/stderr frame on the socket.
type wsWriter struct {
	ctx    context.Context
	conn   *websocket.Conn
	stream string
}

func (w *wsWriter) Write(p []byte) (int, error) {
	if err := writeWSMsg(w.ctx, w.conn, execMsg{Type: w.stream, Data: string(p)}); err != nil {
		return 0, err
	}
	return len(p), nil
}

func writeWSMsg(ctx context.Context, conn *websocket.Conn, msg execMsg) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, b)
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
