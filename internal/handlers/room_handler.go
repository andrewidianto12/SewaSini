package handlers

import (
	"net/http"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	rooms := h.store.GetAllRooms()
	featured := rooms
	if len(rooms) > 3 {
		featured = rooms[:3]
	}
	data := h.newTemplateData(r)
	data.Title = "SewaSini - Sewa Ruangan Terbaik"
	data.Rooms = featured
	h.render(w, "index", data)
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	rooms := h.store.GetAllRooms()
	data := h.newTemplateData(r)
	data.Title = "Daftar Ruangan - SewaSini"
	data.Rooms = rooms
	h.render(w, "rooms", data)
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	room, err := h.store.GetRoomByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data := h.newTemplateData(r)
	data.Title = room.Name + " - SewaSini"
	data.Room = room
	h.render(w, "room_detail", data)
}
