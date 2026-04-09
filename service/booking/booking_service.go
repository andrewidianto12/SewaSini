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
var ErrBookingOwnership = errors.New("booking does not belong to the authenticated user")
var ErrBookingNotEditable = errors.New("booking cannot be updated in its current status")

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

	return toBookingResponse(booking, room), nil
}

func (s *BookingService) ListUserBookings(ctx context.Context, userID string) ([]models.BookingResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	bookings, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.BookingResponse, 0, len(bookings))
	for i := range bookings {
		room, err := s.roomRepo.GetByID(ctx, bookings[i].RuanganID)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *toBookingResponse(&bookings[i], room))
	}

	return responses, nil
}

func (s *BookingService) GetUserBookingByID(ctx context.Context, userID, bookingID string) (*models.BookingResponse, error) {
	booking, room, err := s.getOwnedBookingWithRoom(ctx, userID, bookingID)
	if err != nil {
		return nil, err
	}

	return toBookingResponse(booking, room), nil
}

func (s *BookingService) UpdateBooking(ctx context.Context, userID, bookingID string, req models.UpdateBookingRequest) (*models.BookingResponse, error) {
	booking, room, err := s.getOwnedBookingWithRoom(ctx, userID, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status != models.BookingPending {
		return nil, ErrBookingNotEditable
	}

	updated := *booking
	if !req.TanggalMulai.IsZero() {
		updated.TanggalMulai = req.TanggalMulai.UTC()
	}
	if !req.TanggalSelesai.IsZero() {
		updated.TanggalSelesai = req.TanggalSelesai.UTC()
	}
	if req.JumlahPeserta > 0 {
		updated.JumlahPeserta = req.JumlahPeserta
	}

	if !updated.TanggalSelesai.After(updated.TanggalMulai) {
		return nil, ErrInvalidBookingDate
	}
	if updated.JumlahPeserta > room.Kapasitas {
		return nil, ErrInvalidParticipantCount
	}

	hasOverlap, err := s.repo.HasActiveOverlapExcluding(
		ctx,
		updated.ID,
		updated.RuanganID,
		updated.TanggalMulai.Format(time.RFC3339),
		updated.TanggalSelesai.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	if hasOverlap {
		return nil, ErrRoomUnavailable
	}

	updated.TotalHarga = calculateTotalHarga(room.HargaPerJam, room.HargaPerHari, updated.TanggalMulai, updated.TanggalSelesai)
	if err := s.repo.Update(ctx, &updated); err != nil {
		return nil, err
	}

	return toBookingResponse(&updated, room), nil
}

func (s *BookingService) CancelBooking(ctx context.Context, userID, bookingID string) error {
	booking, _, err := s.getOwnedBookingWithRoom(ctx, userID, bookingID)
	if err != nil {
		return err
	}
	if booking.Status == models.BookingCompleted {
		return ErrBookingNotEditable
	}
	if booking.Status == models.BookingCancelled {
		return nil
	}

	return s.repo.Cancel(ctx, bookingID)
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

func (s *BookingService) getOwnedBookingWithRoom(ctx context.Context, userID, bookingID string) (*models.Booking, *models.RuanganResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil, ErrUserIDRequired
	}

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, nil, err
	}
	if booking.UserID != userID {
		return nil, nil, ErrBookingOwnership
	}

	room, err := s.roomRepo.GetByID(ctx, booking.RuanganID)
	if err != nil {
		return nil, nil, err
	}

	return booking, room, nil
}

func toBookingResponse(booking *models.Booking, room *models.RuanganResponse) *models.BookingResponse {
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
	}
}
