package handlers

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/andrewidianto12/SewaSini/internal/models"
)

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum(nil))
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
	if err != nil || user.Password != hashPassword(password) {
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
