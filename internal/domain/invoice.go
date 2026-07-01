package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusVoid          InvoiceStatus = "void"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
)

type LineItemType string

const (
	LineItemSubscription    LineItemType = "subscription"
	LineItemProrationCredit LineItemType = "proration_credit"
	LineItemProrationDebit  LineItemType = "proration_debit"
)

type Invoice struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	SubscriptionID     uuid.UUID
	CustomerID         uuid.UUID
	Status             InvoiceStatus
	AmountDue          int64
	AmountPaid         int64
	Currency           string
	PeriodStart        time.Time
	PeriodEnd          time.Time
	DueDate            *time.Time
	PaidAt             *time.Time
	VoidedAt           *time.Time
	NombaOrderRef      string
	NombaTransactionID string
	AttemptCount       int
	NextAttemptAt      *time.Time
	Metadata           map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type InvoiceLineItem struct {
	ID          uuid.UUID
	InvoiceID   uuid.UUID
	TenantID    uuid.UUID
	Type        LineItemType
	Description string
	Amount      int64
	Currency    string
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	CreatedAt   time.Time
}

type InvoiceListFilter struct {
	CustomerID     *uuid.UUID
	SubscriptionID *uuid.UUID
	Status         string
	Limit          int
	Offset         int
}

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *Invoice) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Invoice, error)
	List(ctx context.Context, tenantID uuid.UUID, filter InvoiceListFilter) ([]*Invoice, int64, error)
	Update(ctx context.Context, invoice *Invoice) error
	CreateLineItem(ctx context.Context, item *InvoiceLineItem) error
	ListLineItems(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]*InvoiceLineItem, error)
	SumPaidByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (int64, error)
}
