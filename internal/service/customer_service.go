package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type CustomerService struct {
	repo           domain.CustomerRepository
	subscriptions  domain.SubscriptionRepository
	paymentMethods domain.PaymentMethodRepository
}

func NewCustomerService(repo domain.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) SetSubscriptions(subs domain.SubscriptionRepository) {
	s.subscriptions = subs
}

func (s *CustomerService) SetPaymentMethods(pms domain.PaymentMethodRepository) {
	s.paymentMethods = pms
}

// CustomerWithRelations bundles a customer with its subscriptions and payment methods.
type CustomerWithRelations struct {
	Customer       *domain.Customer
	Subscriptions  []*domain.Subscription
	PaymentMethods []*domain.PaymentMethod
}

type CreateCustomerInput struct {
	ExternalID string
	Email      string
	Name       string
	Phone      string
	Metadata   map[string]any
}

type UpdateCustomerInput struct {
	ExternalID string
	Email      string
	Name       string
	Phone      string
	Metadata   map[string]any
}

func (s *CustomerService) Create(ctx context.Context, tenantID uuid.UUID, in CreateCustomerInput) (*domain.Customer, error) {
	if in.Email == "" {
		return nil, fmt.Errorf("%w: email is required", domain.ErrValidation)
	}

	customer := &domain.Customer{
		TenantID:   tenantID,
		ExternalID: in.ExternalID,
		Email:      in.Email,
		Name:       in.Name,
		Phone:      in.Phone,
		Metadata:   in.Metadata,
	}
	if customer.Metadata == nil {
		customer.Metadata = map[string]any{}
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) Get(ctx context.Context, tenantID, id uuid.UUID) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// GetWithRelations returns a customer with their subscriptions and payment methods populated.
func (s *CustomerService) GetWithRelations(ctx context.Context, tenantID, id uuid.UUID) (*CustomerWithRelations, error) {
	customer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	rel := &CustomerWithRelations{Customer: customer}
	if s.subscriptions != nil {
		if subs, _, err := s.subscriptions.List(ctx, tenantID, domain.SubscriptionListFilter{CustomerID: &id, Limit: 100}); err == nil {
			rel.Subscriptions = subs
		}
	}
	if s.paymentMethods != nil {
		if pms, err := s.paymentMethods.ListByCustomer(ctx, tenantID, id); err == nil {
			rel.PaymentMethods = pms
		}
	}
	return rel, nil
}

func (s *CustomerService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Customer, int64, error) {
	customers, err := s.repo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func (s *CustomerService) Update(ctx context.Context, tenantID, id uuid.UUID, in UpdateCustomerInput) (*domain.Customer, error) {
	if in.Email == "" {
		return nil, fmt.Errorf("%w: email is required", domain.ErrValidation)
	}

	customer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	customer.ExternalID = in.ExternalID
	customer.Email = in.Email
	customer.Name = in.Name
	customer.Phone = in.Phone
	if in.Metadata != nil {
		customer.Metadata = in.Metadata
	}

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}
	return customer, nil
}
