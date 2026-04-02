package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/andrewidianto12/SewaSini/internal/models"
)

const pbkdf2Iterations = 10000

// hashPassword derives a key using iterated HMAC-SHA256 (PBKDF2-like) with a random salt.
// Format: <hex-salt>$<hex-derived-key>
func hashPassword(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		// Fallback to a fixed salt if random fails.
		salt = []byte("sewasini-fallback")
	}
	dk := pbkdf2HMACSHA256([]byte(password), salt, pbkdf2Iterations)
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(dk)
}

// checkPassword verifies a plaintext password against a stored hash produced by hashPassword.
func checkPassword(password, stored string) bool {
	parts := strings.SplitN(stored, "$", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	dk := pbkdf2HMACSHA256([]byte(password), salt, pbkdf2Iterations)
	return hmac.Equal([]byte(hex.EncodeToString(dk)), []byte(parts[1]))
}

func pbkdf2HMACSHA256(password, salt []byte, iterations int) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)
	dk := mac.Sum(nil)
	for i := 1; i < iterations; i++ {
		mac.Reset()
		mac.Write(dk)
		dk = mac.Sum(nil)
	}
	return dk
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.Title = "Masuk - SewaSini"
	h.render(w, "login", data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.store.GetUserByEmail(email)
	if err != nil || !checkPassword(password, user.Password) {
		data := h.newTemplateData(r)
		data.Title = "Masuk - SewaSini"
		data.Error = "Email atau password salah"
		h.render(w, "login", data)
		return
	}

	h.setSession(w, &SessionData{UserID: user.ID, Name: user.Name})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	data.Title = "Daftar - SewaSini"
	h.render(w, "register", data)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	password := r.FormValue("password")

	if _, err := h.store.GetUserByEmail(email); err == nil {
		data := h.newTemplateData(r)
		data.Title = "Daftar - SewaSini"
		data.Error = "Email sudah terdaftar"
		h.render(w, "register", data)
		return
	}

	user := &models.User{
		Name:      name,
		Email:     email,
		Phone:     phone,
		Password:  hashPassword(password),
		CreatedAt: time.Now(),
	}

	if err := h.store.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	h.setSession(w, &SessionData{UserID: user.ID, Name: user.Name})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
