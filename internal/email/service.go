package email

import (
	"context"

	"github.com/ireoluwacodes/subsync/internal/config"
)

type MailerService struct {
	strategy EmailStrategy
	from     string
}

func NewMailerService(cfg *config.Config) *MailerService {
	var strategy EmailStrategy = NoopStrategy{}
	if cfg != nil && cfg.ResendAPIKey != "" {
		strategy = NewResendStrategy(cfg.ResendAPIKey)
	}
	from := ""
	if cfg != nil {
		from = cfg.ResendFromEmail
	}
	return &MailerService{strategy: strategy, from: from}
}

func (m *MailerService) Enabled() bool {
	_, ok := m.strategy.(NoopStrategy)
	return !ok
}

func (m *MailerService) Send(ctx context.Context, to, subject, html string) error {
	if to == "" {
		return nil
	}
	return m.strategy.Send(ctx, SendRequest{
		From:    m.from,
		To:      to,
		Subject: subject,
		HTML:    html,
	})
}
