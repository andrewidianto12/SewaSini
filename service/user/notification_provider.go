package user

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"sewasini/util"
)

var ErrOTPNotFound = errors.New("otp not requested")

type EmailNotifier interface {
	Send(ctx context.Context, toEmail, subject, textBody, htmlBody string) error
}

type OTPProvider interface {
	SendOTP(ctx context.Context, phoneNumber string) (string, error)
	VerifyOTP(ctx context.Context, phoneNumber, code string) (bool, error)
}

type NoopEmailNotifier struct{}

func (n *NoopEmailNotifier) Send(_ context.Context, _ string, _ string, _ string, _ string) error {
	return nil
}

type ErrorEmailNotifier struct {
	err error
}

func (n *ErrorEmailNotifier) Send(_ context.Context, _ string, _ string, _ string, _ string) error {
	if n.err == nil {
		return errors.New("email notifier is not configured")
	}

	return n.err
}

type MailjetEmailNotifier struct {
	mailer *util.MailjetMailer
}

func NewMailjetEmailNotifierFromEnv() (*MailjetEmailNotifier, error) {
	mailer, err := util.NewMailjetMailerFromEnv()
	if err != nil {
		return nil, err
	}

	return &MailjetEmailNotifier{mailer: mailer}, nil
}

func (m *MailjetEmailNotifier) Send(ctx context.Context, toEmail, subject, textBody, htmlBody string) error {
	return m.mailer.Send(ctx, toEmail, subject, textBody, htmlBody)
}

type SendGridEmailNotifier struct {
	apiKey   string
	fromName string
	fromMail string
	http     *http.Client
}

