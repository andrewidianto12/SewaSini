package room

import (
	"context"

	"sewasini/models"
)

type Repository interface {
	List(ctx context.Context, filter models.RuanganFilter) (*models.RuanganListResponse, error)
	GetByID(ctx context.Context, id string) (*models.RuanganResponse, error)
	Create(ctx context.Context, room *models.Ruangan) error
	Update(ctx context.Context, id string, room *models.Ruangan) error
	Delete(ctx context.Context, id string) error
}

type Service interface {
	List(ctx context.Context, filter models.RuanganFilter) (*models.RuanganListResponse, error)
	GetByID(ctx context.Context, id string) (*models.RuanganResponse, error)
	CreateRoom(ctx context.Context, req models.CreateRuanganRequest) (*models.RuanganResponse, error)
	UpdateRoom(ctx context.Context, id string, req models.UpdateRuanganRequest) (*models.RuanganResponse, error)
	DeleteRoom(ctx context.Context, id string) error
}
