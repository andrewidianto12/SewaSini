package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/andrewidianto12/SewaSini/internal/handlers"
	"github.com/andrewidianto12/SewaSini/internal/store"
)

func formatCurrency(amount float64) string {
	intAmount := int(math.Round(amount))
	str := fmt.Sprintf("%d", intAmount)
	result := ""
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return "Rp " + result
}

func formatDate(t interface{}) string {
	if v, ok := t.(interface{ Format(string) string }); ok {
		return v.Format("02 Jan 2006")
	}
	return ""
}

func typeLabel(roomType string) string {
	switch roomType {
	case "meeting_room":
		return "Ruang Meeting"
	case "event_hall":
		return "Aula Event"
	case "coworking":
		return "Coworking"
	default:
		return roomType
	}
}

func loadTemplates(funcMap template.FuncMap) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	pages := []string{"index", "rooms", "room_detail", "booking_form", "booking_success", "login", "register", "my_bookings"}
	for _, page := range pages {
		t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles(
			"templates/base.html",
			fmt.Sprintf("templates/%s.html", page),
		))
		templates[page] = t
	}
	return templates
}

func main() {
	funcMap := template.FuncMap{
		"formatCurrency": formatCurrency,
		"formatDate":     formatDate,
		"add":            func(a, b int) int { return a + b },
		"typeLabel":      typeLabel,
		"contains":       strings.Contains,
	}

	templates := loadTemplates(funcMap)
	s := store.NewStore()
	h := handlers.NewHandler(s, templates)

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /{$}", h.Home)
	mux.HandleFunc("GET /rooms", h.GetRooms)
	mux.HandleFunc("GET /rooms/{id}", h.GetRoom)
	mux.HandleFunc("POST /bookings", h.CreateBooking)
	mux.HandleFunc("GET /bookings/success", h.BookingSuccess)
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("GET /register", h.RegisterPage)
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("GET /logout", h.Logout)
	mux.HandleFunc("GET /my-bookings", h.MyBookings)

	log.Println("SewaSini server berjalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
