package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andrewidianto12/SewaSini/internal/handlers"
	"github.com/andrewidianto12/SewaSini/internal/models"
	"github.com/andrewidianto12/SewaSini/internal/store"
)

func newTestHandler() *handlers.Handler {
	funcMap := template.FuncMap{
		"formatCurrency": formatCurrency,
		"formatDate":     formatDate,
		"add":            func(a, b int) int { return a + b },
		"typeLabel":      typeLabel,
		"contains":       strings.Contains,
	}
	templates := loadTemplates(funcMap)
	s := store.NewStore()
	return handlers.NewHandler(s, templates)
}

func TestStoreGetAllRooms(t *testing.T) {
	s := store.NewStore()
	rooms := s.GetAllRooms()
	if len(rooms) != 6 {
		t.Errorf("expected 6 rooms, got %d", len(rooms))
	}
}

func TestStoreCreateBooking(t *testing.T) {
	s := store.NewStore()
	booking := &models.Booking{
		RoomID:      "1",
		RenterName:  "Test User",
		RenterEmail: "test@example.com",
		RenterPhone: "081234567890",
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 1),
		Duration:    1,
		TotalPrice:  1000000,
		Status:      "pending",
	}
	if err := s.CreateBooking(booking); err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	if booking.ID == "" {
		t.Error("booking ID should not be empty after creation")
	}
	fetched, err := s.GetBookingByID(booking.ID)
	if err != nil {
		t.Fatalf("failed to get booking by ID: %v", err)
	}
	if fetched.RenterName != "Test User" {
		t.Errorf("expected renter name 'Test User', got '%s'", fetched.RenterName)
	}
}

func TestStoreGetBookingsByUser(t *testing.T) {
	s := store.NewStore()
	booking := &models.Booking{
		RoomID:      "1",
		UserID:      "user123",
		RenterName:  "Test User",
		RenterEmail: "test@example.com",
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 1),
		Duration:    1,
		TotalPrice:  1000000,
		Status:      "pending",
	}
	s.CreateBooking(booking)
	bookings := s.GetBookingsByUser("user123")
	if len(bookings) != 1 {
		t.Errorf("expected 1 booking, got %d", len(bookings))
	}
}

func TestStoreCreateUser(t *testing.T) {
	s := store.NewStore()
	user := &models.User{
		Name:  "Test User",
		Email: "test@example.com",
		Phone: "081234567890",
	}
	if err := s.CreateUser(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	fetched, err := s.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}
	if fetched.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", fetched.Name)
	}
}

func TestHomePageReturns200(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	h.Home(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestRoomsPageReturns200(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/rooms", nil)
	rr := httptest.NewRecorder()
	h.GetRooms(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestNonExistentRoomReturns404(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/rooms/nonexistent", nil)
	rr := httptest.NewRecorder()
	req.SetPathValue("id", "nonexistent")
	h.GetRoom(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		amount   float64
		expected string
	}{
		{150000, "Rp 150.000"},
		{1000000, "Rp 1.000.000"},
		{50000, "Rp 50.000"},
	}
	for _, tt := range tests {
		result := formatCurrency(tt.amount)
		if result != tt.expected {
			t.Errorf("formatCurrency(%v) = %q, want %q", tt.amount, result, tt.expected)
		}
	}
}

func TestFormatDate(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	result := formatDate(testTime)
	if result != "15 Jan 2024" {
		t.Errorf("formatDate() = %q, want %q", result, "15 Jan 2024")
	}
}
