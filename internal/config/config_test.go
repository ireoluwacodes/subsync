package config_test

import (
	"testing"

	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/stretchr/testify/require"
)

func TestConfig_ProductionRejectsBillingMock(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("POSTGRES_DSN", "postgres://u:p@localhost:5432/subsync?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("NOMBA_CREDENTIALS_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("PUBLIC_BASE_URL", "https://api.example.com")
	t.Setenv("BILLING_MOCK_RESULT", "success")

	_, err := config.Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "BILLING_MOCK_RESULT")
}