func NewSendGridEmailNotifierFromEnv() (*SendGridEmailNotifier, error) {
	apiKey := strings.TrimSpace(os.Getenv("SENDGRID_API_KEY"))
	fromMail := strings.TrimSpace(os.Getenv("SENDGRID_FROM_EMAIL"))
	fromName := strings.TrimSpace(os.Getenv("SENDGRID_FROM_NAME"))
	if apiKey == "" || fromMail == "" {
		return nil, errors.New("missing sendgrid configuration")
	}

	return &SendGridEmailNotifier{
		apiKey:   apiKey,
		fromName: fromName,
		fromMail: fromMail,
		http:     &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (s *SendGridEmailNotifier) Send(ctx context.Context, toEmail, subject, textBody, htmlBody string) error {
	payload := map[string]any{
		"personalizations": []map[string]any{{
			"to": []map[string]string{{"email": toEmail}},
		}},
		"from": map[string]string{
			"email": s.fromMail,
			"name":  s.fromName,
		},
		"subject": subject,
		"content": []map[string]string{
			{"type": "text/plain", "value": textBody},
			{"type": "text/html", "value": htmlBody},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendgrid error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return nil
}

type MailgunEmailNotifier struct {
	apiKey   string
	domain   string
	fromName string
	fromMail string
	baseURL  string
	http     *http.Client
}

func NewMailgunEmailNotifierFromEnv() (*MailgunEmailNotifier, error) {
	apiKey := strings.TrimSpace(os.Getenv("MAILGUN_API_KEY"))
	domain := strings.TrimSpace(os.Getenv("MAILGUN_DOMAIN"))
	fromMail := strings.TrimSpace(os.Getenv("MAILGUN_FROM_EMAIL"))
	fromName := strings.TrimSpace(os.Getenv("MAILGUN_FROM_NAME"))
	baseURL := strings.TrimSpace(os.Getenv("MAILGUN_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://api.mailgun.net"
	}
	if apiKey == "" || domain == "" || fromMail == "" {
		return nil, errors.New("missing mailgun configuration")
	}

	return &MailgunEmailNotifier{
		apiKey:   apiKey,
		domain:   domain,
		fromName: fromName,
		fromMail: fromMail,
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		http:     &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (m *MailgunEmailNotifier) Send(ctx context.Context, toEmail, subject, textBody, htmlBody string) error {
	endpoint := fmt.Sprintf("%s/v3/%s/messages", m.baseURL, m.domain)
	values := url.Values{}
	from := m.fromMail
	if m.fromName != "" {
		from = fmt.Sprintf("%s <%s>", m.fromName, m.fromMail)
	}

	values.Set("from", from)
	values.Set("to", toEmail)
	values.Set("subject", subject)
	values.Set("text", textBody)
	values.Set("html", htmlBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("api", m.apiKey)

	resp, err := m.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailgun error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return nil
}

type LocalOTPProvider struct {
	mu    sync.RWMutex
	store map[string]localOTPValue
}

type localOTPValue struct {
	code      string
	expiresAt time.Time
}

func NewLocalOTPProvider() *LocalOTPProvider {
	return &LocalOTPProvider{store: make(map[string]localOTPValue)}
}

func (l *LocalOTPProvider) SendOTP(_ context.Context, phoneNumber string) (string, error) {
	code, err := generateNumericOTP(6)
	if err != nil {
		return "", err
	}

	l.mu.Lock()
	l.store[phoneNumber] = localOTPValue{code: code, expiresAt: time.Now().Add(5 * time.Minute)}
	l.mu.Unlock()

	return code, nil
}

func (l *LocalOTPProvider) VerifyOTP(_ context.Context, phoneNumber, code string) (bool, error) {
	l.mu.RLock()
	entry, ok := l.store[phoneNumber]
	l.mu.RUnlock()
	if !ok {
		return false, ErrOTPNotFound
	}

	if time.Now().After(entry.expiresAt) {
		l.mu.Lock()
		delete(l.store, phoneNumber)
		l.mu.Unlock()
		return false, nil
	}

	if entry.code != code {
		return false, nil
	}

	l.mu.Lock()
	delete(l.store, phoneNumber)
	l.mu.Unlock()

	return true, nil
}

type TwilioOTPProvider struct {
	accountSID string
	authToken  string
	serviceSID string
	http       *http.Client
}

func NewTwilioOTPProviderFromEnv() (*TwilioOTPProvider, error) {
	accountSID := strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID"))
	authToken := strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN"))
	serviceSID := strings.TrimSpace(os.Getenv("TWILIO_VERIFY_SERVICE_SID"))
	if accountSID == "" || authToken == "" || serviceSID == "" {
		return nil, errors.New("missing twilio configuration")
	}

	return &TwilioOTPProvider{
		accountSID: accountSID,
		authToken:  authToken,
		serviceSID: serviceSID,
		http:       &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (t *TwilioOTPProvider) SendOTP(ctx context.Context, phoneNumber string) (string, error) {
	endpoint := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/Verifications", t.serviceSID)
	form := url.Values{}
	form.Set("To", phoneNumber)
	form.Set("Channel", "sms")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.accountSID, t.authToken)

	resp, err := t.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("twilio send otp failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return "", nil
}

func (t *TwilioOTPProvider) VerifyOTP(ctx context.Context, phoneNumber, code string) (bool, error) {
	endpoint := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/VerificationCheck", t.serviceSID)
	form := url.Values{}
	form.Set("To", phoneNumber)
	form.Set("Code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.accountSID, t.authToken)

	resp, err := t.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("twilio verify otp failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return false, err
	}

	return strings.EqualFold(payload.Status, "approved"), nil
}

type FirebaseOTPProvider struct {
	apiKey         string
	recaptchaToken string
	http           *http.Client
	mu             sync.RWMutex
	sessionByPhone map[string]string
}

func NewFirebaseOTPProviderFromEnv() (*FirebaseOTPProvider, error) {
	apiKey := strings.TrimSpace(os.Getenv("FIREBASE_API_KEY"))
	recaptchaToken := strings.TrimSpace(os.Getenv("FIREBASE_RECAPTCHA_TOKEN"))
	if apiKey == "" || recaptchaToken == "" {
		return nil, errors.New("missing firebase configuration")
	}

	return &FirebaseOTPProvider{
		apiKey:         apiKey,
		recaptchaToken: recaptchaToken,
		http:           &http.Client{Timeout: 15 * time.Second},
		sessionByPhone: make(map[string]string),
	}, nil
}

func (f *FirebaseOTPProvider) SendOTP(ctx context.Context, phoneNumber string) (string, error) {
	endpoint := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:sendVerificationCode?key=%s", url.QueryEscape(f.apiKey))
	payload := map[string]string{
		"phoneNumber":    phoneNumber,
		"recaptchaToken": f.recaptchaToken,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := f.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("firebase send otp failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		SessionInfo string `json:"sessionInfo"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.SessionInfo == "" {
		return "", errors.New("firebase session info missing")
	}

	f.mu.Lock()
	f.sessionByPhone[phoneNumber] = result.SessionInfo
	f.mu.Unlock()

	return "", nil
}

func (f *FirebaseOTPProvider) VerifyOTP(ctx context.Context, phoneNumber, code string) (bool, error) {
	f.mu.RLock()
	sessionInfo, ok := f.sessionByPhone[phoneNumber]
	f.mu.RUnlock()
	if !ok {
		return false, ErrOTPNotFound
	}

	endpoint := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPhoneNumber?key=%s", url.QueryEscape(f.apiKey))
	payload := map[string]string{
		"sessionInfo": sessionInfo,
		"code":        code,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := f.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(respBody), "INVALID_CODE") {
			return false, nil
		}
		return false, fmt.Errorf("firebase verify otp failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	f.mu.Lock()
	delete(f.sessionByPhone, phoneNumber)
	f.mu.Unlock()

	return true, nil
}

func loadEmailNotifierFromEnv() EmailNotifier {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")))
	switch provider {
	case "", "mailjet":
		notifier, err := NewMailjetEmailNotifierFromEnv()
		if err == nil {
			return notifier
		}
		return &ErrorEmailNotifier{err: fmt.Errorf("mailjet configuration error: %w", err)}
	case "sendgrid":
		notifier, err := NewSendGridEmailNotifierFromEnv()
		if err == nil {
			return notifier
		}
	case "mailgun":
		notifier, err := NewMailgunEmailNotifierFromEnv()
		if err == nil {
			return notifier
		}
	}
	return &NoopEmailNotifier{}
}

func LoadEmailNotifierFromEnv() EmailNotifier {
	return loadEmailNotifierFromEnv()
}

func loadOTPProviderFromEnv() OTPProvider {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("OTP_PROVIDER")))
	switch provider {
	case "twilio":
		otp, err := NewTwilioOTPProviderFromEnv()
		if err == nil {
			return otp
		}
	case "firebase":
		otp, err := NewFirebaseOTPProviderFromEnv()
		if err == nil {
			return otp
		}
	case "mailjet":
		return NewLocalOTPProvider()
	}
	return NewLocalOTPProvider()
}

func generateNumericOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid otp length")
	}

	const digits = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[n.Int64()]
	}

	return string(result), nil
}
