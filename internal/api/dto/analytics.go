package dto

import (
	"time"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

type AnalyticsDateRangeParams struct {
	From     string `form:"from"`
	To       string `form:"to"`
	Currency string `form:"currency"`
}

func (p AnalyticsDateRangeParams) ParseRange(defaultFrom, defaultTo time.Time) (time.Time, time.Time, error) {
	from := defaultFrom
	to := defaultTo
	if p.From != "" {
		t, err := time.Parse("2006-01-02", p.From)
		if err != nil {
			return time.Time{}, time.Time{}, NewBindError("invalid from date; use YYYY-MM-DD")
		}
		from = t.UTC()
	}
	if p.To != "" {
		t, err := time.Parse("2006-01-02", p.To)
		if err != nil {
			return time.Time{}, time.Time{}, NewBindError("invalid to date; use YYYY-MM-DD")
		}
		to = t.UTC().Add(24 * time.Hour)
	}
	return from, to, nil
}

type AnalyticsMRRResponse struct {
	MRR      int64  `json:"mrr"`
	Currency string `json:"currency"`
	Active   int64  `json:"active_subscriptions"`
}

type AnalyticsChurnResponse struct {
	CanceledInPeriod int64   `json:"canceled_in_period"`
	ActiveCount      int64   `json:"active_count"`
	ChurnRate        float64 `json:"churn_rate"`
	From             string  `json:"from"`
	To               string  `json:"to"`
}

type AnalyticsDunningResponse struct {
	EnteredPastDue     int64   `json:"entered_past_due"`
	Recovered          int64   `json:"recovered"`
	RecoveryRate       float64 `json:"recovery_rate"`
	CurrentlyPastDue   int64   `json:"currently_past_due"`
	From               string  `json:"from"`
	To                 string  `json:"to"`
}

type RevenueDailyPointResponse struct {
	Date   string `json:"date"`
	Amount int64  `json:"amount"`
}

type AnalyticsRevenueResponse struct {
	Total    int64                       `json:"total"`
	Currency string                      `json:"currency"`
	From     string                      `json:"from"`
	To       string                      `json:"to"`
	Daily    []RevenueDailyPointResponse `json:"daily"`
}

func AnalyticsMRRToResponse(r *domain.AnalyticsMRRResult) AnalyticsMRRResponse {
	return AnalyticsMRRResponse{
		MRR:      r.MRR,
		Currency: r.Currency,
		Active:   r.Active,
	}
}

func AnalyticsChurnToResponse(r *domain.AnalyticsChurnResult) AnalyticsChurnResponse {
	return AnalyticsChurnResponse{
		CanceledInPeriod: r.CanceledInPeriod,
		ActiveCount:      r.ActiveCount,
		ChurnRate:        r.ChurnRate,
		From:             r.From.Format("2006-01-02"),
		To:               r.To.Format("2006-01-02"),
	}
}

func AnalyticsDunningToResponse(r *domain.AnalyticsDunningResult) AnalyticsDunningResponse {
	return AnalyticsDunningResponse{
		EnteredPastDue:   r.EnteredPastDue,
		Recovered:        r.Recovered,
		RecoveryRate:     r.RecoveryRate,
		CurrentlyPastDue: r.CurrentlyPastDue,
		From:             r.From.Format("2006-01-02"),
		To:               r.To.Format("2006-01-02"),
	}
}

func AnalyticsRevenueToResponse(r *domain.AnalyticsRevenueResult) AnalyticsRevenueResponse {
	daily := make([]RevenueDailyPointResponse, len(r.Daily))
	for i, d := range r.Daily {
		daily[i] = RevenueDailyPointResponse{Date: d.Date, Amount: d.Amount}
	}
	return AnalyticsRevenueResponse{
		Total:    r.Total,
		Currency: r.Currency,
		From:     r.From.Format("2006-01-02"),
		To:       r.To.Format("2006-01-02"),
		Daily:    daily,
	}
}
