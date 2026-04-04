package models

import (
	"time"
)

type Review struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id" validate:"required"`
	RuanganID  string    `json:"ruangan_id" db:"ruangan_id" validate:"required"`
	BookingID  string    `json:"booking_id" db:"booking_id" validate:"required"`
	Rating     int       `json:"rating" db:"rating" validate:"required,min=1,max=5"`
	Komentar   string    `json:"komentar" db:"komentar"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CreateReviewRequest struct {
	RuanganID string `json:"ruangan_id" validate:"required"`
	BookingID string `json:"booking_id" validate:"required"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
	Komentar  string `json:"komentar"`
}

type UpdateReviewRequest struct {
	Rating   int    `json:"rating" validate:"min=1,max=5"`
	Komentar string `json:"komentar"`
}

type ReviewResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	RuanganID string    `json:"ruangan_id"`
	Rating    int       `json:"rating"`
	Komentar  string    `json:"komentar"`
	CreatedAt time.Time `json:"created_at"`
}
