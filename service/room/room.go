package room

import (
	"context"

	"sewasini/models"
)

type Repository interface {
	List(ctx context.Context, filter models.RuanganFilter) (*models.RuanganListResponse, error)
	GetByID(ctx context.Context, id string) (*models.RuanganResponse, error)
}

type Service interface {
	List(ctx context.Context, filter models.RuanganFilter) (*models.RuanganListResponse, error)
	GetByID(ctx context.Context, id string) (*models.RuanganResponse, error)
}
