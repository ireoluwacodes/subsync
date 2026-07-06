package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

type stubPMRepo struct {
	byID      map[uuid.UUID]*domain.PaymentMethod
	customer  []*domain.PaymentMethod
	defaultPM *domain.PaymentMethod
}

func (s *stubPMRepo) Create(ctx context.Context, pm *domain.PaymentMethod) error { return nil }
func (s *stubPMRepo) Update(ctx context.Context, pm *domain.PaymentMethod) error { return nil }
func (s *stubPMRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error  { return nil }
func (s *stubPMRepo) SetDefault(ctx context.Context, tenantID, customerID, id uuid.UUID) error {
	return nil
}
func (s *stubPMRepo) ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]*domain.PaymentMethod, error) {
	return s.customer, nil
}
func (s *stubPMRepo) ListPendingMandates(ctx context.Context, limit int) ([]*domain.PaymentMethod, error) {
	return nil, nil
}
func (s *stubPMRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.PaymentMethod, error) {
	if pm, ok := s.byID[id]; ok {
		return pm, nil
	}
	return nil, domain.ErrNotFound
}
func (s *stubPMRepo) GetDefaultForCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*domain.PaymentMethod, error) {
	if s.defaultPM != nil {
		return s.defaultPM, nil
	}
	return nil, domain.ErrNotFound
}
func (s *stubPMRepo) GetDirectDebitForCustomer(ctx context.Context, tenantID, customerID uuid.UUID, preferID *uuid.UUID) (*domain.PaymentMethod, error) {
	if preferID != nil {
		if pm, err := s.GetByID(ctx, tenantID, *preferID); err == nil && pm.MandateReady() {
			return pm, nil
		}
	}
	for _, pm := range s.customer {
		if pm.Type == domain.PaymentMethodDirectDebit && pm.MandateReady() {
			return pm, nil
		}
	}
	return nil, domain.ErrNotFound
}

func TestPaymentMethodResolver_ResolveMandateIgnoresCardOnSubscription(t *testing.T) {
	cardID := uuid.New()
	mandateID := uuid.New()
	sub := &domain.Subscription{
		CustomerID:      uuid.New(),
		PaymentMethodID:   &cardID,
		FallbackPaymentMethodID: &mandateID,
	}
	repo := &stubPMRepo{
		byID: map[uuid.UUID]*domain.PaymentMethod{
			cardID: {
				ID:       cardID,
				Type:     domain.PaymentMethodTokenizedCard,
				TokenKey: "tok",
			},
			mandateID: {
				ID:            mandateID,
				Type:          domain.PaymentMethodDirectDebit,
				MandateID:     "mand-1",
				MandateStatus: domain.MandateStatusReady,
			},
		},
	}
	r := NewPaymentMethodResolver(repo)

	card, err := r.ResolvePrimaryPM(context.Background(), uuid.New(), sub)
	require.NoError(t, err)
	require.Equal(t, domain.PaymentMethodTokenizedCard, card.Type)

	mandate, err := r.ResolveMandatePM(context.Background(), uuid.New(), sub)
	require.NoError(t, err)
	require.Equal(t, mandateID, mandate.ID)
}

func TestPaymentMethodResolver_HasChargeablePM(t *testing.T) {
	mandateID := uuid.New()
	sub := &domain.Subscription{
		CustomerID:              uuid.New(),
		FallbackPaymentMethodID: &mandateID,
	}
	repo := &stubPMRepo{
		byID: map[uuid.UUID]*domain.PaymentMethod{
			mandateID: {
				ID:            mandateID,
				Type:          domain.PaymentMethodDirectDebit,
				MandateID:     "mand-1",
				MandateStatus: domain.MandateStatusReady,
			},
		},
	}
	r := NewPaymentMethodResolver(repo)
	require.True(t, r.HasChargeablePM(context.Background(), uuid.New(), sub))
}
