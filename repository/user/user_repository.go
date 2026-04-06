package user

import (
	"context"
	"database/sql"
	"errors"

	"sewasini/models"
)

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
}

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, user *models.User) error {
	const query = `
		INSERT INTO users (
			email, nama_lengkap, ttl, no_hp, password, role, otp_code, otp_expiry, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.NamaLengkap,
		user.TTL,
		user.NoHP,
		user.Password,
		user.Role,
		user.OTPCode,
		user.OTPExpiry,
		user.IsVerified,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	const query = `
		SELECT id, email, nama_lengkap, ttl, no_hp, password, role, otp_code, otp_expiry, is_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.NamaLengkap,
		&user.TTL,
		&user.NoHP,
		&user.Password,
		&user.Role,
		&user.OTPCode,
		&user.OTPExpiry,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *SQLRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
		SELECT id, email, nama_lengkap, ttl, no_hp, password, role, otp_code, otp_expiry, is_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.NamaLengkap,
		&user.TTL,
		&user.NoHP,
		&user.Password,
		&user.Role,
		&user.OTPCode,
		&user.OTPExpiry,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *SQLRepository) List(ctx context.Context) ([]models.User, error) {
	const query = `
		SELECT id, email, nama_lengkap, ttl, no_hp, password, role, otp_code, otp_expiry, is_verified, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.NamaLengkap,
			&user.TTL,
			&user.NoHP,
			&user.Password,
			&user.Role,
			&user.OTPCode,
			&user.OTPExpiry,
			&user.IsVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *SQLRepository) Update(ctx context.Context, user *models.User) error {
	const query = `
		UPDATE users
		SET email = $2,
			nama_lengkap = $3,
			ttl = $4,
			no_hp = $5,
			password = $6,
			role = $7,
			otp_code = $8,
			otp_expiry = $9,
			is_verified = $10,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.NamaLengkap,
		user.TTL,
		user.NoHP,
		user.Password,
		user.Role,
		user.OTPCode,
		user.OTPExpiry,
		user.IsVerified,
	).Scan(&user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
