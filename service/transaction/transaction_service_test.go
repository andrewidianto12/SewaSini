package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"sewasini/models"
)

func TestCreatePaymentSuccess(t *testing.T) {
	bookingRepo := &stubBookingRepo{
		booking: &models.Booking{
			ID:            "booking-1",
			UserID:        "user-1",
			TotalHarga:    250000,
			BookingCode:   "BOOK-001",
			Status:        models.BookingPending,
			PaymentStatus: models.PaymentUnpaid,
		},
	}
	userRepo := &stubUserRepo{
		user: &models.User{ID: "user-1", Email: "user@example.com"},
	}
	txRepo := &stubTransactionRepo{}
	xendit := &stubXenditClient{
		response: &CreateInvoiceResponse{
			ID:         "inv-123",
			PaymentURL: "https://pay.xendit.test/inv-123",
		},
	}
	emailer := &stubEmailer{}

	service := NewServiceWithDeps(txRepo, bookingRepo, userRepo, xendit, emailer)

	resp, err := service.CreatePayment(context.Background(), "user-1", models.CreateTransactionRequest{
		BookingID:     "booking-1",
		PaymentMethod: "BCA",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	if resp.Status != models.TransactionPending {
		t.Fatalf("expected pending status, got %s", resp.Status)
	}
	if resp.PaymentURL != "https://pay.xendit.test/inv-123" {
		t.Fatalf("unexpected payment url: %s", resp.PaymentURL)
	}
	if txRepo.created == nil {
		t.Fatal("expected transaction to be persisted")
	}
	if txRepo.created.XenditID != "inv-123" {
		t.Fatalf("expected xendit id to be stored, got %s", txRepo.created.XenditID)
	}
	if txRepo.created.ExternalID == "" {
		t.Fatal("expected external id to be generated")
	}
	if xendit.lastRequest == nil || xendit.lastRequest.ExternalID != txRepo.created.ExternalID {
		t.Fatal("expected xendit request to use persisted external id")
	}
}

func TestCreatePaymentRejectsWrongUser(t *testing.T) {
	service := NewServiceWithDeps(
		&stubTransactionRepo{},
		&stubBookingRepo{booking: &models.Booking{ID: "booking-1", UserID: "user-2", PaymentStatus: models.PaymentUnpaid}},
		&stubUserRepo{user: &models.User{ID: "user-1", Email: "user@example.com"}},
		&stubXenditClient{},
		&stubEmailer{},
	)

	_, err := service.CreatePayment(context.Background(), "user-1", models.CreateTransactionRequest{
		BookingID:     "booking-1",
		PaymentMethod: "BCA",
	})
	if !errors.Is(err, ErrBookingOwnership) {
		t.Fatalf("expected ErrBookingOwnership, got %v", err)
	}
}

func TestCreatePaymentRejectsNonPendingBooking(t *testing.T) {
	service := NewServiceWithDeps(
		&stubTransactionRepo{},
		&stubBookingRepo{booking: &models.Booking{ID: "booking-1", UserID: "user-1", Status: models.BookingCancelled, PaymentStatus: models.PaymentUnpaid}},
		&stubUserRepo{user: &models.User{ID: "user-1", Email: "user@example.com"}},
		&stubXenditClient{},
		&stubEmailer{},
	)

	_, err := service.CreatePayment(context.Background(), "user-1", models.CreateTransactionRequest{
		BookingID:     "booking-1",
		PaymentMethod: "BCA",
	})
	if !errors.Is(err, ErrBookingNotPayable) {
		t.Fatalf("expected ErrBookingNotPayable, got %v", err)
	}
}

func TestHandleCallbackMarksBookingPaidAndSendsEmail(t *testing.T) {
	txRepo := &stubTransactionRepo{
		byExternalID: &models.Transaction{
			ID:         "tx-1",
			BookingID:  "booking-1",
			UserID:     "user-1",
			ExternalID: "external-1",
			Status:     models.TransactionPending,
		},
	}
	bookingRepo := &stubBookingRepo{
		booking: &models.Booking{
			ID:            "booking-1",
			UserID:        "user-1",
			PaymentStatus: models.PaymentUnpaid,
		},
	}
	userRepo := &stubUserRepo{
		user: &models.User{ID: "user-1", Email: "user@example.com"},
	}
	emailer := &stubEmailer{}

	service := NewServiceWithDeps(txRepo, bookingRepo, userRepo, &stubXenditClient{}, emailer)

	err := service.HandleCallback(context.Background(), models.XenditCallbackRequest{
		ID:         "xnd-1",
		ExternalID: "external-1",
		Status:     "PAID",
	})
	if err != nil {
		t.Fatalf("HandleCallback returned error: %v", err)
	}

	if txRepo.updatedStatus != models.TransactionSuccess {
		t.Fatalf("expected success status, got %s", txRepo.updatedStatus)
	}
	if !txRepo.successMarked {
		t.Fatal("expected atomic success update to run")
	}
	if emailer.sentTo != "user@example.com" {
		t.Fatalf("expected email to be sent to user@example.com, got %s", emailer.sentTo)
	}
	if txRepo.emailMarked != "external-1" {
		t.Fatalf("expected email_sent_at to be marked for external-1, got %s", txRepo.emailMarked)
	}
}

func TestHandleCallbackMapsExpiredStatus(t *testing.T) {
	txRepo := &stubTransactionRepo{
		byExternalID: &models.Transaction{
			ID:         "tx-1",
			BookingID:  "booking-1",
			UserID:     "user-1",
			ExternalID: "external-1",
		},
	}

	service := NewServiceWithDeps(txRepo, &stubBookingRepo{}, &stubUserRepo{}, &stubXenditClient{}, &stubEmailer{})
	err := service.HandleCallback(context.Background(), models.XenditCallbackRequest{
		ID:         "xnd-1",
		ExternalID: "external-1",
		Status:     "EXPIRED",
	})
	if err != nil {
		t.Fatalf("HandleCallback returned error: %v", err)
	}
	if txRepo.updatedStatus != models.TransactionExpired {
		t.Fatalf("expected expired status, got %s", txRepo.updatedStatus)
	}
}

func TestHandleCallbackDuplicateWebhookIsIgnored(t *testing.T) {
	txRepo := &stubTransactionRepo{
		byExternalID: &models.Transaction{
			ID:            "tx-1",
			BookingID:     "booking-1",
			UserID:        "user-1",
			ExternalID:    "external-1",
			Status:        models.TransactionSuccess,
			LastWebhookID: "webhook-1",
			EmailSentAt:   time.Now(),
		},
	}

	service := NewServiceWithDeps(txRepo, &stubBookingRepo{}, &stubUserRepo{}, &stubXenditClient{}, &stubEmailer{})
	err := service.HandleCallback(context.Background(), models.XenditCallbackRequest{
		ID:         "xnd-1",
		ExternalID: "external-1",
		Status:     "PAID",
		WebhookID:  "webhook-1",
	})
	if err != nil {
		t.Fatalf("HandleCallback returned error: %v", err)
	}
	if txRepo.successMarked {
		t.Fatal("expected duplicate callback to avoid state update")
	}
}

func TestGetPaymentRejectsWrongUser(t *testing.T) {
	service := NewServiceWithDeps(
		&stubTransactionRepo{
			byID: &models.Transaction{ID: "tx-1", UserID: "user-2"},
		},
		&stubBookingRepo{},
		&stubUserRepo{},
		&stubXenditClient{},
		&stubEmailer{},
	)

	_, err := service.GetPayment(context.Background(), "user-1", "tx-1")
	if !errors.Is(err, ErrPaymentOwnership) {
		t.Fatalf("expected ErrPaymentOwnership, got %v", err)
	}
}

type stubTransactionRepo struct {
	created       *models.Transaction
	byID          *models.Transaction
	byExternalID  *models.Transaction
	updatedStatus models.TransactionStatus
	updateXendit  string
	updateWebhook string
	successMarked bool
	emailMarked   string
}

func (s *stubTransactionRepo) Create(_ context.Context, tx *models.Transaction) error {
	copied := *tx
	copied.ID = "tx-created"
	copied.CreatedAt = time.Now()
	s.created = &copied
	tx.ID = copied.ID
	tx.CreatedAt = copied.CreatedAt
	return nil
}

func (s *stubTransactionRepo) GetByID(_ context.Context, id string) (*models.Transaction, error) {
	if s.byID == nil || s.byID.ID != id {
		return nil, ErrTransactionNotFound
	}
	return s.byID, nil
}

func (s *stubTransactionRepo) GetByExternalID(_ context.Context, externalID string) (*models.Transaction, error) {
	if s.byExternalID == nil || s.byExternalID.ExternalID != externalID {
		return nil, ErrTransactionNotFound
	}
	return s.byExternalID, nil
}

func (s *stubTransactionRepo) UpdateStatusByExternalID(_ context.Context, _ string, status models.TransactionStatus, xenditID, webhookID string) error {
	s.updatedStatus = status
	s.updateXendit = xenditID
	s.updateWebhook = webhookID
	return nil
}

func (s *stubTransactionRepo) MarkSuccessAndConfirmBooking(_ context.Context, _, xenditID, webhookID string) error {
	s.successMarked = true
	s.updatedStatus = models.TransactionSuccess
	s.updateXendit = xenditID
	s.updateWebhook = webhookID
	return nil
}

func (s *stubTransactionRepo) MarkEmailSent(_ context.Context, externalID string) error {
	s.emailMarked = externalID
	return nil
}

type stubBookingRepo struct {
	booking             *models.Booking
	markedPaidBookingID string
}

func (s *stubBookingRepo) GetByID(_ context.Context, id string) (*models.Booking, error) {
	if s.booking == nil || s.booking.ID != id {
		return nil, errors.New("booking not found")
	}
	return s.booking, nil
}

func (s *stubBookingRepo) MarkPaidAndConfirmed(_ context.Context, bookingID string) error {
	s.markedPaidBookingID = bookingID
	return nil
}

type stubUserRepo struct {
	user *models.User
}

func (s *stubUserRepo) GetByID(_ context.Context, id string) (*models.User, error) {
	if s.user == nil || s.user.ID != id {
		return nil, errors.New("user not found")
	}
	return s.user, nil
}

type stubXenditClient struct {
	lastRequest *CreateInvoiceRequest
	response    *CreateInvoiceResponse
	err         error
}

func (s *stubXenditClient) CreateInvoice(_ context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	s.lastRequest = &req
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

type stubEmailer struct {
	sentTo string
}

func (s *stubEmailer) Send(_ context.Context, toEmail, _, _, _ string) error {
	s.sentTo = toEmail
	return nil
}
