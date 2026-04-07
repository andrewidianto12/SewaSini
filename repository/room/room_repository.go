package room

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sewasini/models"
)

var ErrRoomNotFound = errors.New("room not found")

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) List(ctx context.Context, filter models.RuanganFilter) ([]models.RuanganResponse, error) {
	baseQuery := `
		SELECT
			r.id::text,
			r.nama_ruangan,
			r.kategori_id::text,
			k.nama_kategori,
			COALESCE(r.deskripsi, ''),
			r.alamat,
			r.kota,
			r.kapasitas,
			r.harga_per_jam,
			r.harga_per_hari,
			r.stock_availability,
			COALESCE(r.fasilitas, '[]'::jsonb),
			COALESCE(r.gambar, '[]'::jsonb),
			r.is_active,
			r.created_at
		FROM ruangan r
		JOIN kategori k ON k.id = r.kategori_id
	`

	conditions := []string{"r.is_active = TRUE"}
	args := make([]any, 0)
	argPos := 1

	if filter.KategoriID != "" {
		conditions = append(conditions, fmt.Sprintf("r.kategori_id = $%d", argPos))
		args = append(args, filter.KategoriID)
		argPos++
	}
	if filter.Kategori != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(k.nama_kategori) = LOWER($%d)", argPos))
		args = append(args, strings.TrimSpace(filter.Kategori))
		argPos++
	}
	if filter.Kota != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(r.kota) = LOWER($%d)", argPos))
		args = append(args, strings.TrimSpace(filter.Kota))
		argPos++
	}
	if filter.MinHarga > 0 {
		conditions = append(conditions, fmt.Sprintf("r.harga_per_hari >= $%d", argPos))
		args = append(args, filter.MinHarga)
		argPos++
	}
	if filter.MaxHarga > 0 {
		conditions = append(conditions, fmt.Sprintf("r.harga_per_hari <= $%d", argPos))
		args = append(args, filter.MaxHarga)
		argPos++
	}
	if filter.Kapasitas > 0 {
		conditions = append(conditions, fmt.Sprintf("r.kapasitas >= $%d", argPos))
		args = append(args, filter.Kapasitas)
		argPos++
	}
	if filter.TanggalKetersediaan != nil {
		startOfDay := filter.TanggalKetersediaan.UTC().Truncate(24 * time.Hour)
		endOfDay := startOfDay.Add(24 * time.Hour)
		conditions = append(conditions, fmt.Sprintf(`
			NOT EXISTS (
				SELECT 1
				FROM bookings b
				WHERE b.ruangan_id = r.id
					AND b.status IN ('pending', 'confirmed')
					AND b.tanggal_mulai < $%d
					AND b.tanggal_selesai > $%d
			)
		`, argPos, argPos+1))
		args = append(args, endOfDay, startOfDay)
		argPos += 2
	}

	query := baseQuery + " WHERE " + strings.Join(conditions, " AND ") + " ORDER BY r.created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]models.RuanganResponse, 0)
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, *room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*models.RuanganResponse, error) {
	const query = `
		SELECT
			r.id::text,
			r.nama_ruangan,
			r.kategori_id::text,
			k.nama_kategori,
			COALESCE(r.deskripsi, ''),
			r.alamat,
			r.kota,
			r.kapasitas,
			r.harga_per_jam,
			r.harga_per_hari,
			r.stock_availability,
			COALESCE(r.fasilitas, '[]'::jsonb),
			COALESCE(r.gambar, '[]'::jsonb),
			r.is_active,
			r.created_at
		FROM ruangan r
		JOIN kategori k ON k.id = r.kategori_id
		WHERE r.id = $1 AND r.is_active = TRUE
	`

	row := r.db.QueryRowContext(ctx, query, id)
	room, err := scanRoom(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	return room, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRoom(s scanner) (*models.RuanganResponse, error) {
	var room models.RuanganResponse
	var fasilitasRaw []byte
	var gambarRaw []byte

	err := s.Scan(
		&room.ID,
		&room.NamaRuangan,
		&room.KategoriID,
		&room.KategoriNama,
		&room.Deskripsi,
		&room.Alamat,
		&room.Kota,
		&room.Kapasitas,
		&room.HargaPerJam,
		&room.HargaPerHari,
		&room.StockAvailability,
		&fasilitasRaw,
		&gambarRaw,
		&room.IsActive,
		&room.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(fasilitasRaw, &room.Fasilitas); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(gambarRaw, &room.Gambar); err != nil {
		return nil, err
	}

	return &room, nil
}
