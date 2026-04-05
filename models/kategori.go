package models

import (
	"time"
)

type Kategori struct {
	ID          string    `json:"id" db:"id"`
	NamaKategori string   `json:"nama_kategori" db:"nama_kategori" validate:"required"`
	Deskripsi   string    `json:"deskripsi" db:"deskripsi"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CreateKategoriRequest struct {
	NamaKategori string `json:"nama_kategori" validate:"required"`
	Deskripsi    string `json:"deskripsi"`
}

type UpdateKategoriRequest struct {
	NamaKategori string `json:"nama_kategori"`
	Deskripsi    string `json:"deskripsi"`
}

type KategoriResponse struct {
	ID           string    `json:"id"`
	NamaKategori string   `json:"nama_kategori"`
	Deskripsi    string    `json:"deskripsi"`
	CreatedAt    time.Time `json:"created_at"`
}
