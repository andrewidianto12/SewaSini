package transaction

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"sewasini/models"
	repositorybooking "sewasini/repository/booking"
	repositorytransaction "sewasini/repository/transaction"
	repositoryuser "sewasini/repository/user"
	serviceuser "sewasini/service/user"
)

var ErrUserIDRequired = errors.New("user id is required")
var ErrPaymentMethodRequired = errors.New("payment method is required")
var ErrBookingAlreadyPaid = errors.New("booking is already paid")
var ErrBookingNotPayable = errors.New("booking is not payable")
var ErrBookingOwnership = errors.New("booking does not belong to the authenticated user")
var ErrTransactionNotFound = errors.New("transaction not found")
var ErrCallbackIgnored = errors.New("callback status ignored")

type EmailNotifier interface {
	Send(ctx context.Context, toEmail, subject, textBody, htmlBody string) error
}

type XenditClient interface {
	CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error)
}

type BookingPaymentService struct {
	repo         Repository
	bookingRepo  BookingRepository
	userRepo     UserRepository
	xenditClient XenditClient
	emailer      EmailNotifier
}

func NewService(repo Repository, bookingRepo BookingRepository, userRepo UserRepository) *BookingPaymentService {
	return NewServiceWithDeps(repo, bookingRepo, userRepo, NewXenditClientFromEnv(), serviceuser.LoadEmailNotifierFromEnv())
}

func NewServiceWithDeps(repo Repository, bookingRepo BookingRepository, userRepo UserRepository, xenditClient XenditClient, emailer EmailNotifier) *BookingPaymentService {
	if emailer == nil {
		emailer = &serviceuser.NoopEmailNotifier{}
	}

	return &BookingPaymentService{
		repo:         repo,
		bookingRepo:  bookingRepo,
		userRepo:     userRepo,
		xenditClient: xenditClient,
		emailer:      emailer,
	}
}

func (s *BookingPaymentService) CreatePayment(ctx context.Context, userID string, req models.CreateTransactionRequest) (*models.TransactionResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	if strings.TrimSpace(req.PaymentMethod) == "" {
		return nil, ErrPaymentMethodRequired
	}

	booking, err := s.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		return nil, err
	}
	if booking.UserID != userID {
		return nil, ErrBookingOwnership
	}
	if booking.PaymentStatus == models.PaymentPaid {
		return nil, ErrBookingAlreadyPaid
	}
	if booking.Status != models.BookingPending {
		return nil, ErrBookingNotPayable
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	externalID := generateExternalID(booking.ID)
	invoice, err := s.xenditClient.CreateInvoice(ctx, CreateInvoiceRequest{
		ExternalID:    externalID,
		Amount:        booking.TotalHarga,
		PayerEmail:    user.Email,
		Description:   fmt.Sprintf("Pembayaran booking %s", booking.BookingCode),
		PaymentMethod: req.PaymentMethod,
	})
	if err != nil {
		return nil, err
	}

	tx := &models.Transaction{
		BookingID:       booking.ID,
		UserID:          booking.UserID,
		Amount:          booking.TotalHarga,
		PaymentMethod:   strings.TrimSpace(req.PaymentMethod),
		TransactionDate: time.Now().UTC(),
		Status:          models.TransactionPending,
		ExternalID:      externalID,
		XenditID:        invoice.ID,
		PaymentURL:      invoice.PaymentURL,
	}
	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, err
	}

	return toTransactionResponse(tx), nil
}

