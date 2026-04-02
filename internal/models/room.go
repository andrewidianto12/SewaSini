package models

import "time"

type Room struct {
	ID           string
	Name         string
	Description  string
	Capacity     int
	PricePerHour float64
	PricePerDay  float64
	Location     string
	Type         string // meeting_room, event_hall, coworking
	Amenities    []string
	ImageURL     string
	IsAvailable  bool
	CreatedAt    time.Time
}
