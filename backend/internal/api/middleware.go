package api

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
)

type ctxKey int

const userIDKey ctxKey = iota

func UserIDFrom(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

func (d *Deps) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		userID, err := d.Issuer.Parse(token)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireInternal gates the /internal/* routes used for sidecar↔backend
// traffic. It checks a shared secret in the X-Internal-Key header using a
// constant-time compare so the secret can't be guessed by timing.
func (d *Deps) RequireInternal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Internal-Key")
		if subtle.ConstantTimeCompare([]byte(got), []byte(d.InternalKey)) != 1 {
			writeErr(w, http.StatusUnauthorized, "invalid internal key")
			return
		}
		next.ServeHTTP(w, r)
	})
}
