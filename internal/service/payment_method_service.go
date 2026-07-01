package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type PaymentMethodService struct {
	repo         domain.PaymentMethodRepository
	customerRepo domain.CustomerRepository
}

func NewPaymentMethodService(repo domain.PaymentMethodRepository, customerRepo domain.CustomerRepository) *PaymentMethodService {
	return &PaymentMethodService{repo: repo, customerRepo: customerRepo}
}

type CreatePaymentMethodInput struct {
	CustomerID uuid.UUID
	Type       domain.PaymentMethodType
	TokenKey   string
	MandateID  string
	CardLast4  string
	CardBrand  string
	CardExpiry string
	IsDefault  bool
}

func (s *PaymentMethodService) Create(ctx context.Context, tenantID uuid.UUID, in CreatePaymentMethodInput) (*domain.PaymentMethod, error) {
	if err := validatePaymentMethodInput(in); err != nil {
		return nil, err
	}

	if _, err := s.customerRepo.GetByID(ctx, tenantID, in.CustomerID); err != nil {
		return nil, err
	}

	pm := &domain.PaymentMethod{
		TenantID:   tenantID,
		CustomerID: in.CustomerID,
		Type:       in.Type,
		TokenKey:   in.TokenKey,
		MandateID:  in.MandateID,
		CardLast4:  in.CardLast4,
		CardBrand:  in.CardBrand,
		CardExpiry: in.CardExpiry,
		IsDefault:  in.IsDefault,
	}

	if err := s.repo.Create(ctx, pm); err != nil {
		return nil, err
	}

	if in.IsDefault {
		if err := s.repo.SetDefault(ctx, tenantID, in.CustomerID, pm.ID); err != nil {
			return nil, err
		}
		pm.IsDefault = true
	}

	return pm, nil
}

func (s *PaymentMethodService) Get(ctx context.Context, tenantID, id uuid.UUID) (*domain.PaymentMethod, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *PaymentMethodService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

func (s *PaymentMethodService) SetDefault(ctx context.Context, tenantID, id uuid.UUID) (*domain.PaymentMethod, error) {
	pm, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SetDefault(ctx, tenantID, pm.CustomerID, id); err != nil {
		return nil, err
	}

	pm.IsDefault = true
	return pm, nil
}

func validatePaymentMethodInput(in CreatePaymentMethodInput) error {
	switch in.Type {
	case domain.PaymentMethodTokenizedCard:
		if in.TokenKey == "" {
			return fmt.Errorf("%w: token_key is required for tokenized_card", domain.ErrValidation)
		}
	case domain.PaymentMethodDirectDebit:
		if in.MandateID == "" {
			return fmt.Errorf("%w: mandate_id is required for direct_debit", domain.ErrValidation)
		}
	default:
		return fmt.Errorf("%w: invalid payment method type", domain.ErrValidation)
	}
	return nil
}
