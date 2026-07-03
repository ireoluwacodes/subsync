package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type NombaEvent struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	EventID     string
	EventType   string
	Payload     map[string]any
	Processed   bool
	ProcessedAt *time.Time
	Error       string
	CreatedAt   time.Time
}

type NombaEventRepository interface {
	Create(ctx context.Context, event *NombaEvent) error
	GetByEventID(ctx context.Context, tenantID uuid.UUID, eventID string) (*NombaEvent, error)
	MarkProcessed(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}
