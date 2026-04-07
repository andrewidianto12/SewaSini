package room

import (
	"context"

	"sewasini/models"
)

type RoomService struct {
	repo Repository
}

func NewService(repo Repository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) List(ctx context.Context, filter models.RuanganFilter) ([]models.RuanganResponse, error) {
	return s.repo.List(ctx, filter)
}

func (s *RoomService) GetByID(ctx context.Context, id string) (*models.RuanganResponse, error) {
	return s.repo.GetByID(ctx, id)
}
