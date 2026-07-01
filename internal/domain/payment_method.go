package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentMethodType string

const (
	PaymentMethodTokenizedCard PaymentMethodType = "tokenized_card"
	PaymentMethodDirectDebit   PaymentMethodType = "direct_debit"
)

type PaymentMethod struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	Type       PaymentMethodType
	TokenKey   string
	MandateID  string
	CardLast4  string
	CardBrand  string
	CardExpiry string
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PaymentMethodRepository interface {
	Create(ctx context.Context, pm *PaymentMethod) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*PaymentMethod, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	SetDefault(ctx context.Context, tenantID, customerID, id uuid.UUID) error
}
