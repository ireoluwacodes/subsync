package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type SettingsService struct {
	tenants       domain.TenantRepository
	nomba         *nomba.Client
	publicBaseURL string
	loader        interface {
		LoadNombaSecret(ctx context.Context, tenant *domain.Tenant) error
	}
}

func NewSettingsService(tenants domain.TenantRepository, nombaClient *nomba.Client, publicBaseURL string) *SettingsService {
	s := &SettingsService{
		tenants:       tenants,
		nomba:         nombaClient,
		publicBaseURL: publicBaseURL,
	}
	if l, ok := tenants.(interface {
		LoadNombaSecret(ctx context.Context, tenant *domain.Tenant) error
	}); ok {
		s.loader = l
	}
	return s
}

type NombaSettingsView struct {
	WebhookURL                string
	NombaClientID             string
	NombaAccountID            string
	NombaSubAccountID         string
	NombaEnv                  string
	NombaWebhookSecretConfigured bool
}

type UpdateGeneralInput struct {
	Name    string
	Email   string
	Website string
}

type UpdateNombaInput struct {
	NombaClientID      string `json:"nomba_client_id"`
	NombaClientSecret  string `json:"nomba_client_secret"`
	NombaAccountID     string `json:"nomba_account_id"`
	NombaSubAccountID  string `json:"nomba_sub_account_id"`
	NombaEnv           string `json:"nomba_env"`
	NombaWebhookSecret string `json:"nomba_webhook_secret"`
}

func (s *SettingsService) Get(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	return s.tenants.GetByID(ctx, tenantID)
}

func (s *SettingsService) GetNomba(ctx context.Context, tenantID uuid.UUID) (*NombaSettingsView, error) {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return s.nombaView(tenant), nil
}

func (s *SettingsService) nombaView(tenant *domain.Tenant) *NombaSettingsView {
	return &NombaSettingsView{
		WebhookURL:                   utils.NombaWebhookURL(s.publicBaseURL, tenant.ID),
		NombaClientID:                utils.MaskClientID(tenant.NombaClientID),
		NombaAccountID:               tenant.NombaAccountID,
		NombaSubAccountID:            tenant.NombaSubAccountID,
		NombaEnv:                     tenant.NombaEnv,
		NombaWebhookSecretConfigured: tenant.HasNombaWebhookSecret,
	}
}

func (s *SettingsService) UpdateGeneral(ctx context.Context, tenantID uuid.UUID, in UpdateGeneralInput) (*domain.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if in.Name != "" {
		tenant.Name = in.Name
	}
	if in.Email != "" {
		tenant.Email = in.Email
	}
	tenant.Website = in.Website
	tenant.UpdatedAt = time.Now()
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *SettingsService) UpdateDunning(ctx context.Context, tenantID uuid.UUID, config map[string]any) (*domain.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	tenant.DunningConfig = config
	tenant.UpdatedAt = time.Now()
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *SettingsService) UpdateBranding(ctx context.Context, tenantID uuid.UUID, branding map[string]any) (*domain.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	tenant.Branding = branding
	tenant.UpdatedAt = time.Now()
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *SettingsService) UpdateBillingEmail(ctx context.Context, tenantID uuid.UUID, billing map[string]any) (*domain.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	tenant.BillingEmail = billing
	tenant.UpdatedAt = time.Now()
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *SettingsService) UpdateNomba(ctx context.Context, tenantID uuid.UUID, in UpdateNombaInput) (*NombaSettingsView, error) {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	onlyWebhookSecret := in.NombaWebhookSecret != "" &&
		in.NombaClientID == "" && in.NombaClientSecret == "" &&
		in.NombaAccountID == "" && in.NombaEnv == "" && in.NombaSubAccountID == ""

	if onlyWebhookSecret {
		tenant.NombaWebhookSecret = in.NombaWebhookSecret
		tenant.UpdatedAt = time.Now()
		if err := s.tenants.Update(ctx, tenant); err != nil {
			return nil, err
		}
		return s.nombaView(tenant), nil
	}

	secret := in.NombaClientSecret
	if secret == "" && s.loader != nil {
		if err := s.loader.LoadNombaSecret(ctx, tenant); err != nil {
			return nil, err
		}
		secret = tenant.NombaClientSecret
	}

	clientID := in.NombaClientID
	if clientID == "" {
		clientID = tenant.NombaClientID
	}
	accountID := in.NombaAccountID
	if accountID == "" {
		accountID = tenant.NombaAccountID
	}
	env := in.NombaEnv
	if env == "" {
		env = tenant.NombaEnv
	}

	if err := utils.ValidateNombaInput(clientID, secret, accountID, env); err != nil {
		return nil, err
	}

	if s.nomba != nil {
		if err := s.nomba.ValidateCredentials(ctx, env, clientID, secret, accountID); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrInvalidNombaCredentials, err)
		}
	}

	tenant.NombaClientID = clientID
	tenant.NombaClientSecret = secret
	tenant.NombaAccountID = accountID
	if in.NombaSubAccountID != "" {
		tenant.NombaSubAccountID = in.NombaSubAccountID
	}
	tenant.NombaEnv = strings.ToLower(env)
	if in.NombaWebhookSecret != "" {
		tenant.NombaWebhookSecret = in.NombaWebhookSecret
	}
	tenant.UpdatedAt = time.Now()

	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	if s.nomba != nil {
		s.nomba.InvalidateTenant(tenantID)
	}
	return s.nombaView(tenant), nil
}

func (s *SettingsService) RotateAPIKey(ctx context.Context, tenantID uuid.UUID) (string, *domain.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return "", nil, err
	}

	apiKey, prefix, hash, err := utils.GenerateAPIKey()
	if err != nil {
		return "", nil, err
	}
	if s.loader != nil {
		if err := s.loader.LoadNombaSecret(ctx, tenant); err != nil {
			return "", nil, err
		}
	}

	tenant.APIKeyPrefix = prefix
	tenant.APIKeyHash = hash
	tenant.UpdatedAt = time.Now()

	if err := s.tenants.Update(ctx, tenant); err != nil {
		return "", nil, err
	}
	tenant.NombaClientSecret = ""
	return apiKey, tenant, nil
}
