package models

import (
	"time"
)

type TransactionStatus string

const (
	TransactionPending  TransactionStatus = "pending"
	TransactionSuccess  TransactionStatus = "success"
	TransactionFailed   TransactionStatus = "failed"
	TransactionExpired  TransactionStatus = "expired"
)

type Transaction struct {
	ID              string            `json:"id" db:"id"`
	BookingID       string            `json:"booking_id" db:"booking_id"`
	UserID          string            `json:"user_id" db:"user_id"`
	Amount          int64             `json:"amount" db:"amount"`
	PaymentMethod   string            `json:"payment_method" db:"payment_method"`
	TransactionDate time.Time         `json:"transaction_date" db:"transaction_date"`
	Status          TransactionStatus `json:"status" db:"status"`
	XenditID        string            `json:"xendit_id" db:"xendit_id"`
	PaymentURL      string            `json:"payment_url" db:"payment_url"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}

type CreateTransactionRequest struct {
	BookingID     string `json:"booking_id" validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required"`
}

type TransactionResponse struct {
	ID              string            `json:"id"`
	BookingID       string            `json:"booking_id"`
	Amount          int64             `json:"amount"`
	PaymentMethod   string            `json:"payment_method"`
	TransactionDate time.Time         `json:"transaction_date"`
	Status          TransactionStatus `json:"status"`
	PaymentURL      string            `json:"payment_url,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type XenditCallbackRequest struct {
	ExternalID string `json:"external_id"`
	Status     string `json:"status"`
}
