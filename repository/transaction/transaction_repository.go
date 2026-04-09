package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"sewasini/models"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, tx *models.Transaction) error {
	const query = `
		INSERT INTO transactions (
			booking_id,
			user_id,
			amount,
			payment_method,
			transaction_date,
			status,
			external_id,
			xendit_id,
			payment_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text, created_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		tx.BookingID,
		tx.UserID,
		tx.Amount,
		tx.PaymentMethod,
		tx.TransactionDate,
		tx.Status,
		tx.ExternalID,
		tx.XenditID,
		tx.PaymentURL,
	).Scan(&tx.ID, &tx.CreatedAt)
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*models.Transaction, error) {
	const query = `
		SELECT
			id::text,
			booking_id::text,
			user_id::text,
			amount,
			payment_method,
			transaction_date,
			status,
			external_id,
			COALESCE(xendit_id, ''),
			COALESCE(last_webhook_id, ''),
			COALESCE(payment_url, ''),
			email_sent_at,
			created_at
		FROM transactions
		WHERE id::text = $1
	`

	tx := &models.Transaction{}
	var emailSentAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.BookingID,
		&tx.UserID,
		&tx.Amount,
		&tx.PaymentMethod,
		&tx.TransactionDate,
		&tx.Status,
		&tx.ExternalID,
		&tx.XenditID,
		&tx.LastWebhookID,
		&tx.PaymentURL,
		&emailSentAt,
		&tx.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	if emailSentAt.Valid {
		tx.EmailSentAt = emailSentAt.Time
	}

	return tx, nil
}

func (r *SQLRepository) ListAll(ctx context.Context) ([]models.Transaction, error) {
	const query = `
		SELECT
			id::text,
			booking_id::text,
			user_id::text,
			amount,
			payment_method,
			transaction_date,
			status,
			external_id,
			COALESCE(xendit_id, ''),
			COALESCE(last_webhook_id, ''),
			COALESCE(payment_url, ''),
			email_sent_at,
			created_at
		FROM transactions
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs := make([]models.Transaction, 0)
	for rows.Next() {
		var tx models.Transaction
		var emailSentAt sql.NullTime
		if err := rows.Scan(
			&tx.ID,
			&tx.BookingID,
			&tx.UserID,
			&tx.Amount,
			&tx.PaymentMethod,
			&tx.TransactionDate,
			&tx.Status,
			&tx.ExternalID,
			&tx.XenditID,
			&tx.LastWebhookID,
			&tx.PaymentURL,
			&emailSentAt,
			&tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		if emailSentAt.Valid {
			tx.EmailSentAt = emailSentAt.Time
		}
		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return txs, nil
}

func (r *SQLRepository) GetByExternalID(ctx context.Context, externalID string) (*models.Transaction, error) {
	const query = `
		SELECT
			id::text,
			booking_id::text,
			user_id::text,
			amount,
			payment_method,
			transaction_date,
			status,
			external_id,
			COALESCE(xendit_id, ''),
			COALESCE(last_webhook_id, ''),
			COALESCE(payment_url, ''),
			email_sent_at,
			created_at
		FROM transactions
		WHERE external_id = $1
	`

	tx := &models.Transaction{}
	var emailSentAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, externalID).Scan(
		&tx.ID,
		&tx.BookingID,
		&tx.UserID,
		&tx.Amount,
		&tx.PaymentMethod,
		&tx.TransactionDate,
		&tx.Status,
		&tx.ExternalID,
		&tx.XenditID,
		&tx.LastWebhookID,
		&tx.PaymentURL,
		&emailSentAt,
		&tx.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	if emailSentAt.Valid {
		tx.EmailSentAt = emailSentAt.Time
	}

	return tx, nil
}

func (r *SQLRepository) UpdateStatusByExternalID(ctx context.Context, externalID string, status models.TransactionStatus, xenditID, webhookID string) error {
	const query = `
		UPDATE transactions
		SET status = $2,
			xendit_id = CASE WHEN $3 = '' THEN xendit_id ELSE $3 END,
			last_webhook_id = CASE WHEN $4 = '' THEN last_webhook_id ELSE $4 END
		WHERE external_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, externalID, status, strings.TrimSpace(xenditID), strings.TrimSpace(webhookID))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

func (r *SQLRepository) MarkSuccessAndConfirmBooking(ctx context.Context, externalID, xenditID, webhookID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var bookingID string
	err = tx.QueryRowContext(ctx, `SELECT booking_id::text FROM transactions WHERE external_id = $1 FOR UPDATE`, externalID).Scan(&bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTransactionNotFound
		}
		return err
	}

	result, err := tx.ExecContext(
		ctx,
		`
			UPDATE transactions
			SET status = 'success',
				xendit_id = CASE WHEN $2 = '' THEN xendit_id ELSE $2 END,
				last_webhook_id = CASE WHEN $3 = '' THEN last_webhook_id ELSE $3 END
			WHERE external_id = $1
		`,
		externalID,
		strings.TrimSpace(xenditID),
		strings.TrimSpace(webhookID),
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTransactionNotFound
	}

	result, err = tx.ExecContext(
		ctx,
		`
			UPDATE bookings
			SET payment_status = 'paid',
				status = 'confirmed',
				updated_at = NOW()
			WHERE id::text = $1
		`,
		bookingID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("confirm booking for transaction %s: %w", externalID, ErrTransactionNotFound)
	}

	err = tx.Commit()
	return err
}

func (r *SQLRepository) MarkEmailSent(ctx context.Context, externalID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE transactions SET email_sent_at = NOW() WHERE external_id = $1`, externalID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

func (r *SQLRepository) GetRevenueAnalytics(ctx context.Context) (*models.RevenueAnalyticsResponse, error) {
	const query = `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'success' THEN amount ELSE 0 END), 0),
			COUNT(*) FILTER (WHERE status = 'success'),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'failed'),
			COUNT(*) FILTER (WHERE status = 'expired')
		FROM transactions
	`

	resp := &models.RevenueAnalyticsResponse{}
	if err := r.db.QueryRowContext(ctx, query).Scan(
		&resp.TotalRevenue,
		&resp.SuccessfulTransactions,
		&resp.PendingTransactions,
		&resp.FailedTransactions,
		&resp.ExpiredTransactions,
	); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *SQLRepository) GetReports(ctx context.Context) (*models.ReportResponse, error) {
	resp := &models.ReportResponse{}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&resp.TotalUsers); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ruangan`).Scan(&resp.TotalRooms); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&resp.TotalBookings); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&resp.TotalTransactions); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE status = 'success'`).Scan(&resp.TotalRevenue); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *SQLRepository) GetDashboard(ctx context.Context) (*models.DashboardResponse, error) {
	resp := &models.DashboardResponse{}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&resp.TotalUsers); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ruangan`).Scan(&resp.TotalRooms); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ruangan WHERE is_active = TRUE`).Scan(&resp.ActiveRooms); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'confirmed'),
			COUNT(*) FILTER (WHERE status = 'cancelled')
		FROM bookings
	`).Scan(&resp.TotalBookings, &resp.PendingBookings, &resp.ConfirmedBookings, &resp.CancelledBookings); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'success'),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'failed'),
			COUNT(*) FILTER (WHERE status = 'expired'),
			COALESCE(SUM(CASE WHEN status = 'success' THEN amount ELSE 0 END), 0)
		FROM transactions
	`).Scan(
		&resp.TotalTransactions,
		&resp.SuccessfulPayments,
		&resp.PendingPayments,
		&resp.FailedPayments,
		&resp.ExpiredPayments,
		&resp.TotalRevenue,
	); err != nil {
		return nil, err
	}
	return resp, nil
}
