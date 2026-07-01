package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type InvoiceRepo struct {
	db *DB
}

func NewInvoiceRepo(db *DB) *InvoiceRepo {
	return &InvoiceRepo{db: db}
}

func (r *InvoiceRepo) Create(ctx context.Context, invoice *domain.Invoice) error {
	m, err := models.InvoiceFromDomain(invoice)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	*invoice = *models.InvoiceToDomain(m)
	return nil
}

func (r *InvoiceRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Invoice, error) {
	var m models.Invoice
	if err := r.db.WithContext(ctx).First(&m, "id = ? AND tenant_id = ?", id, tenantID).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.InvoiceToDomain(&m), nil
}

func (r *InvoiceRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.InvoiceListFilter) ([]*domain.Invoice, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)
	if filter.CustomerID != nil {
		q = q.Where("customer_id = ?", *filter.CustomerID)
	}
	if filter.SubscriptionID != nil {
		q = q.Where("subscription_id = ?", *filter.SubscriptionID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	var rows []models.Invoice
	if err := q.Order("created_at DESC").Limit(limit).Offset(filter.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*domain.Invoice, len(rows))
	for i := range rows {
		out[i] = models.InvoiceToDomain(&rows[i])
	}
	return out, total, nil
}

func (r *InvoiceRepo) Update(ctx context.Context, invoice *domain.Invoice) error {
	m, err := models.InvoiceFromDomain(invoice)
	if err != nil {
		return err
	}
	res := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("id = ? AND tenant_id = ?", invoice.ID, invoice.TenantID).Save(m)
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *InvoiceRepo) CreateLineItem(ctx context.Context, item *domain.InvoiceLineItem) error {
	m := models.InvoiceLineItemFromDomain(item)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	item.ID = m.ID
	item.CreatedAt = m.CreatedAt
	return nil
}

func (r *InvoiceRepo) ListLineItems(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]*domain.InvoiceLineItem, error) {
	var rows []models.InvoiceLineItem
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND invoice_id = ?", tenantID, invoiceID).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.InvoiceLineItem, len(rows))
	for i := range rows {
		out[i] = models.InvoiceLineItemToDomain(&rows[i])
	}
	return out, nil
}

func (r *InvoiceRepo) SumPaidByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (int64, error) {
	var sum int64
	err := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ? AND customer_id = ? AND status = ?", tenantID, customerID, domain.InvoiceStatusPaid).
		Select("COALESCE(SUM(amount_paid), 0)").Scan(&sum).Error
	return sum, err
}

var _ domain.InvoiceRepository = (*InvoiceRepo)(nil)
