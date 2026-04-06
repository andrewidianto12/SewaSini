package user

import (
	"context"

	"sewasini/models"
)

type Service interface {
	RegisterUser(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, error)
	SendOTP(ctx context.Context, req models.OTPSendRequest) error
	VerifyOTP(ctx context.Context, req models.OTPVerifyRequest) (*models.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*models.UserResponse, error)
	ListUsers(ctx context.Context) ([]models.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
}
