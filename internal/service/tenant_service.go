package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
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
	if err := validateNombaInput(in.NombaClientID, in.NombaClientSecret, in.NombaAccountID, in.NombaEnv); err != nil {
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

	apiKey, prefix, hash, err := generateAPIKey()
	if err != nil {
		return nil, err
	}

	webhookSecret, err := generateWebhookSecret()
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
		DunningConfig:     defaultDunningConfig(),
		Branding:          map[string]any{},
		BillingEmail:      map[string]any{},
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	return &CreateTenantResult{Tenant: tenant, APIKey: apiKey}, nil
}

func validateNombaInput(clientID, clientSecret, accountID, nombaEnv string) error {
	if clientID == "" || clientSecret == "" || accountID == "" {
		return fmt.Errorf("%w: nomba_client_id, nomba_client_secret, and nomba_account_id are required", domain.ErrValidation)
	}
	env := strings.ToLower(nombaEnv)
	if env != domain.NombaEnvSandbox && env != domain.NombaEnvProduction {
		return fmt.Errorf("%w: nomba_env must be sandbox or production", domain.ErrValidation)
	}
	return nil
}

func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TenantService) LoadNombaCredentials(ctx context.Context, tenant *domain.Tenant) error {
	if s.loader == nil {
		return fmt.Errorf("tenant credential loader not configured")
	}
	return s.loader.LoadNombaSecret(ctx, tenant)
}

func defaultDunningConfig() map[string]any {
	return map[string]any{
		"steps": []map[string]any{
			{"delay_days": 1, "action": "retry"},
			{"delay_days": 3, "action": "retry_and_notify"},
			{"delay_days": 7, "action": "mandate_fallback"},
			{"delay_days": 14, "action": "cancel"},
		},
	}
}
