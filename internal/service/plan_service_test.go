package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type mockPlanRepo struct {
	plan       *domain.Plan
	referenced bool
}

func (m *mockPlanRepo) Create(ctx context.Context, plan *domain.Plan) error { return nil }
func (m *mockPlanRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Plan, error) {
	return m.plan, nil
}
func (m *mockPlanRepo) List(ctx context.Context, tenantID uuid.UUID, activeOnly bool, limit, offset int) ([]*domain.Plan, error) {
	return nil, nil
}
func (m *mockPlanRepo) Count(ctx context.Context, tenantID uuid.UUID, activeOnly bool) (int64, error) {
	return 0, nil
}
func (m *mockPlanRepo) Update(ctx context.Context, plan *domain.Plan) error { return nil }
func (m *mockPlanRepo) Archive(ctx context.Context, tenantID, id uuid.UUID) error { return nil }
func (m *mockPlanRepo) IsReferenced(ctx context.Context, planID uuid.UUID) (bool, error) {
	return m.referenced, nil
}

func TestPlanService_ValidateInterval(t *testing.T) {
	svc := NewPlanService(&mockPlanRepo{})

	_, err := svc.Create(context.Background(), uuid.New(), CreatePlanInput{
		Name:     "Basic",
		Amount:   1000,
		Interval: domain.PlanIntervalMonthly,
	})
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), uuid.New(), CreatePlanInput{
		Name:     "Bad",
		Amount:   1000,
		Interval: domain.PlanIntervalCustom,
	})
	require.ErrorIs(t, err, domain.ErrValidation)

	days := 30
	_, err = svc.Create(context.Background(), uuid.New(), CreatePlanInput{
		Name:         "Custom",
		Amount:       1000,
		Interval:     domain.PlanIntervalCustom,
		IntervalDays: &days,
	})
	require.NoError(t, err)
}

func TestPlanService_ArchiveReferenced(t *testing.T) {
	planID := uuid.New()
	svc := NewPlanService(&mockPlanRepo{
		plan:       &domain.Plan{ID: planID},
		referenced: true,
	})

	err := svc.Archive(context.Background(), uuid.New(), planID)
	require.ErrorIs(t, err, domain.ErrConflict)
}
