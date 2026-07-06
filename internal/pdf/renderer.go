package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type Renderer struct{}

func NewRenderer() *Renderer { return &Renderer{} }

func (r *Renderer) RenderInvoice(
	tenant *domain.Tenant,
	inv *domain.Invoice,
	items []*domain.InvoiceLineItem,
	customer *domain.Customer,
) ([]byte, error) {
	if tenant == nil || inv == nil {
		return nil, fmt.Errorf("tenant and invoice are required")
	}

	theme := themeFromTenant(tenant)
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	r.drawHeader(pdf, theme, tenant, inv)
	r.drawParties(pdf, theme, tenant, customer, inv)
	r.drawLineItems(pdf, theme, inv, items)
	r.drawTotals(pdf, theme, inv)
	r.drawFooter(pdf, theme)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Renderer) drawHeader(pdf *fpdf.Fpdf, theme brandTheme, tenant *domain.Tenant, inv *domain.Invoice) {
	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetTextColor(theme.AccentR, theme.AccentG, theme.AccentB)
	pdf.Cell(0, 10, theme.CompanyName)
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
	if tenant.Email != "" {
		pdf.Cell(0, 4, tenant.Email)
		pdf.Ln(4)
	}
	if tenant.Website != "" {
		pdf.Cell(0, 4, tenant.Website)
		pdf.Ln(4)
	}

	pdf.Ln(4)
	pdf.SetDrawColor(theme.AccentR, theme.AccentG, theme.AccentB)
	pdf.SetLineWidth(0.6)
	pdf.Line(18, pdf.GetY(), 192, pdf.GetY())
	pdf.Ln(8)

	// Invoice title block (right-aligned meta)
	yStart := pdf.GetY()
	pdf.SetXY(120, yStart-22)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(24, 24, 27)
	pdf.CellFormat(72, 8, "INVOICE", "", 1, "R", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
	pdf.SetX(120)
	pdf.CellFormat(72, 5, "Invoice #"+shortID(inv.ID), "", 1, "R", false, 0, "")
	pdf.SetX(120)
	pdf.CellFormat(72, 5, "Issued "+inv.CreatedAt.UTC().Format("2 Jan 2006"), "", 1, "R", false, 0, "")
	if inv.DueDate != nil {
		pdf.SetX(120)
		pdf.CellFormat(72, 5, "Due "+inv.DueDate.UTC().Format("2 Jan 2006"), "", 1, "R", false, 0, "")
	}

	pdf.SetXY(18, yStart)
	pdf.SetFont("Helvetica", "B", 10)
	status := strings.ToUpper(string(inv.Status))
	cr, cg, cb := statusColor(inv.Status)
	pdf.SetTextColor(cr, cg, cb)
	pdf.Cell(0, 6, "Status: "+status)
	pdf.Ln(8)
}

func (r *Renderer) drawParties(pdf *fpdf.Fpdf, theme brandTheme, tenant *domain.Tenant, customer *domain.Customer, inv *domain.Invoice) {
	colW := 84.0
	y := pdf.GetY()

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
	pdf.SetXY(18, y)
	pdf.Cell(colW, 5, "BILL TO")
	pdf.SetX(18 + colW + 6)
	pdf.Cell(colW, 5, "BILLING PERIOD")

	pdf.Ln(6)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(24, 24, 27)

	name := "Customer"
	email := ""
	if customer != nil {
		if customer.Name != "" {
			name = customer.Name
		}
		email = customer.Email
	}

	pdf.SetX(18)
	pdf.Cell(colW, 5, name)
	pdf.SetX(18 + colW + 6)
	period := inv.PeriodStart.UTC().Format("2 Jan 2006") + " – " + inv.PeriodEnd.UTC().Format("2 Jan 2006")
	pdf.Cell(colW, 5, period)

	if email != "" {
		pdf.Ln(5)
		pdf.SetX(18)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
		pdf.Cell(colW, 5, email)
	}

	pdf.Ln(10)
}

func (r *Renderer) drawLineItems(pdf *fpdf.Fpdf, theme brandTheme, inv *domain.Invoice, items []*domain.InvoiceLineItem) {
	descW := 118.0
	amtW := 56.0

	pdf.SetFillColor(theme.AccentR, theme.AccentG, theme.AccentB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(descW, 8, "  Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(amtW, 8, "Amount  ", "1", 1, "R", true, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(24, 24, 27)

	rows := items
	if len(rows) == 0 {
		rows = []*domain.InvoiceLineItem{{
			Description: "Subscription",
			Amount:      inv.AmountDue,
			Currency:    inv.Currency,
		}}
	}

	fill := false
	for _, item := range rows {
		if fill {
			pdf.SetFillColor(247, 247, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		currency := item.Currency
		if currency == "" {
			currency = inv.Currency
		}
		desc := item.Description
		if desc == "" {
			desc = "Line item"
		}
		pdf.CellFormat(descW, 8, "  "+desc, "LR", 0, "L", true, 0, "")
		pdf.CellFormat(amtW, 8, utils.FormatMoneyDisplay(item.Amount, currency)+"  ", "LR", 1, "R", true, 0, "")
		fill = !fill
	}
	pdf.CellFormat(descW+amtW, 0, "", "T", 1, "", false, 0, "")
	pdf.Ln(4)
}

func (r *Renderer) drawTotals(pdf *fpdf.Fpdf, theme brandTheme, inv *domain.Invoice) {
	labelX := 120.0
	valW := 56.0

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
	pdf.SetX(labelX)
	pdf.CellFormat(30, 7, "Amount due", "", 0, "R", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(24, 24, 27)
	pdf.CellFormat(valW, 7, utils.FormatMoneyDisplay(inv.AmountDue, inv.Currency), "", 1, "R", false, 0, "")

	if inv.AmountPaid > 0 {
		pdf.SetX(labelX)
		pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
		pdf.CellFormat(30, 7, "Amount paid", "", 0, "R", false, 0, "")
		pdf.SetTextColor(22, 101, 52)
		pdf.CellFormat(valW, 7, utils.FormatMoneyDisplay(inv.AmountPaid, inv.Currency), "", 1, "R", false, 0, "")
	}

	if inv.PaidAt != nil {
		pdf.SetX(labelX)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
		pdf.CellFormat(30, 5, "Paid on", "", 0, "R", false, 0, "")
		pdf.CellFormat(valW, 5, inv.PaidAt.UTC().Format("2 Jan 2006, 15:04 UTC"), "", 1, "R", false, 0, "")
	}

	pdf.Ln(2)
	pdf.SetDrawColor(theme.AccentR, theme.AccentG, theme.AccentB)
	pdf.SetLineWidth(0.4)
	pdf.Line(labelX, pdf.GetY(), 192, pdf.GetY())
	pdf.Ln(4)

	pdf.SetX(labelX)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(theme.AccentR, theme.AccentG, theme.AccentB)
	balance := inv.AmountDue - inv.AmountPaid
	if balance < 0 {
		balance = 0
	}
	if inv.Status == domain.InvoiceStatusPaid {
		balance = 0
	}
	pdf.CellFormat(30, 8, "Balance", "", 0, "R", false, 0, "")
	pdf.CellFormat(valW, 8, utils.FormatMoneyDisplay(balance, inv.Currency), "", 1, "R", false, 0, "")
}

func (r *Renderer) drawFooter(pdf *fpdf.Fpdf, theme brandTheme) {
	pdf.SetY(-18)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(theme.MutedR, theme.MutedG, theme.MutedB)
	pdf.CellFormat(0, 5, "Thank you for your business.", "", 0, "C", false, 0, "")
	pdf.Ln(4)
	pdf.CellFormat(0, 4, "Invoice generated by SubSync", "", 0, "C", false, 0, "")
}

func shortID(id uuid.UUID) string {
	s := strings.ToUpper(id.String())
	if len(s) >= 8 {
		return s[:8]
	}
	return s
}

func statusColor(status domain.InvoiceStatus) (r, g, b int) {
	switch status {
	case domain.InvoiceStatusPaid:
		return 22, 101, 52
	case domain.InvoiceStatusOpen, domain.InvoiceStatusProcessing:
		return 180, 83, 9
	case domain.InvoiceStatusVoid, domain.InvoiceStatusUncollectible:
		return 113, 113, 122
	default:
		return 24, 24, 27
	}
}
