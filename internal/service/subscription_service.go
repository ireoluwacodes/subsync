package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type SubscriptionService struct {
	repo      domain.SubscriptionRepository
	plans     domain.PlanRepository
	customers domain.CustomerRepository
	invoices  *InvoiceService
}

func NewSubscriptionService(
	repo domain.SubscriptionRepository,
	plans domain.PlanRepository,
	customers domain.CustomerRepository,
	invoices *InvoiceService,
) *SubscriptionService {
	return &SubscriptionService{repo: repo, plans: plans, customers: customers, invoices: invoices}
}

type CreateSubscriptionInput struct {
	CustomerID      uuid.UUID
	PlanID          uuid.UUID
	PaymentMethodID *uuid.UUID
}

func (s *SubscriptionService) Create(ctx context.Context, tenantID uuid.UUID, in CreateSubscriptionInput) (*domain.Subscription, error) {
	plan, err := s.plans.GetByID(ctx, tenantID, in.PlanID)
	if err != nil {
		return nil, err
	}
	if _, err := s.customers.GetByID(ctx, tenantID, in.CustomerID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	periodEnd := addPlanPeriod(now, plan)

	sub := &domain.Subscription{
		TenantID:           tenantID,
		CustomerID:         in.CustomerID,
		PlanID:             in.PlanID,
		PaymentMethodID:    in.PaymentMethodID,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		Metadata:           map[string]any{},
	}

	if plan.TrialDays > 0 {
		trialEnd := now.AddDate(0, 0, plan.TrialDays)
		sub.State = domain.SubscriptionStateTrialing
		sub.TrialEndsAt = &trialEnd
		sub.NextBillingAt = &trialEnd
	} else {
		sub.State = domain.SubscriptionStateActive
		sub.NextBillingAt = &periodEnd
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	_ = s.repo.RecordTransition(ctx, &domain.SubscriptionTransition{
		SubscriptionID: sub.ID,
		TenantID:       tenantID,
		FromState:      "",
		ToState:        sub.State,
		Reason:         "created",
		Actor:          "system",
		Metadata:       map[string]any{},
	})

	return sub, nil
}

func addPlanPeriod(start time.Time, plan *domain.Plan) time.Time {
	switch plan.Interval {
	case domain.PlanIntervalAnnual:
		return start.AddDate(1, 0, 0)
	case domain.PlanIntervalCustom:
		days := 30
		if plan.IntervalDays != nil {
			days = *plan.IntervalDays
		}
		return start.AddDate(0, 0, days)
	default:
		return start.AddDate(0, 1, 0)
	}
}

func (s *SubscriptionService) Get(ctx context.Context, tenantID, id uuid.UUID) (*domain.Subscription, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *SubscriptionService) List(ctx context.Context, tenantID uuid.UUID, filter domain.SubscriptionListFilter) ([]*domain.Subscription, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

type CancelInput struct {
	CancelAtPeriodEnd bool
	Reason            string
}

func (s *SubscriptionService) Cancel(ctx context.Context, tenantID, id uuid.UUID, in CancelInput, actor string) (*domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	from := sub.State
	if in.CancelAtPeriodEnd {
		sub.CancelAtPeriodEnd = true
		if err := s.repo.Update(ctx, sub); err != nil {
			return nil, err
		}
		return sub, nil
	}

	if err := domain.ValidateTransition(from, domain.SubscriptionStateCanceled); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	sub.State = domain.SubscriptionStateCanceled
	sub.CanceledAt = &now
	sub.CancelAtPeriodEnd = false

	return s.applyTransition(ctx, sub, from, domain.SubscriptionStateCanceled, in.Reason, actor)
}

func (s *SubscriptionService) Pause(ctx context.Context, tenantID, id uuid.UUID, pauseEndsAt *time.Time, actor string) (*domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	from := sub.State
	if err := domain.ValidateTransition(from, domain.SubscriptionStatePaused); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	sub.State = domain.SubscriptionStatePaused
	sub.PauseStartsAt = &now
	sub.PauseEndsAt = pauseEndsAt
	return s.applyTransition(ctx, sub, from, domain.SubscriptionStatePaused, "paused", actor)
}

func (s *SubscriptionService) Resume(ctx context.Context, tenantID, id uuid.UUID, actor string) (*domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	from := sub.State
	if err := domain.ValidateTransition(from, domain.SubscriptionStateActive); err != nil {
		return nil, err
	}
	sub.State = domain.SubscriptionStateActive
	sub.PauseStartsAt = nil
	sub.PauseEndsAt = nil
	next := sub.CurrentPeriodEnd
	sub.NextBillingAt = &next
	return s.applyTransition(ctx, sub, from, domain.SubscriptionStateActive, "resumed", actor)
}

type UpgradeInput struct {
	NewPlanID uuid.UUID
}

func (s *SubscriptionService) PreviewUpgrade(ctx context.Context, tenantID, id uuid.UUID, in UpgradeInput) (*domain.ProrationResult, error) {
	sub, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	oldPlan, err := s.plans.GetByID(ctx, tenantID, sub.PlanID)
	if err != nil {
		return nil, err
	}
	newPlan, err := s.plans.GetByID(ctx, tenantID, in.NewPlanID)
	if err != nil {
		return nil, err
	}

	result := domain.CalculateProration(oldPlan.Amount, newPlan.Amount, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, time.Now().UTC())
	return &result, nil
}

func (s *SubscriptionService) Upgrade(ctx context.Context, tenantID, id uuid.UUID, in UpgradeInput, actor string) (*domain.Subscription, *domain.Invoice, error) {
	sub, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	oldPlan, err := s.plans.GetByID(ctx, tenantID, sub.PlanID)
	if err != nil {
		return nil, nil, err
	}
	newPlan, err := s.plans.GetByID(ctx, tenantID, in.NewPlanID)
	if err != nil {
		return nil, nil, err
	}

	proration := domain.CalculateProration(oldPlan.Amount, newPlan.Amount, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, time.Now().UTC())
	sub.PlanID = newPlan.ID

	invoice, err := s.invoices.CreateUpgradeInvoice(ctx, tenantID, sub, proration, oldPlan, newPlan)
	if err != nil {
		return nil, nil, err
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, nil, err
	}

	_ = s.repo.RecordTransition(ctx, &domain.SubscriptionTransition{
		SubscriptionID: sub.ID,
		TenantID:       tenantID,
		FromState:      sub.State,
		ToState:        sub.State,
		Reason:         fmt.Sprintf("upgraded to plan %s", newPlan.ID),
		Actor:          actor,
		Metadata:       map[string]any{"invoice_id": invoice.ID.String()},
	})

	return sub, invoice, nil
}

func (s *SubscriptionService) ListTransitions(ctx context.Context, tenantID, subscriptionID uuid.UUID) ([]*domain.SubscriptionTransition, error) {
	return s.repo.ListTransitions(ctx, tenantID, subscriptionID)
}

func (s *SubscriptionService) PlanStats(ctx context.Context, tenantID, planID uuid.UUID) (int64, error) {
	return s.repo.CountActiveByPlan(ctx, tenantID, planID)
}

func (s *SubscriptionService) applyTransition(ctx context.Context, sub *domain.Subscription, from, to domain.SubscriptionState, reason, actor string) (*domain.Subscription, error) {
	tr := &domain.SubscriptionTransition{
		SubscriptionID: sub.ID,
		TenantID:       sub.TenantID,
		FromState:      from,
		ToState:        to,
		Reason:         reason,
		Actor:          actor,
		Metadata:       map[string]any{},
	}
	if err := s.repo.Transition(ctx, sub, tr); err != nil {
		return nil, err
	}
	return sub, nil
}
