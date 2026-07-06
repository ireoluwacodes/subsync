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
		Title:                 "Manage subscription",
		Token:                 "test-token",
		TenantName:            "Acme",
		PlanName:              "Pro",
		CustomerEmail:         "user@example.com",
		State:                 "active",
		AwaitingPaymentMethod: true,
		CanSetupDirectDebit:   true,
	})
	require.NoError(t, err)
	html := buf.String()
	require.Contains(t, html, "Add or update card")
	require.Contains(t, html, "direct debit")
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
