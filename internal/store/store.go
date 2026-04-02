package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/andrewidianto12/SewaSini/internal/models"
)

type Store struct {
	mu        sync.RWMutex
	rooms     map[string]*models.Room
	users     map[string]*models.User
	bookings  map[string]*models.Booking
	roomOrder []string
}

func NewStore() *Store {
	s := &Store{
		rooms:    make(map[string]*models.Room),
		users:    make(map[string]*models.User),
		bookings: make(map[string]*models.Booking),
	}
	s.seedRooms()
	return s
}

func (s *Store) seedRooms() {
	rooms := []*models.Room{
		{
			ID:           "1",
			Name:         "Ruang Meeting Anggrek",
			Description:  "Ruang meeting modern yang nyaman untuk rapat bisnis kecil hingga menengah. Dilengkapi dengan peralatan presentasi lengkap dan koneksi internet cepat.",
			Capacity:     10,
			PricePerHour: 150000,
			PricePerDay:  1000000,
			Location:     "Jakarta Selatan",
			Type:         "meeting_room",
			Amenities:    []string{"WiFi", "Proyektor", "AC", "Whiteboard"},
			ImageURL:     "https://placehold.co/800x400/2563eb/white?text=Ruang+Meeting+Anggrek",
			IsAvailable:  true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           "2",
			Name:         "Aula Serbaguna Melati",
			Description:  "Aula besar yang cocok untuk acara pernikahan, seminar, dan event besar lainnya. Memiliki panggung dan sistem audio yang lengkap.",
			Capacity:     200,
			PricePerHour: 500000,
			PricePerDay:  3000000,
			Location:     "Jakarta Pusat",
			Type:         "event_hall",
			Amenities:    []string{"Sound System", "AC", "Panggung", "Parkir"},
			ImageURL:     "https://placehold.co/800x400/7c3aed/white?text=Aula+Serbaguna+Melati",
			IsAvailable:  true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           "3",
			Name:         "Coworking Space Kaktus",
			Description:  "Ruang kerja bersama yang inspiratif dan produktif. Cocok untuk freelancer, startup, dan remote worker yang membutuhkan suasana kerja yang dinamis.",
			Capacity:     50,
			PricePerHour: 50000,
			PricePerDay:  300000,
			Location:     "Bandung",
			Type:         "coworking",
			Amenities:    []string{"WiFi Cepat", "Kopi Gratis", "Printer", "Loker"},
			ImageURL:     "https://placehold.co/800x400/059669/white?text=Coworking+Kaktus",
			IsAvailable:  true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           "4",
			Name:         "Ruang Meeting Mawar",
			Description:  "Ruang meeting premium dengan kapasitas lebih besar, ideal untuk presentasi klien dan rapat manajemen.",
			Capacity:     20,
			PricePerHour: 250000,
			PricePerDay:  1500000,
			Location:     "Surabaya",
			Type:         "meeting_room",
			Amenities:    []string{"WiFi", "TV LED", "AC", "Meja Besar"},
			ImageURL:     "https://placehold.co/800x400/dc2626/white?text=Ruang+Meeting+Mawar",
			IsAvailable:  true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           "5",
			Name:         "Aula Konferensi Lotus",
			Description:  "Aula konferensi profesional yang ideal untuk seminar, konferensi, dan acara korporat.",
			Capacity:     100,
			PricePerHour: 350000,
			PricePerDay:  2000000,
			Location:     "Jakarta Barat",
			Type:         "event_hall",
			Amenities:    []string{"Proyektor", "Sound System", "AC", "Catering"},
			ImageURL:     "https://placehold.co/800x400/d97706/white?text=Aula+Konferensi+Lotus",
			IsAvailable:  true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           "6",
			Name:         "Ruang Kreatif Dahlia",
			Description:  "Ruang coworking yang unik dengan desain kreatif, sempurna untuk tim kreatif, desainer, dan content creator.",
			Capacity:     30,
			PricePerHour: 75000,
			PricePerDay:  450000,
			Location:     "Yogyakarta",
			Type:         "coworking",
			Amenities:    []string{"WiFi", "AC", "Meja Standing", "Ruang Santai"},
			ImageURL:     "https://placehold.co/800x400/db2777/white?text=Ruang+Kreatif+Dahlia",
			IsAvailable:  true,
			CreatedAt:    time.Now(),
		},
	}
	for _, r := range rooms {
		s.rooms[r.ID] = r
		s.roomOrder = append(s.roomOrder, r.ID)
	}
}

func (s *Store) GetAllRooms() []*models.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rooms := make([]*models.Room, 0, len(s.roomOrder))
	for _, id := range s.roomOrder {
		if r, ok := s.rooms[id]; ok {
			rooms = append(rooms, r)
		}
	}
	return rooms
}

func (s *Store) GetRoomByID(id string) (*models.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[id]
	if !ok {
		return nil, errors.New("room not found")
	}
	return r, nil
}

func (s *Store) CreateBooking(b *models.Booking) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b.ID == "" {
		b.ID = fmt.Sprintf("BK%d", time.Now().UnixNano())
	}
	b.CreatedAt = time.Now()
	s.bookings[b.ID] = b
	return nil
}

func (s *Store) GetBookingsByUser(userID string) []*models.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var bookings []*models.Booking
	for _, b := range s.bookings {
		if b.UserID == userID {
			bookings = append(bookings, b)
		}
	}
	return bookings
}

func (s *Store) GetAllBookings() []*models.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bookings := make([]*models.Booking, 0, len(s.bookings))
	for _, b := range s.bookings {
		bookings = append(bookings, b)
	}
	return bookings
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *Store) CreateUser(u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u.ID == "" {
		u.ID = fmt.Sprintf("U%d", time.Now().UnixNano())
	}
	u.CreatedAt = time.Now()
	s.users[u.ID] = u
	return nil
}

func (s *Store) GetBookingByID(id string) (*models.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bookings[id]
	if !ok {
		return nil, errors.New("booking not found")
	}
	return b, nil
}
