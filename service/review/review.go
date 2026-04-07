package review

import (
	"context"
	"errors"
	"strings"

	"sewasini/models"
)

var ErrUserIDRequired = errors.New("user id is required")
var ErrForbiddenReviewAccess = errors.New("forbidden review access")
var ErrBookingMismatch = errors.New("booking does not match room")
var ErrReviewUpdateEmpty = errors.New("at least one field must be provided")
var ErrInsufficientRole = errors.New("only users can create reviews")

type Repository interface {
	Create(ctx context.Context, review *models.Review) error
	GetByID(ctx context.Context, id string) (*models.ReviewResponse, error)
	GetByUserAndBooking(ctx context.Context, userID, bookingID string) (*models.Review, error)
	ListByUser(ctx context.Context, userID string) ([]models.ReviewResponse, error)
	Update(ctx context.Context, reviewID string, rating int, komentar string) error
	Delete(ctx context.Context, id string) error
	GetBookingByID(ctx context.Context, bookingID string) (*models.Booking, error)
}

type Service interface {
	CreateReview(ctx context.Context, userID, userRole string, req models.CreateReviewRequest) (*models.ReviewResponse, error)
	ListReviews(ctx context.Context, userID string) ([]models.ReviewResponse, error)
	GetReviewByID(ctx context.Context, userID, id string) (*models.ReviewResponse, error)
	UpdateReview(ctx context.Context, userID, id string, req models.UpdateReviewRequest) (*models.ReviewResponse, error)
	DeleteReview(ctx context.Context, userID, id string) error
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}
