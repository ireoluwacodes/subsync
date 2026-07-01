package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	ExternalID string
	Email      string
	Name       string
	Phone      string
	Metadata   map[string]any
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Customer, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Customer, error)
	Count(ctx context.Context, tenantID uuid.UUID) (int64, error)
	Update(ctx context.Context, customer *Customer) error
}
