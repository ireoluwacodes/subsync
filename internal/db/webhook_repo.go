package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type WebhookRepo struct {
	db *DB
}

func NewWebhookRepo(db *DB) *WebhookRepo {
	return &WebhookRepo{db: db}
}

func (r *WebhookRepo) CreateEndpoint(ctx context.Context, ep *domain.WebhookEndpoint) error {
	return domain.ErrNotImplemented
}

func (r *WebhookRepo) GetEndpoint(ctx context.Context, tenantID, id uuid.UUID) (*domain.WebhookEndpoint, error) {
	return nil, domain.ErrNotImplemented
}

func (r *WebhookRepo) ListEndpoints(ctx context.Context, tenantID uuid.UUID) ([]*domain.WebhookEndpoint, error) {
	return nil, domain.ErrNotImplemented
}

func (r *WebhookRepo) UpdateEndpoint(ctx context.Context, ep *domain.WebhookEndpoint) error {
	return domain.ErrNotImplemented
}

func (r *WebhookRepo) DeleteEndpoint(ctx context.Context, tenantID, id uuid.UUID) error {
	return domain.ErrNotImplemented
}

func (r *WebhookRepo) CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	return domain.ErrNotImplemented
}

func (r *WebhookRepo) ListDeliveries(ctx context.Context, tenantID, endpointID uuid.UUID) ([]*domain.WebhookDelivery, error) {
	return nil, domain.ErrNotImplemented
}

var _ domain.WebhookRepository = (*WebhookRepo)(nil)
