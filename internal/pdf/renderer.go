package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

//go:embed templates/invoice.html
var invoiceHTMLTemplate string

type Renderer struct{}

func NewRenderer() *Renderer { return &Renderer{} }

func (r *Renderer) RenderInvoice(tenant *domain.Tenant, inv *domain.Invoice, items []*domain.InvoiceLineItem) ([]byte, error) {
	type lineView struct {
		Description string
		Amount      int64
	}
	lines := make([]lineView, len(items))
	for i, item := range items {
		lines[i] = lineView{Description: item.Description, Amount: item.Amount}
	}

	data := map[string]any{
		"TenantName":  tenant.Name,
		"InvoiceID":   inv.ID.String(),
		"Status":      inv.Status,
		"AmountDue":   inv.AmountDue,
		"Currency":    inv.Currency,
		"PeriodStart": inv.PeriodStart.Format("2006-01-02"),
		"PeriodEnd":   inv.PeriodEnd.Format("2006-01-02"),
		"Lines":       lines,
	}

	t, err := template.New("invoice").Parse(invoiceHTMLTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	content := stripHTML(buf.String())
	pdf := fmt.Sprintf("%%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Contents 4 0 R>>endobj\n4 0 obj<</Length %d>>stream\nBT /F1 12 Tf 50 750 Td (%s) Tj ET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000206 00000 n \ntrailer<</Size 5/Root 1 0 R>>\nstartxref\n300\n%%%%EOF",
		len(content)+50, escapePDF(content))
	return []byte(pdf), nil
}

func stripHTML(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	for strings.Contains(s, "<") {
		start := strings.Index(s, "<")
		end := strings.Index(s[start:], ">")
		if end < 0 {
			break
		}
		s = s[:start] + s[start+end+1:]
	}
	return strings.TrimSpace(s)
}

func escapePDF(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}
