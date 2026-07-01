package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

func TestJWTIssueAndParse(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", JWTAccessTTL: time.Hour, JWTRefreshTTL: 24 * time.Hour}
	jwt := NewJWTService(cfg)
	user := &domain.User{
		ID: uuid.New(), TenantID: uuid.New(), Email: "a@b.com", TokenVersion: 1,
	}
	pair, err := jwt.IssuePair(user, user.TenantID)
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)

	claims, err := jwt.Parse(pair.AccessToken)
	require.NoError(t, err)
	require.Equal(t, "access", claims.Type)
	require.Equal(t, user.ID.String(), claims.Subject)
}
