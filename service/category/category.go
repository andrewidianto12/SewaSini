package category

import (
	"context"
	"errors"

	"sewasini/models"
)

var ErrCategoryNameRequired = errors.New("nama_kategori is required")
var ErrCategoryUpdateEmpty = errors.New("at least one field must be provided")

type Repository interface {
	Create(ctx context.Context, category *models.Kategori) error
	GetByID(ctx context.Context, id string) (*models.KategoriResponse, error)
	List(ctx context.Context) ([]models.KategoriResponse, error)
	Update(ctx context.Context, id string, category *models.Kategori) error
	Delete(ctx context.Context, id string) error
}

type Service interface {
	CreateCategory(ctx context.Context, req models.CreateKategoriRequest) (*models.KategoriResponse, error)
	GetCategoryByID(ctx context.Context, id string) (*models.KategoriResponse, error)
	ListCategories(ctx context.Context) ([]models.KategoriResponse, error)
	UpdateCategory(ctx context.Context, id string, req models.UpdateKategoriRequest) (*models.KategoriResponse, error)
	DeleteCategory(ctx context.Context, id string) error
}
