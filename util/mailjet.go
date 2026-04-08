package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
)

type MailjetMailer struct {
	client    *mailjet.Client
	fromName  string
	fromEmail string
}

func getFirstEnv(keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}

	return ""
}

func NewMailjetMailerFromEnv() (*MailjetMailer, error) {
	apiKey := getFirstEnv("MAILJET_API_KEY", "MJ_APIKEY_PUBLIC")
	secretKey := getFirstEnv("MAILJET_SECRET_KEY", "MJ_APIKEY_PRIVATE")
	fromEmail := getFirstEnv("MAILJET_FROM_EMAIL", "SENDER_EMAIL")
	fromName := getFirstEnv("MAILJET_FROM_NAME", "SENDER_NAME")

	if apiKey == "" || secretKey == "" || fromEmail == "" {
		return nil, errors.New("missing mailjet configuration")
	}

	return &MailjetMailer{
		client:    mailjet.NewMailjetClient(apiKey, secretKey),
		fromName:  fromName,
		fromEmail: fromEmail,
	}, nil
}

func (m *MailjetMailer) Send(ctx context.Context, toEmail, subject, textBody, htmlBody string) error {
	if strings.TrimSpace(toEmail) == "" {
		return errors.New("destination email is required")
	}

	recipients := mailjet.RecipientsV31{{
		Email: toEmail,
	}}

	message := mailjet.InfoMessagesV31{
		From: &mailjet.RecipientV31{
			Email: m.fromEmail,
			Name:  m.fromName,
		},
		To:       &recipients,
		Subject:  subject,
		TextPart: textBody,
		HTMLPart: htmlBody,
	}

	payload := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{message},
	}

	if _, err := m.client.SendMailV31(&payload, mailjet.WithContext(ctx)); err != nil {
		return fmt.Errorf("mailjet send failed: %w", err)
	}

	return nil
}