func (s *BookingPaymentService) HandleCallback(ctx context.Context, req models.XenditCallbackRequest) error {
	if strings.TrimSpace(req.ExternalID) == "" {
		return ErrTransactionNotFound
	}

	tx, err := s.repo.GetByExternalID(ctx, req.ExternalID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.WebhookID) != "" && req.WebhookID == tx.LastWebhookID {
		return nil
	}

	status := mapCallbackStatus(req.Status)
	switch status {
	case models.TransactionSuccess:
		if tx.Status != models.TransactionSuccess {
			if err := s.repo.MarkSuccessAndConfirmBooking(ctx, req.ExternalID, req.ID, req.WebhookID); err != nil {
				return err
			}
			tx.Status = models.TransactionSuccess
		}
		if !tx.EmailSentAt.IsZero() {
			return nil
		}
		user, err := s.userRepo.GetByID(ctx, tx.UserID)
		if err == nil {
			if err := s.emailer.Send(
				ctx,
				user.Email,
				"Pembayaran Booking Berhasil",
				fmt.Sprintf("Pembayaran booking %s telah diterima. Status booking Anda sekarang confirmed.", tx.BookingID),
				fmt.Sprintf("<p>Pembayaran booking <strong>%s</strong> telah diterima.</p><p>Status booking Anda sekarang <strong>confirmed</strong>.</p>", tx.BookingID),
			); err == nil {
				return s.repo.MarkEmailSent(ctx, req.ExternalID)
			}
		}
		return nil
	case models.TransactionFailed, models.TransactionExpired:
		return s.repo.UpdateStatusByExternalID(ctx, req.ExternalID, status, req.ID, req.WebhookID)
	default:
		return ErrCallbackIgnored
	}
}

func toTransactionResponse(tx *models.Transaction) *models.TransactionResponse {
	return &models.TransactionResponse{
		ID:              tx.ID,
		BookingID:       tx.BookingID,
		Amount:          tx.Amount,
		PaymentMethod:   tx.PaymentMethod,
		TransactionDate: tx.TransactionDate,
		Status:          tx.Status,
		ExternalID:      tx.ExternalID,
		PaymentURL:      tx.PaymentURL,
		CreatedAt:       tx.CreatedAt,
	}
}

func mapCallbackStatus(status string) models.TransactionStatus {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PAID", "SETTLED", "SUCCEEDED":
		return models.TransactionSuccess
	case "EXPIRED":
		return models.TransactionExpired
	case "FAILED":
		return models.TransactionFailed
	default:
		return ""
	}
}

func generateExternalID(bookingID string) string {
	return fmt.Sprintf("booking-%s-%d-%03d", bookingID, time.Now().UTC().Unix(), rand.Intn(1000))
}

type CreateInvoiceRequest struct {
	ExternalID    string
	Amount        int64
	PayerEmail    string
	Description   string
	PaymentMethod string
}

type CreateInvoiceResponse struct {
	ID         string
	PaymentURL string
}

type xenditClient struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
}

func NewXenditClientFromEnv() XenditClient {
	secretKey := strings.TrimSpace(os.Getenv("XENDIT_SECRET_KEY"))
	baseURL := strings.TrimSpace(os.Getenv("XENDIT_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.xendit.co"
	}

	return &xenditClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (x *xenditClient) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	if strings.TrimSpace(x.secretKey) == "" {
		return nil, errors.New("xendit secret key is not configured")
	}

	payload := map[string]any{
		"external_id":      req.ExternalID,
		"amount":           req.Amount,
		"payer_email":      req.PayerEmail,
		"description":      req.Description,
		"invoice_duration": 86400,
	}
	if method := strings.TrimSpace(req.PaymentMethod); method != "" {
		payload["payment_methods"] = []string{strings.ToUpper(method)}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, x.baseURL+"/v2/invoices", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(x.secretKey+":")))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := x.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("xendit create invoice failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		ID         string `json:"id"`
		InvoiceURL string `json:"invoice_url"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	if parsed.ID == "" || parsed.InvoiceURL == "" {
		return nil, errors.New("xendit create invoice returned incomplete response")
	}

	return &CreateInvoiceResponse{ID: parsed.ID, PaymentURL: parsed.InvoiceURL}, nil
}

var (
	_ Service           = (*BookingPaymentService)(nil)
	_ Repository        = (*repositorytransaction.SQLRepository)(nil)
	_ BookingRepository = (*repositorybooking.SQLRepository)(nil)
	_ UserRepository    = (*repositoryuser.SQLRepository)(nil)
)
