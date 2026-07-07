package dto

import (
	"time"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type InvoiceResponse struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenant_id"`
	SubscriptionID   string `json:"subscription_id"`
	CustomerID       string `json:"customer_id"`
	Status           string `json:"status"`
	AmountDue        int64  `json:"amount_due"`
	AmountDueDisplay string `json:"amount_due_display"`
	AmountPaid       int64  `json:"amount_paid"`
	AmountPaidDisplay string `json:"amount_paid_display"`
	Currency         string `json:"currency"`
	CreatedAt        string `json:"created_at"`

	Subscription *SubscriptionResponse `json:"subscription,omitempty"`
	Customer     *CustomerResponse     `json:"customer,omitempty"`
}

func InvoiceToResponse(inv *domain.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:                inv.ID.String(),
		TenantID:          inv.TenantID.String(),
		SubscriptionID:    inv.SubscriptionID.String(),
		CustomerID:        inv.CustomerID.String(),
		Status:            string(inv.Status),
		AmountDue:         inv.AmountDue,
		AmountDueDisplay:  utils.FormatMoneyDisplay(inv.AmountDue, inv.Currency),
		AmountPaid:        inv.AmountPaid,
		AmountPaidDisplay: utils.FormatMoneyDisplay(inv.AmountPaid, inv.Currency),
		Currency:          inv.Currency,
		CreatedAt:         inv.CreatedAt.Format(time.RFC3339),
	}
}

// InvoiceToResponseWithRelations builds an invoice response with nested
// subscription and customer objects when they are available.
func InvoiceToResponseWithRelations(inv *domain.Invoice, sub *domain.Subscription, customer *domain.Customer) InvoiceResponse {
	resp := InvoiceToResponse(inv)
	if sub != nil {
		s := SubscriptionToResponse(sub)
		resp.Subscription = &s
	}
	if customer != nil {
		c := CustomerToResponse(customer)
		resp.Customer = &c
	}
	return resp
}

type PlanResponse struct {
	ID            string              `json:"id"`
	TenantID      string              `json:"tenant_id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Amount        int64               `json:"amount"`
	AmountDisplay string              `json:"amount_display"`
	Currency      string              `json:"currency"`
	Interval      domain.PlanInterval `json:"interval"`
	IntervalDays  *int                `json:"interval_days,omitempty"`
	TrialDays     int                 `json:"trial_days"`
	Features      []string            `json:"features"`
	IsActive      bool                `json:"is_active"`
	IsArchived    bool                `json:"is_archived"`
	CreatedAt     string              `json:"created_at"`
	UpdatedAt     string              `json:"updated_at"`
}

func PlanToResponse(p *domain.Plan) PlanResponse {
	return PlanResponse{
		ID:            p.ID.String(),
		TenantID:      p.TenantID.String(),
		Name:          p.Name,
		Description:   p.Description,
		Amount:        p.Amount,
		AmountDisplay: utils.FormatMoneyDisplay(p.Amount, p.Currency),
		Currency:      p.Currency,
		Interval:      p.Interval,
		IntervalDays:  p.IntervalDays,
		TrialDays:     p.TrialDays,
		Features:      p.Features,
		IsActive:      p.IsActive,
		IsArchived:    p.IsArchived,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type PlanStatsResponse struct {
	ActiveSubscribers  int64  `json:"active_subscribers"`
	MRREstimate        int64  `json:"mrr_estimate"`
	MRREstimateDisplay string `json:"mrr_estimate_display"`
	Currency           string `json:"currency"`
}

func PlanStatsToResponse(activeSubscribers, mrrEstimate int64, currency string) PlanStatsResponse {
	return PlanStatsResponse{
		ActiveSubscribers:  activeSubscribers,
		MRREstimate:        mrrEstimate,
		MRREstimateDisplay: utils.FormatMoneyDisplay(mrrEstimate, currency),
		Currency:           currency,
	}
}

type ProrationPreviewResponse struct {
	CreditAmount         int64  `json:"credit_amount"`
	CreditAmountDisplay  string `json:"credit_amount_display"`
	DebitAmount          int64  `json:"debit_amount"`
	DebitAmountDisplay   string `json:"debit_amount_display"`
	NetAmount            int64  `json:"net_amount"`
	NetAmountDisplay     string `json:"net_amount_display"`
	Currency             string `json:"currency"`
	DaysRemaining        int    `json:"days_remaining"`
	TotalDays            int    `json:"total_days"`
}

func ProrationToResponse(p *domain.ProrationResult, currency string) ProrationPreviewResponse {
	return ProrationPreviewResponse{
		CreditAmount:        p.CreditAmount,
		CreditAmountDisplay: utils.FormatMoneyDisplay(p.CreditAmount, currency),
		DebitAmount:         p.DebitAmount,
		DebitAmountDisplay:  utils.FormatMoneyDisplay(p.DebitAmount, currency),
		NetAmount:           p.NetAmount,
		NetAmountDisplay:    utils.FormatMoneyDisplay(p.NetAmount, currency),
		Currency:            currency,
		DaysRemaining:       p.DaysRemaining,
		TotalDays:           p.TotalDays,
	}
}
