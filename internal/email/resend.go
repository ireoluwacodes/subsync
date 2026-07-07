package email

import (
	"context"

	"github.com/ireoluwacodes/subsync/internal/observability"
	"github.com/resend/resend-go/v3"
)

type ResendStrategy struct {
	client *resend.Client
}

func NewResendStrategy(apiKey string) *ResendStrategy {
	return &ResendStrategy{client: resend.NewClient(apiKey)}
}

func (s *ResendStrategy) Send(ctx context.Context, req SendRequest) error {
	params := &resend.SendEmailRequest{
		From:    req.From,
		To:      []string{req.To},
		Subject: req.Subject,
		Html:    req.HTML,
	}
	if req.Text != "" {
		params.Text = req.Text
	}
	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		observability.CaptureExternalAPIError("resend", "send_email", err, map[string]any{
			"email.to":      req.To,
			"email.subject": req.Subject,
		})
	}
	return err
}
