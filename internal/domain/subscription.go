package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SubscriptionState string

const (
	SubscriptionStateTrialing SubscriptionState = "trialing"
	SubscriptionStateActive   SubscriptionState = "active"
	SubscriptionStatePastDue  SubscriptionState = "past_due"
	SubscriptionStateCanceled SubscriptionState = "canceled"
	SubscriptionStateExpired  SubscriptionState = "expired"
	SubscriptionStatePaused   SubscriptionState = "paused"
)

type Subscription struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	CustomerID         uuid.UUID
	PlanID             uuid.UUID
	PaymentMethodID    *uuid.UUID
	State              SubscriptionState
	TrialEndsAt        *time.Time
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	NextBillingAt      *time.Time
	CanceledAt         *time.Time
	CancelAtPeriodEnd  bool
	PauseStartsAt      *time.Time
	PauseEndsAt        *time.Time
	DunningStep        int
	DunningStartedAt   *time.Time
	Metadata           map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SubscriptionTransition struct {
	ID             uuid.UUID
	SubscriptionID uuid.UUID
	TenantID       uuid.UUID
	FromState      SubscriptionState
	ToState        SubscriptionState
	Reason         string
	Actor          string
	Metadata       map[string]any
	CreatedAt      time.Time
}

type SubscriptionListFilter struct {
	CustomerID  *uuid.UUID
	PlanID      *uuid.UUID
	State       string
	Limit       int
	Offset      int
	Sort        string
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *Subscription) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Subscription, error)
	List(ctx context.Context, tenantID uuid.UUID, filter SubscriptionListFilter) ([]*Subscription, int64, error)
	Update(ctx context.Context, sub *Subscription) error
	RecordTransition(ctx context.Context, t *SubscriptionTransition) error
	ListTransitions(ctx context.Context, tenantID, subscriptionID uuid.UUID) ([]*SubscriptionTransition, error)
	CountActiveByPlan(ctx context.Context, tenantID, planID uuid.UUID) (int64, error)
	Transition(ctx context.Context, sub *Subscription, t *SubscriptionTransition) error
}
