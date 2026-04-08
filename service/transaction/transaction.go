package transaction

import (
	"context"

	"sewasini/models"
)

type Service interface {
	CreatePayment(ctx context.Context, userID string, req models.CreateTransactionRequest) (*models.TransactionResponse, error)
	HandleCallback(ctx context.Context, req models.XenditCallbackRequest) error
}
