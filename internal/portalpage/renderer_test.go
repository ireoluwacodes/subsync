package portalpage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderer_RenderHome_AwaitingPaymentMethod(t *testing.T) {
	r, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = r.RenderHome(&buf, HomeData{
		Title:                   "Manage subscription",
		Token:                   "test-token",
		TenantName:              "Acme",
		PlanName:                "Pro",
		CustomerEmail:           "user@example.com",
		State:                   "active",
		AwaitingPaymentMethod:   true,
		CanManagePaymentMethods: true,
		CanSetupDirectDebit:     true,
	})
	require.NoError(t, err)
	html := buf.String()
	require.Contains(t, html, "Add or update card")
	require.Contains(t, html, "direct debit")
}

func TestRenderer_RenderHome_CanceledHidesPaymentMethods(t *testing.T) {
	r, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = r.RenderHome(&buf, HomeData{
		Title:                   "Manage subscription",
		Token:                   "test-token",
		TenantName:              "Acme",
		PlanName:                "Pro",
		CustomerEmail:           "user@example.com",
		State:                   "canceled",
		CurrentPeriodStart:      "6 Jul 2026",
		CurrentPeriodEnd:        "7 Jul 2026",
		CanManagePaymentMethods: false,
	})
	require.NoError(t, err)
	html := buf.String()
	require.Contains(t, html, "Subscription canceled")
	require.Contains(t, html, "6 Jul 2026")
	require.NotContains(t, html, "Add or update card")
}

func TestRenderer_RenderDirectDebitForm_BankSelect(t *testing.T) {
	r, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = r.RenderDirectDebitForm(&buf, DirectDebitFormData{
		Title:        "Set up direct debit",
		Token:        "test-token",
		TenantName:   "Acme",
		PlanName:     "Pro",
		CustomerName: "Jane Doe",
		Banks: []PortalBank{
			{Code: "058", Name: "Guaranty Trust Bank"},
			{Code: "011", Name: "First Bank of Nigeria"},
		},
	})
	require.NoError(t, err)
	html := buf.String()
	require.Contains(t, html, `<select id="bank_code"`)
	require.Contains(t, html, "Guaranty Trust Bank")
	require.NotContains(t, html, "Bank code")
}

func TestRenderer_RenderBillingSuccess(t *testing.T) {
	r, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = r.RenderBillingSuccess(&buf, BillingSuccessData{
		Title:          "Payment successful",
		TenantName:     "Acme",
		PlanName:       "Pro",
		StatusLabel:    "Payment successful",
		StatusMessage:  "Your payment was received.",
		Outcome:        "success",
		OutcomeBadge:   "ok",
		OrderReference: "e50530f0-2a12-44b8-b3c1-40c9e654bff3",
	})
	require.NoError(t, err)
	html := buf.String()
	require.Contains(t, html, "Payment successful")
	require.Contains(t, html, "e50530f0-2a12-44b8-b3c1-40c9e654bff3")
}
