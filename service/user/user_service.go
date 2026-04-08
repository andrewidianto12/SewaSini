package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sewasini/models"
	repositoryuser "sewasini/repository/user"
	"sewasini/util"
)

var ErrEmailAlreadyUsed = errors.New("email already used")
var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrUserNotVerified = errors.New("user not verified")
var ErrInvalidOTP = errors.New("invalid otp")
var ErrOTPExpiredOrNotFound = errors.New("otp expired or not found")
var ErrPhoneNumberRequired = errors.New("phone number is required for otp")
var ErrOTPEmailSendFailed = errors.New("failed to send otp email")

type UserService struct {
	repo     Repository
	emailer  EmailNotifier
	otpAgent OTPProvider
}

func NewService(repo Repository) *UserService {
	return NewServiceWithProvider(repo, loadEmailNotifierFromEnv(), loadOTPProviderFromEnv())
}

func NewServiceWithProvider(repo Repository, emailer EmailNotifier, otpAgent OTPProvider) *UserService {
	if emailer == nil {
		emailer = &NoopEmailNotifier{}
	}
	if otpAgent == nil {
		otpAgent = NewLocalOTPProvider()
	}

	return &UserService{
		repo:     repo,
		emailer:  emailer,
		otpAgent: otpAgent,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(req.Email))
	if _, err := s.repo.GetByEmail(ctx, normalizedEmail); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repositoryuser.ErrUserNotFound) {
		return nil, err
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	otpCode, err := generateNumericOTP(6)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:       normalizedEmail,
		NamaLengkap: strings.TrimSpace(req.NamaLengkap),
		TTL:         strings.TrimSpace(req.TTL),
		NoHP:        strings.TrimSpace(req.NoHP),
		Password:    hashedPassword,
		Role:        models.RoleUser,
		OTPCode:     otpCode,
		OTPExpiry:   time.Now().Add(5 * time.Minute),
		IsVerified:  false,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.emailer.Send(
		ctx,
		user.Email,
		"Kode OTP Verifikasi SewaSini",
		fmt.Sprintf("Akun berhasil dibuat. Kode OTP verifikasi Anda: %s. Berlaku 5 menit.", otpCode),
		fmt.Sprintf("<p>Akun berhasil dibuat.</p><p>Kode OTP verifikasi Anda: <strong>%s</strong></p><p>Berlaku 5 menit.</p>", otpCode),
	); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOTPEmailSendFailed, err)
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(req.Email))
	user, err := s.repo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		return nil, err
	}

	if !user.IsVerified {
		return nil, ErrUserNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := util.GenerateToken(user.ID, user.NamaLengkap)
	if err != nil {
		return nil, err
	}

	response := models.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        toUserResponse(user),
	}

	return &response, nil
}

func (s *UserService) SendOTP(ctx context.Context, req models.OTPSendRequest) error {
	normalizedEmail := strings.TrimSpace(strings.ToLower(req.Email))
	user, err := s.repo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		return err
	}

	generatedCode, err := generateNumericOTP(6)
	if err != nil {
		return err
	}

	user.OTPCode = generatedCode
	user.OTPExpiry = time.Now().Add(5 * time.Minute)
	if err := s.repo.Update(ctx, user); err != nil {
		return err
	}

	if err := s.emailer.Send(
		ctx,
		user.Email,
		"Kode OTP SewaSini",
		fmt.Sprintf("Kode OTP Anda: %s. Berlaku 5 menit.", generatedCode),
		fmt.Sprintf("<p>Kode OTP Anda: <strong>%s</strong></p><p>Berlaku 5 menit.</p>", generatedCode),
	); err != nil {
		return fmt.Errorf("%w: %v", ErrOTPEmailSendFailed, err)
	}

	return nil
}

func (s *UserService) VerifyOTP(ctx context.Context, req models.OTPVerifyRequest) (*models.UserResponse, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(req.Email))
	user, err := s.repo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(user.OTPCode) == "" || user.OTPExpiry.IsZero() {
		return nil, ErrOTPExpiredOrNotFound
	}

	if time.Now().After(user.OTPExpiry) {
		return nil, ErrOTPExpiredOrNotFound
	}

	if strings.TrimSpace(req.OTPCode) != user.OTPCode {
		return nil, ErrInvalidOTP
	}

	user.IsVerified = true
	user.OTPCode = ""
	user.OTPExpiry = time.Time{}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	_ = s.emailer.Send(
		ctx,
		user.Email,
		"Verifikasi Berhasil",
		"Akun Anda telah berhasil diverifikasi.",
		"<p>Akun Anda telah berhasil diverifikasi.</p>",
	)

	response := toUserResponse(user)
	return &response, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]models.UserResponse, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]models.UserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, toUserResponse(&users[i]))
	}

	return responses, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req models.UpdateUserRequest) (*models.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != "" {
		normalizedEmail := strings.TrimSpace(strings.ToLower(req.Email))
		if normalizedEmail != user.Email {
			existing, err := s.repo.GetByEmail(ctx, normalizedEmail)
			if err == nil && existing.ID != user.ID {
				return nil, ErrEmailAlreadyUsed
			}
			if err != nil && !errors.Is(err, repositoryuser.ErrUserNotFound) {
				return nil, err
			}
		}
		user.Email = normalizedEmail
	}
	if req.NamaLengkap != "" {
		user.NamaLengkap = strings.TrimSpace(req.NamaLengkap)
	}
	if req.TTL != "" {
		user.TTL = strings.TrimSpace(req.TTL)
	}
	if req.NoHP != "" {
		user.NoHP = strings.TrimSpace(req.NoHP)
	}
	if req.Password != "" {
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashedPassword
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.IsVerified != nil {
		user.IsVerified = *req.IsVerified
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func toUserResponse(user *models.User) models.UserResponse {
	return models.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		NamaLengkap: user.NamaLengkap,
		NoHP:        user.NoHP,
		Role:        user.Role,
		IsVerified:  user.IsVerified,
		CreatedAt:   user.CreatedAt,
	}
}
