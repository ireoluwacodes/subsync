package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type WebhookRepo struct {
	db *DB
}

func NewWebhookRepo(db *DB) *WebhookRepo {
	return &WebhookRepo{db: db}
}

func (r *WebhookRepo) CreateEndpoint(ctx context.Context, ep *domain.WebhookEndpoint) error {
	m := models.WebhookEndpointFromDomain(ep)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	*ep = *models.WebhookEndpointToDomain(m)
	return nil
}

func (r *WebhookRepo) GetEndpoint(ctx context.Context, tenantID, id uuid.UUID) (*domain.WebhookEndpoint, error) {
	var m models.WebhookEndpoint
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&m).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.WebhookEndpointToDomain(&m), nil
}

func (r *WebhookRepo) ListEndpoints(ctx context.Context, tenantID uuid.UUID) ([]*domain.WebhookEndpoint, error) {
	var rows []models.WebhookEndpoint
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.WebhookEndpoint, len(rows))
	for i := range rows {
		out[i] = models.WebhookEndpointToDomain(&rows[i])
	}
	return out, nil
}

func (r *WebhookRepo) CountEndpoints(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.WebhookEndpoint{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

func (r *WebhookRepo) UpdateEndpoint(ctx context.Context, ep *domain.WebhookEndpoint) error {
	m := models.WebhookEndpointFromDomain(ep)
	res := r.db.WithContext(ctx).Model(&models.WebhookEndpoint{}).
		Where("id = ? AND tenant_id = ?", ep.ID, ep.TenantID).
		Updates(map[string]any{
			"url":        m.URL,
			"events":     m.Events,
			"is_active":  m.IsActive,
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WebhookRepo) DeleteEndpoint(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&models.WebhookEndpoint{})
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WebhookRepo) CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	m, err := models.WebhookDeliveryFromDomain(d)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	*d = *models.WebhookDeliveryToDomain(m)
	return nil
}

func (r *WebhookRepo) GetDelivery(ctx context.Context, tenantID, id uuid.UUID) (*domain.WebhookDelivery, error) {
	var m models.WebhookDelivery
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&m).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.WebhookDeliveryToDomain(&m), nil
}

func (r *WebhookRepo) UpdateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	m, err := models.WebhookDeliveryFromDomain(d)
	if err != nil {
		return err
	}
	res := r.db.WithContext(ctx).Model(&models.WebhookDelivery{}).
		Where("id = ? AND tenant_id = ?", d.ID, d.TenantID).
		Updates(map[string]any{
			"attempt_count": m.AttemptCount,
			"last_status":   m.LastStatus,
			"last_error":    m.LastError,
			"delivered_at":  m.DeliveredAt,
			"next_retry_at": m.NextRetryAt,
		})
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WebhookRepo) ListDeliveries(ctx context.Context, tenantID, endpointID uuid.UUID) ([]*domain.WebhookDelivery, error) {
	var rows []models.WebhookDelivery
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND endpoint_id = ?", tenantID, endpointID).
		Order("created_at DESC").
		Limit(50).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.WebhookDelivery, len(rows))
	for i := range rows {
		out[i] = models.WebhookDeliveryToDomain(&rows[i])
	}
	return out, nil
}

func (r *WebhookRepo) ListPendingDeliveries(ctx context.Context, before time.Time, limit int) ([]*domain.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []models.WebhookDelivery
	err := r.db.WithContext(ctx).
		Where("delivered_at IS NULL AND next_retry_at IS NOT NULL AND next_retry_at <= ?", before).
		Order("next_retry_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.WebhookDelivery, len(rows))
	for i := range rows {
		out[i] = models.WebhookDeliveryToDomain(&rows[i])
	}
	return out, nil
}

var _ domain.WebhookRepository = (*WebhookRepo)(nil)
