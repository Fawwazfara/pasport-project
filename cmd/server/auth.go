package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"passport/internal/auth"
	"passport/internal/model"
	"passport/internal/user_repository"
)

const sessionCookie = "session"

var usernamePattern = regexp.MustCompile(`^[a-z0-9._-]{3,32}$`)

func normalizeUsername(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validPassword(p string) bool {
	return len(p) >= 8
}

func (s *server) currentUser(r *http.Request) (user_repository.User, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		return user_repository.User{}, false
	}
	username, ok := s.sessions.Lookup(cookie.Value)
	if !ok {
		return user_repository.User{}, false
	}
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return user_repository.User{}, false
	}
	return user, true
}

func (s *server) startSession(w http.ResponseWriter, username string) error {
	token, err := s.sessions.Create(username)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(auth.SessionTTL),
	})
	return nil
}

func (s *server) authRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	username := normalizeUsername(req.Username)
	nama := strings.TrimSpace(req.Nama)
	if !usernamePattern.MatchString(username) {
		writeError(w, http.StatusBadRequest, "username 3-32 karakter (huruf, angka, titik, garis bawah, tanda hubung)")
		return
	}
	if nama == "" {
		writeError(w, http.StatusBadRequest, "nama lengkap wajib diisi")
		return
	}
	if !validPassword(req.Password) {
		writeError(w, http.StatusBadRequest, "kata sandi minimal 8 karakter")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user, err := s.users.Create(username, nama, hash)
	if err != nil {
		if errors.Is(err, user_repository.ErrUsernameTaken) {
			writeError(w, http.StatusConflict, "username sudah terdaftar")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := s.startSession(w, user.Username); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, model.UserResponse{Username: user.Username, Nama: user.FullName})
}

func (s *server) authLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	user, err := s.users.FindByUsername(normalizeUsername(req.Username))
	if err != nil || !auth.VerifyPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "username atau kata sandi salah")
		return
	}

	if err := s.startSession(w, user.Username); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, model.UserResponse{Username: user.Username, Nama: user.FullName})
}

func (s *server) authLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	if cookie, err := r.Cookie(sessionCookie); err == nil {
		s.sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *server) authMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	user, ok := s.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "belum masuk")
		return
	}
	writeJSON(w, model.UserResponse{Username: user.Username, Nama: user.FullName})
}

func (s *server) authReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	var req model.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	username := normalizeUsername(req.Username)
	if _, err := s.users.FindByUsername(username); err != nil {
		writeError(w, http.StatusNotFound, "username tidak ditemukan")
		return
	}
	if !validPassword(req.Password) {
		writeError(w, http.StatusBadRequest, "kata sandi minimal 8 karakter")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.users.UpdatePassword(username, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}
