package booking

import (
	"context"
	"database/sql"

	"sewasini/models"
)

type SQLRepository struct {
	db *sql.DB
}

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
		RETURNING id, created_at, updated_at
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
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE ruangan_id = $1
				AND status IN ('pending', 'confirmed')
				AND tanggal_mulai < $3::timestamp
				AND tanggal_selesai > $2::timestamp
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, ruanganID, startDate, endDate).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
