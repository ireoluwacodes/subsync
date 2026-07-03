package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	WebhookEventSubscriptionCreated    = "subscription.created"
	WebhookEventSubscriptionUpdated    = "subscription.updated"
	WebhookEventSubscriptionCanceled   = "subscription.canceled"
	WebhookEventInvoiceCreated         = "invoice.created"
	WebhookEventInvoicePaid            = "invoice.paid"
	WebhookEventInvoicePaymentFailed   = "invoice.payment_failed"
	WebhookEventPaymentMethodAttached  = "payment_method.attached"
)

const MaxWebhookEndpointsPerTenant = 5

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
	CountEndpoints(ctx context.Context, tenantID uuid.UUID) (int64, error)
	UpdateEndpoint(ctx context.Context, ep *WebhookEndpoint) error
	DeleteEndpoint(ctx context.Context, tenantID, id uuid.UUID) error
	CreateDelivery(ctx context.Context, d *WebhookDelivery) error
	GetDelivery(ctx context.Context, tenantID, id uuid.UUID) (*WebhookDelivery, error)
	UpdateDelivery(ctx context.Context, d *WebhookDelivery) error
	ListDeliveries(ctx context.Context, tenantID, endpointID uuid.UUID) ([]*WebhookDelivery, error)
	ListPendingDeliveries(ctx context.Context, before time.Time, limit int) ([]*WebhookDelivery, error)
}
