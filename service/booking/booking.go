package booking

import (
	"context"

	"sewasini/models"
)

type Repository interface {
	Create(ctx context.Context, booking *models.Booking) error
	HasActiveOverlap(ctx context.Context, ruanganID string, startDate, endDate string) (bool, error)
	HasActiveOverlapExcluding(ctx context.Context, bookingID, ruanganID, startDate, endDate string) (bool, error)
	GetByID(ctx context.Context, id string) (*models.Booking, error)
	ListByUser(ctx context.Context, userID string) ([]models.Booking, error)
	ListAll(ctx context.Context) ([]models.Booking, error)
	Update(ctx context.Context, booking *models.Booking) error
	UpdateStatus(ctx context.Context, booking *models.Booking) error
	Cancel(ctx context.Context, id string) error
}

type RoomRepository interface {
	GetByID(ctx context.Context, id string) (*models.RuanganResponse, error)
}

type Service interface {
	CreateBooking(ctx context.Context, userID string, req models.CreateBookingRequest) (*models.BookingResponse, error)
	ListUserBookings(ctx context.Context, userID string) ([]models.BookingResponse, error)
	GetUserBookingByID(ctx context.Context, userID, bookingID string) (*models.BookingResponse, error)
	UpdateBooking(ctx context.Context, userID, bookingID string, req models.UpdateBookingRequest) (*models.BookingResponse, error)
	CancelBooking(ctx context.Context, userID, bookingID string) error
	AdminListBookings(ctx context.Context) ([]models.BookingResponse, error)
	AdminGetBookingByID(ctx context.Context, bookingID string) (*models.BookingResponse, error)
	AdminUpdateBooking(ctx context.Context, bookingID string, req models.AdminUpdateBookingRequest) (*models.BookingResponse, error)
}
