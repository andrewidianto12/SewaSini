package booking

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"sewasini/models"
)

var ErrUserIDRequired = errors.New("user id is required")
var ErrInvalidBookingDate = errors.New("tanggal_selesai must be later than tanggal_mulai")
var ErrInvalidParticipantCount = errors.New("jumlah_peserta exceeds room capacity")
var ErrRoomUnavailable = errors.New("room is not available for the selected date")

type BookingService struct {
	repo     Repository
	roomRepo RoomRepository
}

func NewService(repo Repository, roomRepo RoomRepository) *BookingService {
	return &BookingService{
		repo:     repo,
		roomRepo: roomRepo,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID string, req models.CreateBookingRequest) (*models.BookingResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	if !req.TanggalSelesai.After(req.TanggalMulai) {
		return nil, ErrInvalidBookingDate
	}

	room, err := s.roomRepo.GetByID(ctx, req.RuanganID)
	if err != nil {
		return nil, err
	}

	if req.JumlahPeserta > room.Kapasitas {
		return nil, ErrInvalidParticipantCount
	}

	hasOverlap, err := s.repo.HasActiveOverlap(
		ctx,
		req.RuanganID,
		req.TanggalMulai.UTC().Format(time.RFC3339),
		req.TanggalSelesai.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	if hasOverlap {
		return nil, ErrRoomUnavailable
	}

	booking := &models.Booking{
		UserID:         userID,
		RuanganID:      req.RuanganID,
		TanggalMulai:   req.TanggalMulai.UTC(),
		TanggalSelesai: req.TanggalSelesai.UTC(),
		JumlahPeserta:  req.JumlahPeserta,
		TotalHarga:     calculateTotalHarga(room.HargaPerJam, room.HargaPerHari, req.TanggalMulai, req.TanggalSelesai),
		Status:         models.BookingPending,
		PaymentStatus:  models.PaymentUnpaid,
		BookingCode:    generateBookingCode(),
	}

	if err := s.repo.Create(ctx, booking); err != nil {
		return nil, err
	}

	return &models.BookingResponse{
		ID:             booking.ID,
		UserID:         booking.UserID,
		RuanganID:      booking.RuanganID,
		RuanganNama:    room.NamaRuangan,
		TanggalMulai:   booking.TanggalMulai,
		TanggalSelesai: booking.TanggalSelesai,
		JumlahPeserta:  booking.JumlahPeserta,
		TotalHarga:     booking.TotalHarga,
		Status:         booking.Status,
		PaymentStatus:  booking.PaymentStatus,
		BookingCode:    booking.BookingCode,
		CreatedAt:      booking.CreatedAt,
	}, nil
}

func calculateTotalHarga(hargaPerJam, hargaPerHari int64, tanggalMulai, tanggalSelesai time.Time) int64 {
	durationHours := tanggalSelesai.Sub(tanggalMulai).Hours()
	roundedHours := int64(math.Ceil(durationHours))
	durationDays := int64(math.Ceil(durationHours / 24))

	switch {
	case roundedHours <= 24 && hargaPerJam > 0:
		if roundedHours < 1 {
			roundedHours = 1
		}
		return hargaPerJam * roundedHours
	case hargaPerHari > 0:
		if durationDays < 1 {
			durationDays = 1
		}
		return hargaPerHari * durationDays
	case hargaPerJam > 0:
		if roundedHours < 1 {
			roundedHours = 1
		}
		return hargaPerJam * roundedHours
	default:
		return 0
	}
}

func generateBookingCode() string {
	return fmt.Sprintf("BOOK-%s-%03d", time.Now().UTC().Format("20060102150405"), rand.Intn(1000))
}
