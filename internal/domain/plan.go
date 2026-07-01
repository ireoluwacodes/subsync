package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PlanInterval string

const (
	PlanIntervalMonthly PlanInterval = "monthly"
	PlanIntervalAnnual  PlanInterval = "annual"
	PlanIntervalCustom  PlanInterval = "custom"
)

type Plan struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Name         string
	Description  string
	Amount       int64
	Currency     string
	Interval     PlanInterval
	IntervalDays *int
	TrialDays    int
	Features     []string
	IsActive     bool
	IsArchived   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PlanRepository interface {
	Create(ctx context.Context, plan *Plan) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Plan, error)
	List(ctx context.Context, tenantID uuid.UUID, activeOnly bool, limit, offset int) ([]*Plan, error)
	Count(ctx context.Context, tenantID uuid.UUID, activeOnly bool) (int64, error)
	Update(ctx context.Context, plan *Plan) error
	Archive(ctx context.Context, tenantID, id uuid.UUID) error
	IsReferenced(ctx context.Context, planID uuid.UUID) (bool, error)
}
