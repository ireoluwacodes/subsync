package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WebhookEndpoint struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	URL       string
	Events    []string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WebhookDelivery struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	EndpointID   uuid.UUID
	EventType    string
	Payload      map[string]any
	AttemptCount int
	LastStatus   *int
	LastError    string
	DeliveredAt  *time.Time
	NextRetryAt  *time.Time
	CreatedAt    time.Time
}

type WebhookRepository interface {
	CreateEndpoint(ctx context.Context, ep *WebhookEndpoint) error
	GetEndpoint(ctx context.Context, tenantID, id uuid.UUID) (*WebhookEndpoint, error)
	ListEndpoints(ctx context.Context, tenantID uuid.UUID) ([]*WebhookEndpoint, error)
	UpdateEndpoint(ctx context.Context, ep *WebhookEndpoint) error
	DeleteEndpoint(ctx context.Context, tenantID, id uuid.UUID) error
	CreateDelivery(ctx context.Context, d *WebhookDelivery) error
	ListDeliveries(ctx context.Context, tenantID, endpointID uuid.UUID) ([]*WebhookDelivery, error)
}
