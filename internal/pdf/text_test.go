package pdf

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestPdfText_stripsUnicode(t *testing.T) {
	require.Equal(t, "5 Jul 2026 - 6 Jul 2026", pdfText("5 Jul 2026 – 6 Jul 2026"))
	require.Equal(t, "Pro plan - monthly", pdfText("Pro plan — monthly"))
}

func TestFormatMoneyPDF_usesASCII(t *testing.T) {
	require.Equal(t, "NGN 100", formatMoneyPDF(10000, "NGN"))
	require.Equal(t, "USD 1.50", formatMoneyPDF(150, "USD"))
}

func TestInvoiceIssuedAt_fallbacks(t *testing.T) {
	period := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	inv := &domain.Invoice{ID: uuid.New(), PeriodStart: period}
	require.Equal(t, period, invoiceIssuedAt(inv))

	created := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	inv.CreatedAt = created
	require.Equal(t, created, invoiceIssuedAt(inv))
}
