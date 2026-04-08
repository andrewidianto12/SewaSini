package review

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"sewasini/models"
)

var ErrReviewNotFound = errors.New("review not found")
var ErrBookingNotFound = errors.New("booking not found")
var ErrReviewAlreadyExists = errors.New("review already exists")

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, review *models.Review) error {
	const query = `
		INSERT INTO reviews (
			user_id,
			ruangan_id,
			booking_id,
			rating,
			komentar
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, created_at
	`

	if err := r.db.QueryRowContext(
		ctx,
		query,
		review.UserID,
		review.RuanganID,
		review.BookingID,
		review.Rating,
		review.Komentar,
	).Scan(&review.ID, &review.CreatedAt); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrReviewAlreadyExists
		}
		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*models.ReviewResponse, error) {
	const query = `
		SELECT
			r.id::text,
			r.booking_id::text,
			r.user_id::text,
			u.nama_lengkap,
			r.ruangan_id::text,
			r.rating,
			COALESCE(r.komentar, ''),
			r.created_at
		FROM reviews r
		JOIN users u ON u.id::text = r.user_id::text
		WHERE r.id::text = $1
	`

	review := &models.ReviewResponse{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.BookingID,
		&review.UserID,
		&review.UserName,
		&review.RuanganID,
		&review.Rating,
		&review.Komentar,
		&review.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}

	return review, nil
}

func (r *SQLRepository) GetByUserAndBooking(ctx context.Context, userID, bookingID string) (*models.Review, error) {
	const query = `
		SELECT
			r.id::text,
			r.user_id::text,
			r.ruangan_id::text,
			r.booking_id::text,
			r.rating,
			COALESCE(r.komentar, ''),
			r.created_at
		FROM reviews r
		WHERE r.user_id::text = $1 AND r.booking_id::text = $2
	`

	review := &models.Review{}
	err := r.db.QueryRowContext(ctx, query, userID, bookingID).Scan(
		&review.ID,
		&review.UserID,
		&review.RuanganID,
		&review.BookingID,
		&review.Rating,
		&review.Komentar,
		&review.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}

	return review, nil
}

func (r *SQLRepository) ListByUser(ctx context.Context, userID string) ([]models.ReviewResponse, error) {
	const query = `
		SELECT
			r.id::text,
			r.booking_id::text,
			r.user_id::text,
			u.nama_lengkap,
			r.ruangan_id::text,
			r.rating,
			COALESCE(r.komentar, ''),
			r.created_at
		FROM reviews r
		JOIN users u ON u.id::text = r.user_id::text
		WHERE r.user_id::text = $1
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]models.ReviewResponse, 0)
	for rows.Next() {
		var review models.ReviewResponse
		if err := rows.Scan(
			&review.ID,
			&review.BookingID,
			&review.UserID,
			&review.UserName,
			&review.RuanganID,
			&review.Rating,
			&review.Komentar,
			&review.CreatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *SQLRepository) Update(ctx context.Context, reviewID string, rating int, komentar string) error {
	const query = `
		UPDATE reviews
		SET rating = $2,
			komentar = $3
		WHERE id::text = $1
	`

	result, err := r.db.ExecContext(ctx, query, reviewID, rating, komentar)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrReviewNotFound
	}

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM reviews WHERE id::text = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrReviewNotFound
	}

	return nil
}

func (r *SQLRepository) GetBookingByID(ctx context.Context, bookingID string) (*models.Booking, error) {
	const query = `
		SELECT
			id::text,
			user_id::text,
			ruangan_id::text,
			status,
			payment_status
		FROM bookings
		WHERE id::text = $1
	`

	var booking models.Booking
	var status string
	var paymentStatus string
	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.RuanganID,
		&status,
		&paymentStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	booking.Status = models.BookingStatus(status)
	booking.PaymentStatus = models.PaymentStatus(paymentStatus)
	return &booking, nil
}
