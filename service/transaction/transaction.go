package transaction

import (
	"context"

	"sewasini/models"
)

type Service interface {
	CreatePayment(ctx context.Context, userID string, req models.CreateTransactionRequest) (*models.TransactionResponse, error)
	GetPayment(ctx context.Context, userID, paymentID string) (*models.TransactionResponse, error)
	GetInvoice(ctx context.Context, userID, paymentID string) (*models.TransactionResponse, error)
	HandleCallback(ctx context.Context, req models.XenditCallbackRequest) error
	AdminListTransactions(ctx context.Context) ([]models.TransactionResponse, error)
	AdminReports(ctx context.Context) (*models.ReportResponse, error)
	AdminRevenues(ctx context.Context) (*models.RevenueAnalyticsResponse, error)
	AdminDashboard(ctx context.Context) (*models.DashboardResponse, error)
}
