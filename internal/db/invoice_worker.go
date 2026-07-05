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
			string(domain.InvoiceStatusProcessing),
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

func (r *InvoiceRepo) GetByNombaOrderRef(ctx context.Context, tenantID uuid.UUID, orderRef string) (*domain.Invoice, error) {
	var m models.Invoice
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND nomba_order_ref = ?", tenantID, orderRef).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.InvoiceToDomain(&m), nil
}

func (r *InvoiceRepo) GetByNombaTransactionID(ctx context.Context, tenantID uuid.UUID, transactionID string) (*domain.Invoice, error) {
	var m models.Invoice
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND nomba_transaction_id = ?", tenantID, transactionID).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.InvoiceToDomain(&m), nil
}

func (r *InvoiceRepo) GetOpenBySubscription(ctx context.Context, tenantID, subscriptionID uuid.UUID) (*domain.Invoice, error) {
	var m models.Invoice
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND subscription_id = ? AND status IN ?", tenantID, subscriptionID, []string{
			string(domain.InvoiceStatusOpen),
			string(domain.InvoiceStatusProcessing),
		}).
		Order("created_at DESC").
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.InvoiceToDomain(&m), nil
}

func (r *InvoiceRepo) ListProcessingBefore(ctx context.Context, before time.Time, limit int) ([]*domain.Invoice, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Invoice
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.InvoiceStatusProcessing).
		Where("updated_at < ?", before).
		Order("updated_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Invoice, len(rows))
	for i := range rows {
		out[i] = models.InvoiceToDomain(&rows[i])
	}
	return out, nil
}
