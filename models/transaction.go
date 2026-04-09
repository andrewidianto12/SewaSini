package models

import (
	"time"
)

type TransactionStatus string

const (
	TransactionPending TransactionStatus = "pending"
	TransactionSuccess TransactionStatus = "success"
	TransactionFailed  TransactionStatus = "failed"
	TransactionExpired TransactionStatus = "expired"
)

type Transaction struct {
	ID              string            `json:"id" db:"id"`
	BookingID       string            `json:"booking_id" db:"booking_id"`
	UserID          string            `json:"user_id" db:"user_id"`
	Amount          int64             `json:"amount" db:"amount"`
	PaymentMethod   string            `json:"payment_method" db:"payment_method"`
	TransactionDate time.Time         `json:"transaction_date" db:"transaction_date"`
	Status          TransactionStatus `json:"status" db:"status"`
	ExternalID      string            `json:"external_id" db:"external_id"`
	XenditID        string            `json:"xendit_id" db:"xendit_id"`
	LastWebhookID   string            `json:"-" db:"last_webhook_id"`
	PaymentURL      string            `json:"payment_url" db:"payment_url"`
	EmailSentAt     time.Time         `json:"-" db:"email_sent_at"`
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
	ExternalID      string            `json:"external_id"`
	PaymentURL      string            `json:"payment_url,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type XenditCallbackRequest struct {
	ID            string `json:"id"`
	ExternalID    string `json:"external_id"`
	ExternalIDAlt string `json:"externalId"`
	InvoiceID     string `json:"invoice_id"`
	Status        string `json:"status"`
	WebhookID     string `json:"-"`
}

type RevenueAnalyticsResponse struct {
	TotalRevenue           int64 `json:"total_revenue"`
	SuccessfulTransactions int   `json:"successful_transactions"`
	PendingTransactions    int   `json:"pending_transactions"`
	FailedTransactions     int   `json:"failed_transactions"`
	ExpiredTransactions    int   `json:"expired_transactions"`
}

type ReportResponse struct {
	TotalUsers        int   `json:"total_users"`
	TotalRooms        int   `json:"total_rooms"`
	TotalBookings     int   `json:"total_bookings"`
	TotalTransactions int   `json:"total_transactions"`
	TotalRevenue      int64 `json:"total_revenue"`
}

type DashboardResponse struct {
	TotalUsers         int   `json:"total_users"`
	TotalRooms         int   `json:"total_rooms"`
	ActiveRooms        int   `json:"active_rooms"`
	TotalBookings      int   `json:"total_bookings"`
	PendingBookings    int   `json:"pending_bookings"`
	ConfirmedBookings  int   `json:"confirmed_bookings"`
	CancelledBookings  int   `json:"cancelled_bookings"`
	TotalTransactions  int   `json:"total_transactions"`
	SuccessfulPayments int   `json:"successful_payments"`
	PendingPayments    int   `json:"pending_payments"`
	FailedPayments     int   `json:"failed_payments"`
	ExpiredPayments    int   `json:"expired_payments"`
	TotalRevenue       int64 `json:"total_revenue"`
}
