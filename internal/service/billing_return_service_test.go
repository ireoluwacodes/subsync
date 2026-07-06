package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestBillingReturnPresentation_paidInvoice(t *testing.T) {
	inv := &domain.Invoice{Status: domain.InvoiceStatusPaid}
	outcome, label, _ := billingReturnPresentation(inv, nil, false)
	require.Equal(t, BillingReturnSuccess, outcome)
	require.Equal(t, "Payment successful", label)
}

func TestBillingReturnPresentation_activeSubscription(t *testing.T) {
	sub := &domain.Subscription{State: domain.SubscriptionStateActive}
	outcome, label, _ := billingReturnPresentation(nil, sub, false)
	require.Equal(t, BillingReturnSuccess, outcome)
	require.Equal(t, "Subscription active", label)
}

func TestBillingReturnPresentation_nombaVerify(t *testing.T) {
	inv := &domain.Invoice{Status: domain.InvoiceStatusOpen}
	outcome, _, _ := billingReturnPresentation(inv, nil, true)
	require.Equal(t, BillingReturnSuccess, outcome)
}

func TestResolveCheckoutSuccessURL_defaultsToHostedPage(t *testing.T) {
	cfg := &config.Config{PublicBaseURL: "http://localhost:8080", AppEnv: "development"}
	url, err := resolveCheckoutSuccessURL("", cfg)
	require.NoError(t, err)
	require.Equal(t, "http://localhost:8080/billing/success", url)
}

func TestResolveCheckoutSuccessURL_customURL(t *testing.T) {
	cfg := &config.Config{PublicBaseURL: "http://localhost:8080", AppEnv: "development"}
	url, err := resolveCheckoutSuccessURL("http://localhost:3000/billing/success", cfg)
	require.NoError(t, err)
	require.Equal(t, "http://localhost:3000/billing/success", url)
}

func TestSubscriptionFromOrderRefPrefixes(t *testing.T) {
	subID := uuid.New()
	checkoutRef := domain.CheckoutOrderRefPrefix + subID.String() + "-" + uuid.New().String()
	parsed, ok := ParseCheckoutSubscriptionID(checkoutRef)
	require.True(t, ok)
	require.Equal(t, subID, parsed)

	captureRef := domain.CardCaptureOrderRefPrefix + subID.String()
	parsed, ok = ParseCardCaptureSubscriptionID(captureRef)
	require.True(t, ok)
	require.Equal(t, subID, parsed)
}
