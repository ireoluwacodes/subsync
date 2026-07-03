package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PortalToken struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	SubscriptionID uuid.UUID
	CustomerID     uuid.UUID
	TokenHash      string
	ExpiresAt      time.Time
	UsedAt         *time.Time
	CreatedAt      time.Time
}

type PortalTokenRepository interface {
	Create(ctx context.Context, token *PortalToken) error
	GetByID(ctx context.Context, id uuid.UUID) (*PortalToken, error)
	GetValidByTokenHash(ctx context.Context, tokenHash string) (*PortalToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}
