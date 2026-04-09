package room

import (
	"context"
	"errors"
	"strings"

	"sewasini/models"
)

type RoomService struct {
	repo Repository
}

var ErrRoomUpdateEmpty = errors.New("at least one field must be provided")

func NewService(repo Repository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) List(ctx context.Context, filter models.RuanganFilter) (*models.RuanganListResponse, error) {
	return s.repo.List(ctx, filter)
}

func (s *RoomService) GetByID(ctx context.Context, id string) (*models.RuanganResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RoomService) CreateRoom(ctx context.Context, req models.CreateRuanganRequest) (*models.RuanganResponse, error) {
	room := &models.Ruangan{
		NamaRuangan:       strings.TrimSpace(req.NamaRuangan),
		KategoriID:        strings.TrimSpace(req.KategoriID),
		Deskripsi:         strings.TrimSpace(req.Deskripsi),
		Alamat:            strings.TrimSpace(req.Alamat),
		Kota:              strings.TrimSpace(req.Kota),
		Kapasitas:         req.Kapasitas,
		HargaPerJam:       req.HargaPerJam,
		HargaPerHari:      req.HargaPerHari,
		StockAvailability: req.StockAvailability,
		Fasilitas:         req.Fasilitas,
		Gambar:            req.Gambar,
		IsActive:          true,
	}
	if room.StockAvailability == 0 {
		room.StockAvailability = 1
	}

	if err := s.repo.Create(ctx, room); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, room.ID)
}

func (s *RoomService) UpdateRoom(ctx context.Context, id string, req models.UpdateRuanganRequest) (*models.RuanganResponse, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.NamaRuangan == "" && req.KategoriID == "" && req.Deskripsi == "" && req.Alamat == "" && req.Kota == "" &&
		req.Kapasitas == 0 && req.HargaPerJam == 0 && req.HargaPerHari == 0 && req.StockAvailability == 0 &&
		len(req.Fasilitas) == 0 && len(req.Gambar) == 0 && req.IsActive == nil {
		return nil, ErrRoomUpdateEmpty
	}

	room := &models.Ruangan{
		NamaRuangan:       current.NamaRuangan,
		KategoriID:        current.KategoriID,
		Deskripsi:         current.Deskripsi,
		Alamat:            current.Alamat,
		Kota:              current.Kota,
		Kapasitas:         current.Kapasitas,
		HargaPerJam:       current.HargaPerJam,
		HargaPerHari:      current.HargaPerHari,
		StockAvailability: current.StockAvailability,
		Fasilitas:         current.Fasilitas,
		Gambar:            current.Gambar,
		IsActive:          current.IsActive,
	}

	if req.NamaRuangan != "" {
		room.NamaRuangan = strings.TrimSpace(req.NamaRuangan)
	}
	if req.KategoriID != "" {
		room.KategoriID = strings.TrimSpace(req.KategoriID)
	}
	if req.Deskripsi != "" {
		room.Deskripsi = strings.TrimSpace(req.Deskripsi)
	}
	if req.Alamat != "" {
		room.Alamat = strings.TrimSpace(req.Alamat)
	}
	if req.Kota != "" {
		room.Kota = strings.TrimSpace(req.Kota)
	}
	if req.Kapasitas > 0 {
		room.Kapasitas = req.Kapasitas
	}
	if req.HargaPerJam > 0 {
		room.HargaPerJam = req.HargaPerJam
	}
	if req.HargaPerHari > 0 {
		room.HargaPerHari = req.HargaPerHari
	}
	if req.StockAvailability > 0 {
		room.StockAvailability = req.StockAvailability
	}
	if len(req.Fasilitas) > 0 {
		room.Fasilitas = req.Fasilitas
	}
	if len(req.Gambar) > 0 {
		room.Gambar = req.Gambar
	}
	if req.IsActive != nil {
		room.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, id, room); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *RoomService) DeleteRoom(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
