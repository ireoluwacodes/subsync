package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type PaymentMethodResolver struct {
	repo domain.PaymentMethodRepository
}

func NewPaymentMethodResolver(repo domain.PaymentMethodRepository) *PaymentMethodResolver {
	return &PaymentMethodResolver{repo: repo}
}

func (r *PaymentMethodResolver) ResolvePrimaryPM(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription) (*domain.PaymentMethod, error) {
	if sub.PaymentMethodID != nil {
		pm, err := r.repo.GetByID(ctx, tenantID, *sub.PaymentMethodID)
		if err != nil {
			return nil, err
		}
		if pm.Type == domain.PaymentMethodTokenizedCard && pm.TokenKey != "" {
			return pm, nil
		}
	}
	pms, err := r.repo.ListByCustomer(ctx, tenantID, sub.CustomerID)
	if err != nil {
		return nil, err
	}
	for _, pm := range pms {
		if pm.Type == domain.PaymentMethodTokenizedCard && pm.TokenKey != "" {
			return pm, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *PaymentMethodResolver) ResolveMandatePM(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription) (*domain.PaymentMethod, error) {
	return r.repo.GetDirectDebitForCustomer(ctx, tenantID, sub.CustomerID, sub.FallbackPaymentMethodID)
}

func (r *PaymentMethodResolver) HasChargeablePM(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription) bool {
	if pm, err := r.ResolvePrimaryPM(ctx, tenantID, sub); err == nil && pm != nil {
		return true
	}
	if pm, err := r.ResolveMandatePM(ctx, tenantID, sub); err == nil && pm != nil && pm.MandateReady() {
		return true
	}
	return false
}

func (r *PaymentMethodResolver) HasPendingMandate(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription) bool {
	if sub.FallbackPaymentMethodID != nil {
		pm, err := r.repo.GetByID(ctx, tenantID, *sub.FallbackPaymentMethodID)
		if err == nil && pm.Type == domain.PaymentMethodDirectDebit && pm.MandateStatus == domain.MandateStatusPending {
			return true
		}
	}
	pms, err := r.repo.ListByCustomer(ctx, tenantID, sub.CustomerID)
	if err != nil {
		return false
	}
	for _, pm := range pms {
		if pm.Type == domain.PaymentMethodDirectDebit && pm.MandateStatus == domain.MandateStatusPending {
			return true
		}
	}
	return false
}

func (r *PaymentMethodResolver) CustomerHasDefaultCard(ctx context.Context, tenantID, customerID uuid.UUID) bool {
	pm, err := r.repo.GetDefaultForCustomer(ctx, tenantID, customerID)
	return err == nil && pm != nil && pm.Type == domain.PaymentMethodTokenizedCard
}
