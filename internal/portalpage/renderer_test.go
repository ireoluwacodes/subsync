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
