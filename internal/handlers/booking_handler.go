package handlers

import (
	"net/http"
	"time"

	"github.com/andrewidianto12/SewaSini/internal/models"
)

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	roomID := r.FormValue("room_id")
	room, err := h.store.GetRoomByID(roomID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renterName := r.FormValue("renter_name")
	renterEmail := r.FormValue("renter_email")
	renterPhone := r.FormValue("renter_phone")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	notes := r.FormValue("notes")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		data := h.newTemplateData(r)
		data.Title = "Pemesanan - SewaSini"
		data.Room = room
		data.Error = "Format tanggal mulai tidak valid"
		h.render(w, "room_detail", data)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		data := h.newTemplateData(r)
		data.Title = "Pemesanan - SewaSini"
		data.Room = room
		data.Error = "Format tanggal selesai tidak valid"
		h.render(w, "room_detail", data)
		return
	}

	duration := int(endDate.Sub(startDate).Hours()/24) + 1
	if duration < 1 {
		duration = 1
	}

	totalPrice := float64(duration) * room.PricePerDay

	session := h.getSession(r)
	userID := ""
	if session != nil {
		userID = session.UserID
	}

	booking := &models.Booking{
		RoomID:      roomID,
		UserID:      userID,
		RenterName:  renterName,
		RenterEmail: renterEmail,
		RenterPhone: renterPhone,
		StartDate:   startDate,
		EndDate:     endDate,
		Duration:    duration,
		TotalPrice:  totalPrice,
		Status:      "pending",
		Notes:       notes,
	}

	if err := h.store.CreateBooking(booking); err != nil {
		http.Error(w, "Failed to create booking", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/bookings/success?id="+booking.ID, http.StatusSeeOther)
}

func (h *Handler) BookingSuccess(w http.ResponseWriter, r *http.Request) {
	bookingID := r.URL.Query().Get("id")
	booking, err := h.store.GetBookingByID(bookingID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	room, _ := h.store.GetRoomByID(booking.RoomID)

	data := h.newTemplateData(r)
	data.Title = "Pemesanan Berhasil - SewaSini"
	data.Booking = booking
	data.Room = room
	h.render(w, "booking_success", data)
}

func (h *Handler) MyBookings(w http.ResponseWriter, r *http.Request) {
	session := h.getSession(r)
	if session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	bookings := h.store.GetBookingsByUser(session.UserID)
	data := h.newTemplateData(r)
	data.Title = "Pemesanan Saya - SewaSini"
	data.Bookings = bookings
	h.render(w, "my_bookings", data)
}
