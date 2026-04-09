package models

import (
	"time"
)

type Ruangan struct {
	ID                string    `json:"id" db:"id"`
	NamaRuangan       string    `json:"nama_ruangan" db:"nama_ruangan" validate:"required"`
	KategoriID        string    `json:"kategori_id" db:"kategori_id" validate:"required"`
	Deskripsi         string    `json:"deskripsi" db:"deskripsi"`
	Alamat            string    `json:"alamat" db:"alamat" validate:"required"`
	Kota              string    `json:"kota" db:"kota" validate:"required"`
	Kapasitas         int       `json:"kapasitas" db:"kapasitas" validate:"required,min=1"`
	HargaPerJam       int64     `json:"harga_per_jam" db:"harga_per_jam" validate:"required,min=0"`
	HargaPerHari      int64     `json:"harga_per_hari" db:"harga_per_hari" validate:"required,min=0"`
	StockAvailability int       `json:"stock_availability" db:"stock_availability"`
	Fasilitas         []string  `json:"fasilitas" db:"fasilitas"`
	Gambar            []string  `json:"gambar" db:"gambar"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type CreateRuanganRequest struct {
	NamaRuangan       string   `json:"nama_ruangan" validate:"required"`
	KategoriID        string   `json:"kategori_id" validate:"required"`
	Deskripsi         string   `json:"deskripsi"`
	Alamat            string   `json:"alamat" validate:"required"`
	Kota              string   `json:"kota" validate:"required"`
	Kapasitas         int      `json:"kapasitas" validate:"required,min=1"`
	HargaPerJam       int64    `json:"harga_per_jam" validate:"required,min=0"`
	HargaPerHari      int64    `json:"harga_per_hari" validate:"required,min=0"`
	StockAvailability int      `json:"stock_availability"`
	Fasilitas         []string `json:"fasilitas"`
	Gambar            []string `json:"gambar"`
}

type UpdateRuanganRequest struct {
	NamaRuangan       string   `json:"nama_ruangan"`
	KategoriID        string   `json:"kategori_id"`
	Deskripsi         string   `json:"deskripsi"`
	Alamat            string   `json:"alamat"`
	Kota              string   `json:"kota"`
	Kapasitas         int      `json:"kapasitas"`
	HargaPerJam       int64    `json:"harga_per_jam"`
	HargaPerHari      int64    `json:"harga_per_hari"`
	StockAvailability int      `json:"stock_availability"`
	Fasilitas         []string `json:"fasilitas"`
	Gambar            []string `json:"gambar"`
	IsActive          *bool    `json:"is_active"`
}

type RuanganResponse struct {
	ID                string    `json:"id"`
	NamaRuangan       string    `json:"nama_ruangan"`
	KategoriID        string    `json:"kategori_id"`
	KategoriNama      string    `json:"kategori_nama"`
	Deskripsi         string    `json:"deskripsi"`
	Alamat            string    `json:"alamat"`
	Kota              string    `json:"kota"`
	Kapasitas         int       `json:"kapasitas"`
	HargaPerJam       int64     `json:"harga_per_jam"`
	HargaPerHari      int64     `json:"harga_per_hari"`
	StockAvailability int       `json:"stock_availability"`
	Fasilitas         []string  `json:"fasilitas"`
	Gambar            []string  `json:"gambar"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

type RuanganFilter struct {
	Search              string     `query:"search"`
	Kategori            string     `query:"kategori"`
	KategoriID          string     `query:"kategori_id"`
	Kota                string     `query:"kota"`
	MinHarga            int64      `query:"min_harga"`
	MaxHarga            int64      `query:"max_harga"`
	Kapasitas           int        `query:"kapasitas"`
	TanggalKetersediaan *time.Time `query:"tanggal_ketersediaan"`
	Page                int        `query:"page"`
	Limit               int        `query:"limit"`
}

type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type RuanganListResponse struct {
	Data       []RuanganResponse   `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
