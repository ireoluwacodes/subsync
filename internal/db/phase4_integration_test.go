//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func TestNombaEventRepo_Integration(t *testing.T) {
	database := testDB(t)
	repos := db.NewRepos(database, testEncryptor(t))
	ctx := context.Background()

	tenant := createTestTenant(t, repos.Tenants, ctx)
	event := &domain.NombaEvent{
		TenantID:  tenant.ID,
		EventID:   "req-" + uuid.NewString(),
		EventType: "payment_success",
		Payload:   map[string]any{"event_type": "payment_success"},
	}
	require.NoError(t, repos.NombaEvents.Create(ctx, event))

	got, err := repos.NombaEvents.GetByEventID(ctx, tenant.ID, event.EventID)
	require.NoError(t, err)
	require.Equal(t, event.EventID, got.EventID)

	require.NoError(t, repos.NombaEvents.MarkProcessed(ctx, event.ID))
}

func TestWebhookRepo_Integration(t *testing.T) {
	database := testDB(t)
	repos := db.NewRepos(database, testEncryptor(t))
	ctx := context.Background()

	tenant := createTestTenant(t, repos.Tenants, ctx)
	ep := &domain.WebhookEndpoint{
		TenantID: tenant.ID,
		URL:      "https://example.com/hooks",
		Events:   []string{"invoice.paid"},
		IsActive: true,
	}
	require.NoError(t, repos.Webhooks.CreateEndpoint(ctx, ep))

	list, err := repos.Webhooks.ListEndpoints(ctx, tenant.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	delivery := &domain.WebhookDelivery{
		TenantID:   tenant.ID,
		EndpointID: ep.ID,
		EventType:  "invoice.paid",
		Payload:    map[string]any{"id": "inv-1"},
	}
	require.NoError(t, repos.Webhooks.CreateDelivery(ctx, delivery))

	deliveries, err := repos.Webhooks.ListDeliveries(ctx, tenant.ID, ep.ID)
	require.NoError(t, err)
	require.Len(t, deliveries, 1)
}

func TestPortalTokenRepo_Integration(t *testing.T) {
	database := testDB(t)
	repos := db.NewRepos(database, testEncryptor(t))
	ctx := context.Background()

	tenant := createTestTenant(t, repos.Tenants, ctx)
	customer := createTestCustomer(t, repos.Customers, ctx, tenant.ID)
	plan := createTestPlan(t, repos.Plans, ctx, tenant.ID)
	sub := createTestSubscription(t, repos.Subscriptions, ctx, tenant.ID, customer.ID, plan.ID)

	raw := "portal-token-" + uuid.NewString()
	token := &domain.PortalToken{
		TenantID:       tenant.ID,
		SubscriptionID: sub.ID,
		CustomerID:     customer.ID,
		TokenHash:      utils.HashResetSecret(raw),
		ExpiresAt:      time.Now().UTC().Add(24 * time.Hour),
	}
	require.NoError(t, repos.PortalTokens.Create(ctx, token))

	got, err := repos.PortalTokens.GetValidByTokenHash(ctx, utils.HashResetSecret(raw))
	require.NoError(t, err)
	require.Equal(t, token.ID, got.ID)
}

func createTestTenant(t *testing.T, repo *db.TenantRepo, ctx context.Context) *domain.Tenant {
	t.Helper()
	prefix := uuid.NewString()[:8]
	hash, err := bcrypt.GenerateFromPassword([]byte("test-key-"+prefix), bcrypt.DefaultCost)
	require.NoError(t, err)
	tenant := &domain.Tenant{
		Name:              "Phase4 " + prefix,
		Email:             "p4-" + prefix + "@example.com",
		NombaClientID:     "client",
		NombaClientSecret: "secret",
		NombaAccountID:    "acct",
		NombaEnv:          domain.NombaEnvSandbox,
		APIKeyPrefix:      prefix,
		APIKeyHash:        string(hash),
		WebhookSecret:     "whsec_test",
		DunningConfig:     map[string]any{"steps": []any{}},
	}
	require.NoError(t, repo.Create(ctx, tenant))
	return tenant
}

func createTestCustomer(t *testing.T, repo *db.CustomerRepo, ctx context.Context, tenantID uuid.UUID) *domain.Customer {
	t.Helper()
	c := &domain.Customer{TenantID: tenantID, Email: "cust-" + uuid.NewString()[:8] + "@example.com"}
	require.NoError(t, repo.Create(ctx, c))
	return c
}

func createTestPlan(t *testing.T, repo *db.PlanRepo, ctx context.Context, tenantID uuid.UUID) *domain.Plan {
	t.Helper()
	p := &domain.Plan{
		TenantID: tenantID,
		Name:     "Basic",
		Amount:   1000,
		Currency: "NGN",
		Interval: "month",
	}
	require.NoError(t, repo.Create(ctx, p))
	return p
}

func createTestSubscription(t *testing.T, repo *db.SubscriptionRepo, ctx context.Context, tenantID, customerID, planID uuid.UUID) *domain.Subscription {
	t.Helper()
	now := time.Now().UTC()
	sub := &domain.Subscription{
		TenantID:           tenantID,
		CustomerID:         customerID,
		PlanID:             planID,
		State:              domain.SubscriptionStateActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		Metadata:           map[string]any{},
	}
	require.NoError(t, repo.Create(ctx, sub))
	return sub
}
