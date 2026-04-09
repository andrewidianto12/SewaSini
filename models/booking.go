package models

import (
	"time"
)

type BookingStatus string
type PaymentStatus string

const (
	BookingPending    BookingStatus = "pending"
	BookingConfirmed  BookingStatus = "confirmed"
	BookingCancelled  BookingStatus = "cancelled"
	BookingCompleted  BookingStatus = "completed"

	PaymentUnpaid   PaymentStatus = "unpaid"
	PaymentPaid     PaymentStatus = "paid"
	PaymentRefunded PaymentStatus = "refunded"
)

type Booking struct {
	ID             string        `json:"id" db:"id"`
	UserID         string        `json:"user_id" db:"user_id" validate:"required"`
	RuanganID      string        `json:"ruangan_id" db:"ruangan_id" validate:"required"`
	TanggalMulai   time.Time     `json:"tanggal_mulai" db:"tanggal_mulai" validate:"required"`
	TanggalSelesai time.Time     `json:"tanggal_selesai" db:"tanggal_selesai" validate:"required"`
	JumlahPeserta  int           `json:"jumlah_peserta" db:"jumlah_peserta" validate:"required,min=1"`
	TotalHarga     int64         `json:"total_harga" db:"total_harga"`
	Status         BookingStatus `json:"status" db:"status"`
	PaymentStatus  PaymentStatus `json:"payment_status" db:"payment_status"`
	BookingCode    string        `json:"booking_code" db:"booking_code"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

type CreateBookingRequest struct {
	RuanganID      string    `json:"ruangan_id" validate:"required"`
	TanggalMulai   time.Time `json:"tanggal_mulai" validate:"required"`
	TanggalSelesai time.Time `json:"tanggal_selesai" validate:"required"`
	JumlahPeserta  int       `json:"jumlah_peserta" validate:"required,min=1"`
}

type UpdateBookingRequest struct {
	TanggalMulai   time.Time `json:"tanggal_mulai"`
	TanggalSelesai time.Time `json:"tanggal_selesai"`
	JumlahPeserta  int       `json:"jumlah_peserta"`
}

type AdminUpdateBookingRequest struct {
	Status        *BookingStatus `json:"status"`
	PaymentStatus *PaymentStatus `json:"payment_status"`
}

type BookingResponse struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id"`
	RuanganID      string        `json:"ruangan_id"`
	RuanganNama    string        `json:"ruangan_nama"`
	TanggalMulai   time.Time     `json:"tanggal_mulai"`
	TanggalSelesai time.Time     `json:"tanggal_selesai"`
	JumlahPeserta  int           `json:"jumlah_peserta"`
	TotalHarga     int64         `json:"total_harga"`
	Status         BookingStatus `json:"status"`
	PaymentStatus  PaymentStatus `json:"payment_status"`
	BookingCode    string        `json:"booking_code"`
	CreatedAt      time.Time     `json:"created_at"`
}
