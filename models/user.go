package models

import (
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID          string    `json:"id" db:"id"`
	Email       string    `json:"email" db:"email" validate:"required,email"`
	NamaLengkap string    `json:"nama_lengkap" db:"nama_lengkap" validate:"required"`
	TTL         string    `json:"ttl" db:"ttl" validate:"required"`
	NoHP        string    `json:"no_hp" db:"no_hp" validate:"required"`
	Password    string    `json:"-" db:"password"`
	Role        UserRole  `json:"role" db:"role"`
	OTPCode     string    `json:"-" db:"otp_code"`
	OTPExpiry   time.Time `json:"-" db:"otp_expiry"`
	IsVerified  bool      `json:"is_verified" db:"is_verified"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	NamaLengkap string `json:"nama_lengkap" validate:"required"`
	TTL         string `json:"ttl" validate:"required"`
	NoHP        string `json:"no_hp" validate:"required"`
	Password    string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	User        UserResponse `json:"user"`
}

type UpdateUserRequest struct {
	Email       string   `json:"email"`
	NamaLengkap string   `json:"nama_lengkap"`
	TTL         string   `json:"ttl"`
	NoHP        string   `json:"no_hp"`
	Password    string   `json:"password"`
	Role        UserRole `json:"role"`
	IsVerified  *bool    `json:"is_verified"`
}

type OTPVerifyRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

type OTPSendRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTPCode     string `json:"otp_code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	NamaLengkap string    `json:"nama_lengkap"`
	NoHP        string    `json:"no_hp"`
	Role        UserRole  `json:"role"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
}
