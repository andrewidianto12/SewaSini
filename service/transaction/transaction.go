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
}
