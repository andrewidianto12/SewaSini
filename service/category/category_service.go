package category

import (
	"context"
	"strings"

	"sewasini/models"
)

type CategoryService struct {
	repo Repository
}

func NewService(repo Repository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, req models.CreateKategoriRequest) (*models.KategoriResponse, error) {
	nama := strings.TrimSpace(req.NamaKategori)
	if nama == "" {
		return nil, ErrCategoryNameRequired
	}

	category := &models.Kategori{
		NamaKategori: nama,
		Deskripsi:    strings.TrimSpace(req.Deskripsi),
	}
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}

	return &models.KategoriResponse{
		ID:           category.ID,
		NamaKategori: category.NamaKategori,
		Deskripsi:    category.Deskripsi,
		CreatedAt:    category.CreatedAt,
	}, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id string) (*models.KategoriResponse, error) {
	return s.repo.GetByID(ctx, strings.TrimSpace(id))
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]models.KategoriResponse, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id string, req models.UpdateKategoriRequest) (*models.KategoriResponse, error) {
	id = strings.TrimSpace(id)

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.NamaKategori == "" && req.Deskripsi == "" {
		return nil, ErrCategoryUpdateEmpty
	}

	nama := current.NamaKategori
	if req.NamaKategori != "" {
		nama = strings.TrimSpace(req.NamaKategori)
		if nama == "" {
			return nil, ErrCategoryNameRequired
		}
	}

	deskripsi := current.Deskripsi
	if req.Deskripsi != "" {
		deskripsi = strings.TrimSpace(req.Deskripsi)
	}

	payload := &models.Kategori{NamaKategori: nama, Deskripsi: deskripsi}
	if err := s.repo.Update(ctx, id, payload); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(id))
}
