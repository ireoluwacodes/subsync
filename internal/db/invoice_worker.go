package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

func (r *InvoiceRepo) FindForBillingPeriod(ctx context.Context, tenantID, subscriptionID uuid.UUID, periodStart, periodEnd time.Time) (*domain.Invoice, error) {
	var m models.Invoice
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND subscription_id = ?", tenantID, subscriptionID).
		Where("period_start = ? AND period_end = ?", periodStart, periodEnd).
		Where("status IN ?", []string{
			string(domain.InvoiceStatusOpen),
			string(domain.InvoiceStatusPaid),
		}).
		Order("created_at DESC").
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.InvoiceToDomain(&m), nil
}

func (r *InvoiceRepo) LatestOpenForSubscription(ctx context.Context, tenantID, subscriptionID uuid.UUID) (*domain.Invoice, error) {
	var m models.Invoice
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND subscription_id = ? AND status = ?", tenantID, subscriptionID, domain.InvoiceStatusOpen).
		Order("created_at DESC").
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.InvoiceToDomain(&m), nil
}
