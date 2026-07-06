package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type AnalyticsRepo struct {
	db *DB
}

func NewAnalyticsRepo(db *DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) MRR(ctx context.Context, tenantID uuid.UUID, currency string) (*domain.AnalyticsMRRResult, error) {
	type row struct {
		MRR      int64
		Currency string
		Active   int64
	}
	var rows []row
	q := r.db.WithContext(ctx).Raw(`
		SELECT
			p.currency,
			COALESCE(SUM(
				CASE
					WHEN p.interval = 'annual' THEN p.amount / 12
					WHEN p.interval = 'custom' AND COALESCE(p.interval_days, 0) > 0
						THEN (p.amount * 30) / p.interval_days
					ELSE p.amount
				END
			), 0)::bigint AS mrr,
			COUNT(s.id) AS active
		FROM subscriptions s
		JOIN plans p ON p.id = s.plan_id AND p.tenant_id = s.tenant_id
		WHERE s.tenant_id = ?
			AND s.state IN ('active', 'trialing', 'past_due')
		GROUP BY p.currency
	`, tenantID)
	if currency != "" {
		q = r.db.WithContext(ctx).Raw(`
			SELECT
				p.currency,
				COALESCE(SUM(
					CASE
						WHEN p.interval = 'annual' THEN p.amount / 12
						WHEN p.interval = 'custom' AND COALESCE(p.interval_days, 0) > 0
							THEN (p.amount * 30) / p.interval_days
						ELSE p.amount
					END
				), 0)::bigint AS mrr,
				COUNT(s.id) AS active
			FROM subscriptions s
			JOIN plans p ON p.id = s.plan_id AND p.tenant_id = s.tenant_id
			WHERE s.tenant_id = ?
				AND s.state IN ('active', 'trialing', 'past_due')
				AND p.currency = ?
			GROUP BY p.currency
		`, tenantID, currency)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		cur := currency
		if cur == "" {
			cur = "NGN"
		}
		return &domain.AnalyticsMRRResult{Currency: cur}, nil
	}
	return &domain.AnalyticsMRRResult{
		MRR:      rows[0].MRR,
		Currency: rows[0].Currency,
		Active:   rows[0].Active,
	}, nil
}

func (r *AnalyticsRepo) Churn(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (*domain.AnalyticsChurnResult, error) {
	var canceled int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT subscription_id)
		FROM subscription_transitions
		WHERE tenant_id = ? AND to_state = 'canceled'
			AND created_at >= ? AND created_at < ?
	`, tenantID, from, to).Scan(&canceled).Error; err != nil {
		return nil, err
	}

	var active int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM subscriptions
		WHERE tenant_id = ? AND state IN ('active', 'trialing', 'past_due')
	`, tenantID).Scan(&active).Error; err != nil {
		return nil, err
	}

	denominator := active + canceled
	var rate float64
	if denominator > 0 {
		rate = float64(canceled) / float64(denominator)
	}

	return &domain.AnalyticsChurnResult{
		CanceledInPeriod: canceled,
		ActiveCount:      active,
		ChurnRate:        rate,
		From:             from,
		To:               to,
	}, nil
}

func (r *AnalyticsRepo) Dunning(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (*domain.AnalyticsDunningResult, error) {
	var entered int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT subscription_id)
		FROM subscription_transitions
		WHERE tenant_id = ? AND to_state = 'past_due'
			AND created_at >= ? AND created_at < ?
	`, tenantID, from, to).Scan(&entered).Error; err != nil {
		return nil, err
	}

	var recovered int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT subscription_id)
		FROM subscription_transitions
		WHERE tenant_id = ? AND from_state = 'past_due' AND to_state = 'active'
			AND reason = 'payment_succeeded'
			AND created_at >= ? AND created_at < ?
	`, tenantID, from, to).Scan(&recovered).Error; err != nil {
		return nil, err
	}

	var currentPastDue int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM subscriptions
		WHERE tenant_id = ? AND state = 'past_due'
	`, tenantID).Scan(&currentPastDue).Error; err != nil {
		return nil, err
	}

	var rate float64
	if entered > 0 {
		rate = float64(recovered) / float64(entered)
	}

	return &domain.AnalyticsDunningResult{
		EnteredPastDue:     entered,
		Recovered:          recovered,
		RecoveryRate:       rate,
		CurrentlyPastDue:   currentPastDue,
		From:               from,
		To:                 to,
	}, nil
}

func (r *AnalyticsRepo) Revenue(ctx context.Context, tenantID uuid.UUID, from, to time.Time, currency string) (*domain.AnalyticsRevenueResult, error) {
	type totalRow struct {
		Total    int64
		Currency string
	}
	var totals []totalRow
	q := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount_paid), 0)::bigint AS total, currency
		FROM invoices
		WHERE tenant_id = ? AND status = 'paid'
			AND paid_at >= ? AND paid_at < ?
		GROUP BY currency
	`, tenantID, from, to)
	if currency != "" {
		q = r.db.WithContext(ctx).Raw(`
			SELECT COALESCE(SUM(amount_paid), 0)::bigint AS total, currency
			FROM invoices
			WHERE tenant_id = ? AND status = 'paid'
				AND paid_at >= ? AND paid_at < ?
				AND currency = ?
			GROUP BY currency
		`, tenantID, from, to, currency)
	}
	if err := q.Scan(&totals).Error; err != nil {
		return nil, err
	}

	result := &domain.AnalyticsRevenueResult{
		From: from,
		To:   to,
	}
	if len(totals) > 0 {
		result.Total = totals[0].Total
		result.Currency = totals[0].Currency
	} else if currency != "" {
		result.Currency = currency
	} else {
		result.Currency = "NGN"
	}

	type dailyRow struct {
		Date   string
		Amount int64
	}
	var daily []dailyRow
	var dailyErr error
	if currency != "" {
		dailyErr = r.db.WithContext(ctx).Raw(`
			SELECT TO_CHAR(paid_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date,
				COALESCE(SUM(amount_paid), 0)::bigint AS amount
			FROM invoices
			WHERE tenant_id = ? AND status = 'paid'
				AND paid_at >= ? AND paid_at < ?
				AND currency = ?
			GROUP BY 1
			ORDER BY 1
		`, tenantID, from, to, currency).Scan(&daily).Error
	} else {
		dailyErr = r.db.WithContext(ctx).Raw(`
			SELECT TO_CHAR(paid_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date,
				COALESCE(SUM(amount_paid), 0)::bigint AS amount
			FROM invoices
			WHERE tenant_id = ? AND status = 'paid'
				AND paid_at >= ? AND paid_at < ?
			GROUP BY 1
			ORDER BY 1
		`, tenantID, from, to).Scan(&daily).Error
	}
	if dailyErr != nil {
		return nil, dailyErr
	}
	for _, d := range daily {
		result.Daily = append(result.Daily, domain.RevenueDailyPoint{
			Date:   d.Date,
			Amount: d.Amount,
		})
	}
	return result, nil
}
