package pdf

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestRenderer_RenderInvoice_producesValidPDF(t *testing.T) {
	r := NewRenderer()
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	periodEnd := now.AddDate(0, 1, 0)
	tenant := &domain.Tenant{
		ID:      uuid.New(),
		Name:    "Acme SaaS",
		Email:   "billing@acme.test",
		Website: "https://acme.test",
		Branding: map[string]any{
			"primary_color": "#134E4A",
		},
	}
	inv := &domain.Invoice{
		ID:          uuid.MustParse("e50530f0-2a12-44b8-b3c1-40c9e654bff3"),
		Status:      domain.InvoiceStatusPaid,
		AmountDue:   100000,
		AmountPaid:  100000,
		Currency:    "NGN",
		PeriodStart: now,
		PeriodEnd:   periodEnd,
		CreatedAt:   now,
		PaidAt:      &now,
	}
	items := []*domain.InvoiceLineItem{{
		Description: "Pro plan — monthly",
		Amount:      100000,
		Currency:    "NGN",
	}}
	customer := &domain.Customer{
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}

	out, err := r.RenderInvoice(tenant, inv, items, customer)
	require.NoError(t, err)
	require.True(t, len(out) > 800, "pdf should have substantial content, got %d bytes", len(out))
	require.Equal(t, "%PDF", string(out[:4]))
}

func TestParseHexColor(t *testing.T) {
	r, g, b, ok := parseHexColor("#134E4A")
	require.True(t, ok)
	require.Equal(t, 19, r)
	require.Equal(t, 78, g)
	require.Equal(t, 74, b)
}
