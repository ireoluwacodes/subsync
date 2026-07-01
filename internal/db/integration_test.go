//go:build integration

package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/crypto"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func testDB(t *testing.T) *db.DB {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}
	database, err := db.Connect(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(database.Close)
	return database
}

func testEncryptor(t *testing.T) *crypto.CredentialEncryptor {
	t.Helper()
	key, err := crypto.ParseKey("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	enc, err := crypto.NewCredentialEncryptor(key)
	require.NoError(t, err)
	return enc
}

func TestTenantRepo_Integration(t *testing.T) {
	database := testDB(t)
	repo := db.NewTenantRepo(database, testEncryptor(t))
	ctx := context.Background()

	prefix := uuid.NewString()[:8]
	fullKey := "ssk_" + prefix + "_integration-secret"
	hash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	require.NoError(t, err)

	tenant := &domain.Tenant{
		Name:              "Test Tenant " + uuid.NewString()[:8],
		Email:             "tenant-" + uuid.NewString()[:8] + "@example.com",
		NombaClientID:     "client-id",
		NombaClientSecret: "client-secret",
		NombaAccountID:    "acct-merchant",
		NombaEnv:          domain.NombaEnvSandbox,
		APIKeyPrefix:      prefix,
		APIKeyHash:        string(hash),
		WebhookSecret:     "whsec_test",
		DunningConfig:     map[string]any{"steps": []any{}},
	}
	require.NoError(t, repo.Create(ctx, tenant))

	got, err := repo.GetByID(ctx, tenant.ID)
	require.NoError(t, err)
	require.Equal(t, tenant.Email, got.Email)
	require.Empty(t, got.NombaClientSecret)

	loaded := *got
	require.NoError(t, repo.LoadNombaSecret(ctx, &loaded))
	require.Equal(t, "client-secret", loaded.NombaClientSecret)

	auth, err := repo.AuthenticateAPIKey(ctx, fullKey)
	require.NoError(t, err)
	require.Equal(t, tenant.ID, auth.ID)

	_, err = repo.AuthenticateAPIKey(ctx, "ssk_"+prefix+"_wrong-secret")
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTenantRepo_NombaWebhookSecret_Integration(t *testing.T) {
	database := testDB(t)
	repo := db.NewTenantRepo(database, testEncryptor(t))
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("ssk_test1234_key"), bcrypt.DefaultCost)
	require.NoError(t, err)

	tenant := &domain.Tenant{
		Name:              "Webhook Tenant " + uuid.NewString()[:8],
		Email:             "wh-" + uuid.NewString()[:8] + "@example.com",
		NombaClientID:     "client-id",
		NombaClientSecret: "client-secret",
		NombaAccountID:    "acct-merchant",
		NombaEnv:          domain.NombaEnvSandbox,
		APIKeyPrefix:      uuid.NewString()[:8],
		APIKeyHash:        string(hash),
		WebhookSecret:     "whsec_test",
		DunningConfig:     map[string]any{},
	}
	require.NoError(t, repo.Create(ctx, tenant))
	require.False(t, tenant.HasNombaWebhookSecret)

	tenant.NombaWebhookSecret = "nomba-whsec-from-dashboard"
	require.NoError(t, repo.Update(ctx, tenant))
	require.True(t, tenant.HasNombaWebhookSecret)

	got, err := repo.GetByID(ctx, tenant.ID)
	require.NoError(t, err)
	require.True(t, got.HasNombaWebhookSecret)

	loaded := *got
	require.NoError(t, repo.LoadNombaWebhookSecret(ctx, &loaded))
	require.Equal(t, "nomba-whsec-from-dashboard", loaded.NombaWebhookSecret)
}

func TestPlanRepo_Integration(t *testing.T) {
	database := testDB(t)
	tenantRepo := db.NewTenantRepo(database, testEncryptor(t))
	planRepo := db.NewPlanRepo(database)
	ctx := context.Background()

	tenant := createTestTenant(t, tenantRepo, ctx)

	plan := &domain.Plan{
		TenantID: tenant.ID,
		Name:     "Pro",
		Amount:   500000,
		Currency: "NGN",
		Interval: domain.PlanIntervalMonthly,
		IsActive: true,
	}
	require.NoError(t, planRepo.Create(ctx, plan))

	got, err := planRepo.GetByID(ctx, tenant.ID, plan.ID)
	require.NoError(t, err)
	require.Equal(t, "Pro", got.Name)

	plans, err := planRepo.List(ctx, tenant.ID, true, 10, 0)
	require.NoError(t, err)
	require.Len(t, plans, 1)

	count, err := planRepo.Count(ctx, tenant.ID, true)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestCustomerRepo_Integration(t *testing.T) {
	database := testDB(t)
	tenantRepo := db.NewTenantRepo(database, testEncryptor(t))
	customerRepo := db.NewCustomerRepo(database)
	ctx := context.Background()

	tenant := createTestTenant(t, tenantRepo, ctx)

	customer := &domain.Customer{
		TenantID: tenant.ID,
		Email:    "user-" + uuid.NewString()[:8] + "@example.com",
		Name:     "Jane",
	}
	require.NoError(t, customerRepo.Create(ctx, customer))

	dup := &domain.Customer{TenantID: tenant.ID, Email: customer.Email}
	require.ErrorIs(t, customerRepo.Create(ctx, dup), domain.ErrConflict)
}

func createTestTenant(t *testing.T, repo *db.TenantRepo, ctx context.Context) *domain.Tenant {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("ssk_test1234_key"), bcrypt.DefaultCost)
	require.NoError(t, err)
	tenant := &domain.Tenant{
		Name:              "Tenant " + uuid.NewString()[:8],
		Email:             "t-" + uuid.NewString()[:8] + "@example.com",
		NombaClientID:     "client",
		NombaClientSecret: "secret",
		NombaAccountID:    "acct",
		NombaEnv:          domain.NombaEnvSandbox,
		APIKeyPrefix:      "test1234",
		APIKeyHash:        string(hash),
		WebhookSecret:     "whsec",
		DunningConfig:     map[string]any{},
	}
	require.NoError(t, repo.Create(ctx, tenant))
	return tenant
}
