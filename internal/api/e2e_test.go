//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/api"
	"github.com/ireoluwacodes/subsync/internal/auth"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/crypto"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/service"
)

func TestPhase2_E2E(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	cfg := &config.Config{
		AppEnv:                        "development",
		BootstrapSecret:               "test-bootstrap-secret",
		PostgresDSN:                   dsn,
		RedisURL:                      redisURL,
		JWTSecret:                     "test-jwt-secret",
		NombaCredentialsEncryptionKey: "0123456789abcdef0123456789abcdef",
		BillingMockResult:             "success",
	}
	cfg.JWTAccessTTL = 24 * time.Hour
	cfg.JWTRefreshTTL = 168 * time.Hour

	ctx := context.Background()
	database, err := db.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(database.Close)

	q, err := queue.Connect(redisURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = q.Close() })

	key, err := crypto.ParseKey(cfg.DevEncryptionKey())
	require.NoError(t, err)
	enc, err := crypto.NewCredentialEncryptor(key)
	require.NoError(t, err)

	repos := db.NewRepos(database, enc)

	nombaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(nomba.APIResponse[nomba.TokenResponse]{
			Code: nomba.ResponseCodeSuccess,
			Data: nomba.TokenResponse{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339)},
		})
	}))
	t.Cleanup(nombaSrv.Close)

	nombaClient := nomba.NewClient(nil, nombaSrv.Client())
	nombaClient.SetTestBaseURL(nombaSrv.URL)
	jwtSvc := auth.NewJWTService(cfg)
	svcs := service.NewServices(repos, cfg, nombaClient, jwtSvc)

	router := api.NewRouter(api.RouterDeps{
		Config: cfg,
		DB:     database,
		Queue:  q,
		Repos:  repos,
		Svcs:   svcs,
		Nomba:  nombaClient,
	})

	email := "e2e-" + uuid.NewString() + "@example.com"
	createBody, _ := json.Marshal(map[string]string{
		"name":                 "E2E Tenant",
		"email":                email,
		"nomba_client_id":      "client-id",
		"nomba_client_secret":  "client-secret",
		"nomba_account_id":     "acct-merchant",
		"nomba_env":            "sandbox",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bootstrap-Secret", "test-bootstrap-secret")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var createResp struct {
		Data struct {
			APIKey string `json:"api_key"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	require.NotEmpty(t, createResp.Data.APIKey)

	meReq := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+createResp.Data.APIKey)
	meW := httptest.NewRecorder()
	router.ServeHTTP(meW, meReq)
	require.Equal(t, http.StatusOK, meW.Code)

	planBody, _ := json.Marshal(map[string]any{
		"name": "Starter", "amount": 100000, "currency": "NGN", "interval": "monthly",
	})
	planReq := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewReader(planBody))
	planReq.Header.Set("Authorization", "Bearer "+createResp.Data.APIKey)
	planReq.Header.Set("Content-Type", "application/json")
	planW := httptest.NewRecorder()
	router.ServeHTTP(planW, planReq)
	require.Equal(t, http.StatusCreated, planW.Code)
}
