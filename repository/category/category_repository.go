package category

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"sewasini/models"
)

var ErrCategoryNotFound = errors.New("category not found")
var ErrCategoryAlreadyExists = errors.New("category already exists")

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, category *models.Kategori) error {
	const query = `
		INSERT INTO kategori (nama_kategori, deskripsi)
		VALUES ($1, $2)
		RETURNING id::text, created_at
	`

	if err := r.db.QueryRowContext(ctx, query, category.NamaKategori, category.Deskripsi).Scan(&category.ID, &category.CreatedAt); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrCategoryAlreadyExists
		}
		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*models.KategoriResponse, error) {
	const query = `
		SELECT id::text, nama_kategori, COALESCE(deskripsi, ''), created_at
		FROM kategori
		WHERE id::text = $1
	`

	category := &models.KategoriResponse{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.NamaKategori,
		&category.Deskripsi,
		&category.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	return category, nil
}

func (r *SQLRepository) List(ctx context.Context) ([]models.KategoriResponse, error) {
	const query = `
		SELECT id::text, nama_kategori, COALESCE(deskripsi, ''), created_at
		FROM kategori
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.KategoriResponse, 0)
	for rows.Next() {
		var category models.KategoriResponse
		if err := rows.Scan(
			&category.ID,
			&category.NamaKategori,
			&category.Deskripsi,
			&category.CreatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *SQLRepository) Update(ctx context.Context, id string, category *models.Kategori) error {
	const query = `
		UPDATE kategori
		SET nama_kategori = $2,
			deskripsi = $3
		WHERE id::text = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, category.NamaKategori, category.Deskripsi)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrCategoryAlreadyExists
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM kategori WHERE id::text = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
