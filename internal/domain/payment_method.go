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

type MandateStatus string

const (
	MandateStatusPending MandateStatus = "pending"
	MandateStatusReady   MandateStatus = "ready"
	MandateStatusFailed  MandateStatus = "failed"
)

func (pm *PaymentMethod) MandateReady() bool {
	return pm != nil &&
		pm.Type == PaymentMethodDirectDebit &&
		pm.MandateStatus == MandateStatusReady &&
		pm.MandateID != ""
}

type PaymentMethod struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	CustomerID    uuid.UUID
	Type          PaymentMethodType
	TokenKey      string
	MandateID     string
	MandateStatus MandateStatus
	CardLast4     string
	CardBrand     string
	CardExpiry    string
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentMethodRepository interface {
	Create(ctx context.Context, pm *PaymentMethod) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*PaymentMethod, error)
	GetDefaultForCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*PaymentMethod, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	SetDefault(ctx context.Context, tenantID, customerID, id uuid.UUID) error
	ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]*PaymentMethod, error)
	GetDirectDebitForCustomer(ctx context.Context, tenantID, customerID uuid.UUID, preferID *uuid.UUID) (*PaymentMethod, error)
	ListPendingMandates(ctx context.Context, limit int) ([]*PaymentMethod, error)
	Update(ctx context.Context, pm *PaymentMethod) error
}
