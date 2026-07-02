package service

import (
	"testing"

	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/stretchr/testify/require"
)

func TestNewBillingService(t *testing.T) {
	cfg := &config.Config{BillingMockResult: "success"}
	inv := NewInvoiceService(nil, cfg, nil, nil)
	svc := NewBillingService(cfg, nil, nil, inv, &SubscriptionService{}, email.NewMailerService(nil), nil)
	require.NotNil(t, svc)
}

func TestNewDunningService(t *testing.T) {
	svc := NewDunningService(nil, nil, nil, &SubscriptionService{}, nil, email.NewMailerService(nil), nil)
	require.NotNil(t, svc)
}
