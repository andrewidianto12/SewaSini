package booking

import (
	"context"
	"database/sql"
	"errors"

	"sewasini/models"
)

type SQLRepository struct {
	db *sql.DB
}

var ErrBookingNotFound = errors.New("booking not found")

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, booking *models.Booking) error {
	const query = `
		INSERT INTO bookings (
			user_id,
			ruangan_id,
			tanggal_mulai,
			tanggal_selesai,
			jumlah_peserta,
			total_harga,
			status,
			payment_status,
			booking_code
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text, created_at, updated_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		booking.UserID,
		booking.RuanganID,
		booking.TanggalMulai,
		booking.TanggalSelesai,
		booking.JumlahPeserta,
		booking.TotalHarga,
		booking.Status,
		booking.PaymentStatus,
		booking.BookingCode,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)
}

func (r *SQLRepository) HasActiveOverlap(ctx context.Context, ruanganID string, startDate, endDate string) (bool, error) {
	const query = `
		SELECT (
			COALESCE((
				SELECT COUNT(1)
				FROM bookings b
				WHERE b.ruangan_id::text = $1
					AND b.status IN ('pending', 'confirmed')
					AND b.tanggal_mulai < $3::timestamptz
					AND b.tanggal_selesai > $2::timestamptz
			), 0)
			>=
			COALESCE((
				SELECT r.stock_availability
				FROM ruangan r
				WHERE r.id::text = $1
			), 1)
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, ruanganID, startDate, endDate).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *SQLRepository) HasActiveOverlapExcluding(ctx context.Context, bookingID, ruanganID, startDate, endDate string) (bool, error) {
	const query = `
		SELECT (
			COALESCE((
				SELECT COUNT(1)
				FROM bookings b
				WHERE b.ruangan_id::text = $1
					AND b.id::text <> $2
					AND b.status IN ('pending', 'confirmed')
					AND b.tanggal_mulai < $4::timestamptz
					AND b.tanggal_selesai > $3::timestamptz
			), 0)
			>=
			COALESCE((
				SELECT r.stock_availability
				FROM ruangan r
				WHERE r.id::text = $1
			), 1)
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, ruanganID, bookingID, startDate, endDate).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*models.Booking, error) {
	const query = `
		SELECT
			id::text,
			user_id::text,
			ruangan_id::text,
			tanggal_mulai,
			tanggal_selesai,
			jumlah_peserta,
			total_harga,
			status,
			payment_status,
			booking_code,
			created_at,
			updated_at
		FROM bookings
		WHERE id::text = $1
	`

	booking := &models.Booking{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.RuanganID,
		&booking.TanggalMulai,
		&booking.TanggalSelesai,
		&booking.JumlahPeserta,
		&booking.TotalHarga,
		&booking.Status,
		&booking.PaymentStatus,
		&booking.BookingCode,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	return booking, nil
}

func (r *SQLRepository) ListByUser(ctx context.Context, userID string) ([]models.Booking, error) {
	const query = `
		SELECT
			id::text,
			user_id::text,
			ruangan_id::text,
			tanggal_mulai,
			tanggal_selesai,
			jumlah_peserta,
			total_harga,
			status,
			payment_status,
			booking_code,
			created_at,
			updated_at
		FROM bookings
		WHERE user_id::text = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]models.Booking, 0)
	for rows.Next() {
		var booking models.Booking
		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.RuanganID,
			&booking.TanggalMulai,
			&booking.TanggalSelesai,
			&booking.JumlahPeserta,
			&booking.TotalHarga,
			&booking.Status,
			&booking.PaymentStatus,
			&booking.BookingCode,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *SQLRepository) ListAll(ctx context.Context) ([]models.Booking, error) {
	const query = `
		SELECT
			id::text,
			user_id::text,
			ruangan_id::text,
			tanggal_mulai,
			tanggal_selesai,
			jumlah_peserta,
			total_harga,
			status,
			payment_status,
			booking_code,
			created_at,
			updated_at
		FROM bookings
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]models.Booking, 0)
	for rows.Next() {
		var booking models.Booking
		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.RuanganID,
			&booking.TanggalMulai,
			&booking.TanggalSelesai,
			&booking.JumlahPeserta,
			&booking.TotalHarga,
			&booking.Status,
			&booking.PaymentStatus,
			&booking.BookingCode,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *SQLRepository) Update(ctx context.Context, booking *models.Booking) error {
	const query = `
		UPDATE bookings
		SET tanggal_mulai = $2,
			tanggal_selesai = $3,
			jumlah_peserta = $4,
			total_harga = $5,
			status = $6,
			payment_status = $7,
			updated_at = NOW()
		WHERE id::text = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		booking.ID,
		booking.TanggalMulai,
		booking.TanggalSelesai,
		booking.JumlahPeserta,
		booking.TotalHarga,
		booking.Status,
		booking.PaymentStatus,
	).Scan(&booking.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBookingNotFound
		}
		return err
	}

	return nil
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, booking *models.Booking) error {
	const query = `
		UPDATE bookings
		SET status = $2,
			payment_status = $3,
			updated_at = NOW()
		WHERE id::text = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query, booking.ID, booking.Status, booking.PaymentStatus).Scan(&booking.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBookingNotFound
		}
		return err
	}

	return nil
}

func (r *SQLRepository) Cancel(ctx context.Context, id string) error {
	const query = `
		UPDATE bookings
		SET status = 'cancelled',
			updated_at = NOW()
		WHERE id::text = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *SQLRepository) MarkPaidAndConfirmed(ctx context.Context, bookingID string) error {
	const query = `
		UPDATE bookings
		SET payment_status = 'paid',
			status = 'confirmed',
			updated_at = NOW()
		WHERE id::text = $1
	`

	result, err := r.db.ExecContext(ctx, query, bookingID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}
