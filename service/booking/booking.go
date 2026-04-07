package booking

import (
	"context"

	"sewasini/models"
)

type Repository interface {
	Create(ctx context.Context, booking *models.Booking) error
	HasActiveOverlap(ctx context.Context, ruanganID string, startDate, endDate string) (bool, error)
}

type RoomRepository interface {
	GetByID(ctx context.Context, id string) (*models.RuanganResponse, error)
}

type Service interface {
	CreateBooking(ctx context.Context, userID string, req models.CreateBookingRequest) (*models.BookingResponse, error)
}
