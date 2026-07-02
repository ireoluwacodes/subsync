package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type PlanService struct {
	repo domain.PlanRepository
}

func NewPlanService(repo domain.PlanRepository) *PlanService {
	return &PlanService{repo: repo}
}

type CreatePlanInput struct {
	Name         string
	Description  string
	Amount       int64
	Currency     string
	Interval     domain.PlanInterval
	IntervalDays *int
	TrialDays    int
	Features     []string
	IsActive     bool
}

type UpdatePlanInput struct {
	Name         string
	Description  string
	Amount       int64
	Currency     string
	Interval     domain.PlanInterval
	IntervalDays *int
	TrialDays    int
	Features     []string
	IsActive     bool
}

func (s *PlanService) Create(ctx context.Context, tenantID uuid.UUID, in CreatePlanInput) (*domain.Plan, error) {
	if err := utils.ValidatePlanInput(in.Interval, in.IntervalDays, in.Amount); err != nil {
		return nil, err
	}

	currency := in.Currency
	if currency == "" {
		currency = "NGN"
	}

	plan := &domain.Plan{
		TenantID:     tenantID,
		Name:         in.Name,
		Description:  in.Description,
		Amount:       in.Amount,
		Currency:     currency,
		Interval:     in.Interval,
		IntervalDays: in.IntervalDays,
		TrialDays:    in.TrialDays,
		Features:     in.Features,
		IsActive:     in.IsActive,
	}

	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanService) Get(ctx context.Context, tenantID, id uuid.UUID) (*domain.Plan, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *PlanService) List(ctx context.Context, tenantID uuid.UUID, activeOnly bool, limit, offset int) ([]*domain.Plan, int64, error) {
	plans, err := s.repo.List(ctx, tenantID, activeOnly, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, tenantID, activeOnly)
	if err != nil {
		return nil, 0, err
	}
	return plans, total, nil
}

func (s *PlanService) Update(ctx context.Context, tenantID, id uuid.UUID, in UpdatePlanInput) (*domain.Plan, error) {
	if err := utils.ValidatePlanInput(in.Interval, in.IntervalDays, in.Amount); err != nil {
		return nil, err
	}

	plan, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	plan.Name = in.Name
	plan.Description = in.Description
	plan.Amount = in.Amount
	plan.Currency = in.Currency
	plan.Interval = in.Interval
	plan.IntervalDays = in.IntervalDays
	plan.TrialDays = in.TrialDays
	plan.Features = in.Features
	plan.IsActive = in.IsActive

	if err := s.repo.Update(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanService) Archive(ctx context.Context, tenantID, id uuid.UUID) error {
	referenced, err := s.repo.IsReferenced(ctx, id)
	if err != nil {
		return err
	}
	if referenced {
		return fmt.Errorf("%w: plan is referenced by subscriptions", domain.ErrConflict)
	}
	return s.repo.Archive(ctx, tenantID, id)
}
