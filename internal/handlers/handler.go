package handlers

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/andrewidianto12/SewaSini/internal/models"
	"github.com/andrewidianto12/SewaSini/internal/store"
)

type Handler struct {
	store     *store.Store
	templates map[string]*template.Template
}

func NewHandler(s *store.Store, templates map[string]*template.Template) *Handler {
	return &Handler{store: s, templates: templates}
}

type SessionData struct {
	UserID string
	Name   string
}

type TemplateData struct {
	Title    string
	Flash    string
	User     *models.User
	Room     *models.Room
	Rooms    []*models.Room
	Booking  *models.Booking
	Bookings []*models.Booking
	Error    string
	Session  *SessionData
}

func (h *Handler) getSession(r *http.Request) *SessionData {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil
	}
	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil
	}
	return &session
}

func (h *Handler) setSession(w http.ResponseWriter, session *SessionData) {
	data, err := json.Marshal(session)
	if err != nil {
		return
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) render(w http.ResponseWriter, tmplName string, data *TemplateData) {
	tmpl, ok := h.templates[tmplName]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) newTemplateData(r *http.Request) *TemplateData {
	return &TemplateData{
		Session: h.getSession(r),
	}
}
