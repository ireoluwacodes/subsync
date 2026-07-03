package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type TenantService struct {
	repo   domain.TenantRepository
	nomba  *nomba.Client
	loader interface {
		LoadNombaSecret(ctx context.Context, tenant *domain.Tenant) error
	}
}

func NewTenantService(repo domain.TenantRepository, nombaClient *nomba.Client) *TenantService {
	s := &TenantService{repo: repo, nomba: nombaClient}
	if l, ok := repo.(interface {
		LoadNombaSecret(ctx context.Context, tenant *domain.Tenant) error
	}); ok {
		s.loader = l
	}
	return s
}

type CreateTenantInput struct {
	Name              string
	Email             string
	NombaClientID     string
	NombaClientSecret string
	NombaAccountID    string
	NombaSubAccountID   string
	NombaEnv            string
	NombaWebhookSecret  string
}

type CreateTenantResult struct {
	Tenant *domain.Tenant
	APIKey string
}

func (s *TenantService) CreateTenant(ctx context.Context, in CreateTenantInput) (*CreateTenantResult, error) {
	if err := utils.ValidateNombaInput(in.NombaClientID, in.NombaClientSecret, in.NombaAccountID, in.NombaEnv); err != nil {
		return nil, err
	}
	if in.Name == "" || in.Email == "" {
		return nil, fmt.Errorf("%w: name and email are required", domain.ErrValidation)
	}

	if s.nomba != nil {
		if err := s.nomba.ValidateCredentials(ctx, in.NombaEnv, in.NombaClientID, in.NombaClientSecret, in.NombaAccountID); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrInvalidNombaCredentials, err)
		}
	}

	apiKey, prefix, hash, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	webhookSecret, err := utils.GenerateWebhookSecret()
	if err != nil {
		return nil, err
	}

	tenant := &domain.Tenant{
		Name:              in.Name,
		Email:             in.Email,
		NombaClientID:     in.NombaClientID,
		NombaClientSecret: in.NombaClientSecret,
		NombaAccountID:    in.NombaAccountID,
		NombaSubAccountID: in.NombaSubAccountID,
		NombaEnv:          in.NombaEnv,
		NombaWebhookSecret: in.NombaWebhookSecret,
		APIKeyPrefix:      prefix,
		APIKeyHash:        hash,
		WebhookSecret:     webhookSecret,
		DunningConfig:     utils.DefaultDunningConfig(),
		Branding:          map[string]any{},
		BillingEmail:      map[string]any{},
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	return &CreateTenantResult{Tenant: tenant, APIKey: apiKey}, nil
}

func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.GetTenant(ctx, id)
}

func (s *TenantService) LoadNombaWebhookSecret(ctx context.Context, tenant *domain.Tenant) error {
	if l, ok := s.repo.(interface {
		LoadNombaWebhookSecret(ctx context.Context, tenant *domain.Tenant) error
	}); ok {
		return l.LoadNombaWebhookSecret(ctx, tenant)
	}
	return fmt.Errorf("tenant webhook secret loader not configured")
}

func (s *TenantService) LoadNombaCredentials(ctx context.Context, tenant *domain.Tenant) error {
	if s.loader == nil {
		return fmt.Errorf("tenant credential loader not configured")
	}
	return s.loader.LoadNombaSecret(ctx, tenant)
}
