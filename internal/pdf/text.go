package pdf

import (
	"fmt"
	"strings"
	"time"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// pdfText strips/replaces characters Helvetica cannot render in core PDF fonts.
func pdfText(s string) string {
	s = strings.NewReplacer(
		"\u2013", "-",
		"\u2014", "-",
		"\u2018", "'",
		"\u2019", "'",
		"\u201c", `"`,
		"\u201d", `"`,
	).Replace(s)

	var b strings.Builder
	for _, r := range s {
		if r < 128 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatMoneyPDF(minor int64, currency string) string {
	major := float64(minor) / 100
	cur := strings.ToUpper(strings.TrimSpace(currency))
	amount := formatMajorPDF(major)
	switch cur {
	case "NGN":
		return "NGN " + amount
	case "USD":
		return "USD " + amount
	case "GBP":
		return "GBP " + amount
	case "EUR":
		return "EUR " + amount
	case "":
		return amount
	default:
		return cur + " " + amount
	}
}

func formatMajorPDF(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}

func formatDatePDF(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2 Jan 2006")
}

func formatPeriodPDF(start, end time.Time) string {
	if start.IsZero() && end.IsZero() {
		return ""
	}
	if end.IsZero() {
		return formatDatePDF(start)
	}
	if start.IsZero() {
		return formatDatePDF(end)
	}
	return formatDatePDF(start) + " - " + formatDatePDF(end)
}

func invoiceIssuedAt(inv *domain.Invoice) time.Time {
	if inv == nil {
		return time.Time{}
	}
	if !inv.CreatedAt.IsZero() {
		return inv.CreatedAt.UTC()
	}
	if !inv.PeriodStart.IsZero() {
		return inv.PeriodStart.UTC()
	}
	if inv.PaidAt != nil && !inv.PaidAt.IsZero() {
		return inv.PaidAt.UTC()
	}
	return time.Time{}
}
