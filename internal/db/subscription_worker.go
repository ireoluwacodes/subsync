package db

import (
	"context"
	"time"

	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

func (r *SubscriptionRepo) ListDueForBilling(ctx context.Context, before time.Time, limit int) ([]*domain.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Subscription
	err := r.db.WithContext(ctx).
		Where("state IN ?", []string{
			string(domain.SubscriptionStateActive),
			string(domain.SubscriptionStatePastDue),
		}).
		Where("next_billing_at IS NOT NULL AND next_billing_at <= ?", before).
		Order("next_billing_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Subscription, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionToDomain(&rows[i])
	}
	return out, nil
}

func (r *SubscriptionRepo) ListTrialsEnding(ctx context.Context, before time.Time, limit int) ([]*domain.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Subscription
	err := r.db.WithContext(ctx).
		Where("state = ?", domain.SubscriptionStateTrialing).
		Where("trial_ends_at IS NOT NULL AND trial_ends_at <= ?", before).
		Order("trial_ends_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Subscription, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionToDomain(&rows[i])
	}
	return out, nil
}

func (r *SubscriptionRepo) ListCancelAtPeriodEnd(ctx context.Context, before time.Time, limit int) ([]*domain.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Subscription
	err := r.db.WithContext(ctx).
		Where("cancel_at_period_end = ?", true).
		Where("current_period_end <= ?", before).
		Where("state IN ?", []string{
			string(domain.SubscriptionStateActive),
			string(domain.SubscriptionStateTrialing),
		}).
		Order("current_period_end ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Subscription, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionToDomain(&rows[i])
	}
	return out, nil
}

func (r *SubscriptionRepo) ListResumingFromPause(ctx context.Context, before time.Time, limit int) ([]*domain.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Subscription
	err := r.db.WithContext(ctx).
		Where("state = ?", domain.SubscriptionStatePaused).
		Where("pause_ends_at IS NOT NULL AND pause_ends_at <= ?", before).
		Order("pause_ends_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Subscription, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionToDomain(&rows[i])
	}
	return out, nil
}

func (r *SubscriptionRepo) ListAwaitingPaymentMethodBeforeBilling(ctx context.Context, now time.Time, limit int) ([]*domain.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Subscription
	err := r.db.WithContext(ctx).
		Where("state IN ?", []string{
			string(domain.SubscriptionStateActive),
			string(domain.SubscriptionStateTrialing),
		}).
		Where("payment_method_id IS NULL").
		Where("next_billing_at IS NOT NULL AND next_billing_at > ?", now).
		Order("next_billing_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Subscription, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionToDomain(&rows[i])
	}
	return out, nil
}

func (r *SubscriptionRepo) ListPastDueForDunning(ctx context.Context, limit int) ([]*domain.Subscription, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Subscription
	err := r.db.WithContext(ctx).
		Where("state = ?", domain.SubscriptionStatePastDue).
		Order("dunning_started_at ASC NULLS FIRST").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Subscription, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionToDomain(&rows[i])
	}
	return out, nil
}
