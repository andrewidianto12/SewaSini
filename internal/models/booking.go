package models

import "time"

type Booking struct {
	ID          string
	RoomID      string
	UserID      string
	RenterName  string
	RenterEmail string
	RenterPhone string
	StartDate   time.Time
	EndDate     time.Time
	Duration    int // days
	TotalPrice  float64
	Status      string // pending, confirmed, cancelled
	Notes       string
	CreatedAt   time.Time
}
