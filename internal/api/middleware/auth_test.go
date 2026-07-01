package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type mockTenantRepo struct {
	tenant *domain.Tenant
	err    error
}

func (m *mockTenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error { return nil }
func (m *mockTenantRepo) Update(ctx context.Context, tenant *domain.Tenant) error { return nil }
func (m *mockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if m.tenant != nil && m.tenant.ID == id {
		return m.tenant, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockTenantRepo) AuthenticateAPIKey(ctx context.Context, plaintextKey string) (*domain.Tenant, error) {
	return m.tenant, m.err
}

func TestAuth_ValidAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenant := &domain.Tenant{ID: uuid.New(), Name: "Acme"}
	repo := &mockTenantRepo{tenant: tenant}

	r := gin.New()
	r.Use(middleware.Auth(repo, nil, nil))
	r.GET("/test", func(c *gin.Context) {
		got, ok := middleware.TenantFromContext(c)
		require.True(t, ok)
		require.Equal(t, tenant.ID, got.ID)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ssk_deadbeef_secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_MissingAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Auth(&mockTenantRepo{}, nil, nil))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockTenantRepo{err: domain.ErrNotFound}

	r := gin.New()
	r.Use(middleware.Auth(repo, nil, nil))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer bad-key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
