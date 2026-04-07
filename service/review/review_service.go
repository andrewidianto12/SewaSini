package review

import (
	"context"
	"errors"
	"strings"

	"sewasini/models"
	repositoryreview "sewasini/repository/review"
)

type ReviewService struct {
	repo Repository
}

func NewService(repo Repository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) CreateReview(ctx context.Context, userID string, req models.CreateReviewRequest) (*models.ReviewResponse, error) {
	userID = normalizeText(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	booking, err := s.repo.GetBookingByID(ctx, normalizeText(req.BookingID))
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, ErrForbiddenReviewAccess
	}
	if booking.RuanganID != normalizeText(req.RuanganID) {
		return nil, ErrBookingMismatch
	}

	if existing, err := s.repo.GetByUserAndBooking(ctx, userID, req.BookingID); err == nil && existing != nil {
		return nil, repositoryreview.ErrReviewAlreadyExists
	} else if err != nil && !errors.Is(err, repositoryreview.ErrReviewNotFound) {
		return nil, err
	}

	review := &models.Review{
		UserID:    userID,
		RuanganID: normalizeText(req.RuanganID),
		BookingID: normalizeText(req.BookingID),
		Rating:    req.Rating,
		Komentar:  strings.TrimSpace(req.Komentar),
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, review.ID)
}

func (s *ReviewService) ListReviews(ctx context.Context, userID string) ([]models.ReviewResponse, error) {
	userID = normalizeText(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	return s.repo.ListByUser(ctx, userID)
}

func (s *ReviewService) GetReviewByID(ctx context.Context, userID, id string) (*models.ReviewResponse, error) {
	review, err := s.repo.GetByID(ctx, normalizeText(id))
	if err != nil {
		return nil, err
	}

	if review.UserID != normalizeText(userID) {
		return nil, ErrForbiddenReviewAccess
	}

	return review, nil
}

func (s *ReviewService) UpdateReview(ctx context.Context, userID, id string, req models.UpdateReviewRequest) (*models.ReviewResponse, error) {
	if req.Rating == nil && req.Komentar == nil {
		return nil, ErrReviewUpdateEmpty
	}

	review, err := s.repo.GetByID(ctx, normalizeText(id))
	if err != nil {
		return nil, err
	}

	if review.UserID != normalizeText(userID) {
		return nil, ErrForbiddenReviewAccess
	}

	rating := review.Rating
	if req.Rating != nil {
		rating = *req.Rating
	}

	komentar := review.Komentar
	if req.Komentar != nil {
		komentar = strings.TrimSpace(*req.Komentar)
	}

	if err := s.repo.Update(ctx, review.ID, rating, komentar); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, review.ID)
}

func (s *ReviewService) DeleteReview(ctx context.Context, userID, id string) error {
	review, err := s.repo.GetByID(ctx, normalizeText(id))
	if err != nil {
		return err
	}

	if review.UserID != normalizeText(userID) {
		return ErrForbiddenReviewAccess
	}

	return s.repo.Delete(ctx, review.ID)
}
