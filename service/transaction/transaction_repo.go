package transaction

import (
	"context"

	"sewasini/models"
)

type Repository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	GetByID(ctx context.Context, id string) (*models.Transaction, error)
	ListAll(ctx context.Context) ([]models.Transaction, error)
	GetByExternalID(ctx context.Context, externalID string) (*models.Transaction, error)
	UpdateStatusByExternalID(ctx context.Context, externalID string, status models.TransactionStatus, xenditID, webhookID string) error
	MarkSuccessAndConfirmBooking(ctx context.Context, externalID, xenditID, webhookID string) error
	MarkEmailSent(ctx context.Context, externalID string) error
	GetRevenueAnalytics(ctx context.Context) (*models.RevenueAnalyticsResponse, error)
	GetReports(ctx context.Context) (*models.ReportResponse, error)
	GetDashboard(ctx context.Context) (*models.DashboardResponse, error)
}

type BookingRepository interface {
	GetByID(ctx context.Context, id string) (*models.Booking, error)
	MarkPaidAndConfirmed(ctx context.Context, bookingID string) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
}
