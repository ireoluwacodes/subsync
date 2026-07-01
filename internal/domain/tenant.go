package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	NombaEnvSandbox    = "sandbox"
	NombaEnvProduction = "production"
)

type Tenant struct {
	ID                uuid.UUID
	Name              string
	Email             string
	Website           string
	NombaClientID     string
	NombaClientSecret string // plaintext in memory only when explicitly loaded
	NombaAccountID    string
	NombaSubAccountID string
	NombaEnv                  string
	NombaWebhookSecret        string // plaintext in memory only when explicitly loaded
	HasNombaWebhookSecret     bool
	APIKeyPrefix              string
	APIKeyHash        string
	WebhookSecret     string
	DunningConfig     map[string]any
	Branding          map[string]any
	BillingEmail      map[string]any
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (t *Tenant) NombaBaseURL() string {
	if t.NombaEnv == NombaEnvProduction {
		return "https://api.nomba.com"
	}
	return "https://sandbox.nomba.com"
}

func (t *Tenant) NombaOrderAccountID() string {
	if t.NombaSubAccountID != "" {
		return t.NombaSubAccountID
	}
	return t.NombaAccountID
}

type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	AuthenticateAPIKey(ctx context.Context, plaintextKey string) (*Tenant, error)
}
