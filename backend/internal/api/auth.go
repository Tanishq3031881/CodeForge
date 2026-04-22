package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Tanishq3031881/CodeForge/backend/internal/users"
)

type signupReq struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResp struct {
	Token string       `json:"token"`
	User  *users.User  `json:"user"`
}

func (d *Deps) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	token, u, err := d.Users.Signup(r.Context(), users.SignupInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, users.ErrInvalidInput):
			writeErr(w, http.StatusBadRequest, "email, username required; password must be ≥8 chars")
		case errors.Is(err, users.ErrEmailTaken):
			writeErr(w, http.StatusConflict, "email already in use")
		case errors.Is(err, users.ErrUsernameTaken):
			writeErr(w, http.StatusConflict, "username already in use")
		default:
			writeErr(w, http.StatusInternalServerError, "signup failed")
		}
		return
	}
	writeJSON(w, http.StatusOK, authResp{Token: token, User: u})
}

func (d *Deps) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	token, u, err := d.Users.Login(r.Context(), users.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, users.ErrBadLogin) {
			writeErr(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		writeErr(w, http.StatusInternalServerError, "login failed")
		return
	}
	writeJSON(w, http.StatusOK, authResp{Token: token, User: u})
}

func (d *Deps) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	u, err := d.Store.ByID(r.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
