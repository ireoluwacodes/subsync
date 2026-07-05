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
	"github.com/ireoluwacodes/subsync/internal/auth"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/crypto"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/router"
	"github.com/ireoluwacodes/subsync/internal/service"
)

func TestPhase5_LiveBillingWebhook(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	webhookSecret := "phase5-webhook-secret"
	cfg := &config.Config{
		AppEnv:                        "development",
		PostgresDSN:                   dsn,
		RedisURL:                      redisURL,
		JWTSecret:                     "test-jwt-secret",
		NombaCredentialsEncryptionKey: "0123456789abcdef0123456789abcdef",
		NombaWebhookSigningKey:        webhookSecret,
		PublicBaseURL:                 "http://localhost:8080",
		BillingMockResult:             "",
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

	var capturedOrderRef string
	nombaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/token/issue":
			_ = json.NewEncoder(w).Encode(nomba.APIResponse[nomba.TokenResponse]{
				Code: nomba.ResponseCodeSuccess,
				Data: nomba.TokenResponse{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339)},
			})
		case "/v1/checkout/tokenized-card-payment":
			var body nomba.TokenizedCardPaymentRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			capturedOrderRef = body.Order.OrderReference
			_ = json.NewEncoder(w).Encode(nomba.APIResponse[nomba.TokenizedCardPaymentResult]{
				Code: nomba.ResponseCodeSuccess,
				Data: nomba.TokenizedCardPaymentResult{Status: true, Message: "pending-tx"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(nombaSrv.Close)

	nombaClient := nomba.NewClient(nil, nombaSrv.Client())
	nombaClient.SetTestBaseURL(nombaSrv.URL)
	jwtSvc := auth.NewJWTService(cfg)
	svcs := service.NewServices(repos, cfg, nombaClient, jwtSvc, q)
	engine := router.Setup(cfg, database, q, repos, svcs)

	email := "phase5-" + uuid.NewString() + "@example.com"
	registerBody, _ := json.Marshal(map[string]string{
		"email":               email,
		"password":            "securepass123",
		"name":                "Phase5 Tenant",
		"nomba_client_id":     "client-id",
		"nomba_client_secret": "client-secret",
		"nomba_account_id":    "acct-merchant",
		"nomba_env":           "sandbox",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var registerResp struct {
		Data struct {
			APIKey string `json:"api_key"`
			Tenant struct {
				ID string `json:"id"`
			} `json:"tenant"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &registerResp))
	apiKey := registerResp.Data.APIKey
	tenantID, err := uuid.Parse(registerResp.Data.Tenant.ID)
	require.NoError(t, err)

	planBody, _ := json.Marshal(map[string]any{
		"name": "Starter", "amount": 100000, "currency": "NGN", "interval": "monthly",
	})
	planReq := httptest.NewRequest(http.MethodPost, "/api/v1/plans", bytes.NewReader(planBody))
	planReq.Header.Set("Authorization", "Bearer "+apiKey)
	planReq.Header.Set("Content-Type", "application/json")
	planW := httptest.NewRecorder()
	engine.ServeHTTP(planW, planReq)
	require.Equal(t, http.StatusCreated, planW.Code)

	var planResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(planW.Body.Bytes(), &planResp))

	custBody, _ := json.Marshal(map[string]any{
		"email": "cust-" + uuid.NewString() + "@example.com",
		"name":  "Jane Doe",
	})
	custReq := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(custBody))
	custReq.Header.Set("Authorization", "Bearer "+apiKey)
	custReq.Header.Set("Content-Type", "application/json")
	custW := httptest.NewRecorder()
	engine.ServeHTTP(custW, custReq)
	require.Equal(t, http.StatusCreated, custW.Code)

	var custResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(custW.Body.Bytes(), &custResp))

	pmBody, _ := json.Marshal(map[string]any{
		"customer_id": custResp.Data.ID,
		"type":        "tokenized_card",
		"token_key":   "tok_test",
		"is_default":  true,
	})
	pmReq := httptest.NewRequest(http.MethodPost, "/api/v1/payment-methods", bytes.NewReader(pmBody))
	pmReq.Header.Set("Authorization", "Bearer "+apiKey)
	pmReq.Header.Set("Content-Type", "application/json")
	pmW := httptest.NewRecorder()
	engine.ServeHTTP(pmW, pmReq)
	require.Equal(t, http.StatusCreated, pmW.Code)

	var pmResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(pmW.Body.Bytes(), &pmResp))

	now := time.Now().UTC()
	subBody, _ := json.Marshal(map[string]any{
		"customer_id":       custResp.Data.ID,
		"plan_id":           planResp.Data.ID,
		"payment_method_id": pmResp.Data.ID,
	})
	subReq := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewReader(subBody))
	subReq.Header.Set("Authorization", "Bearer "+apiKey)
	subReq.Header.Set("Content-Type", "application/json")
	subW := httptest.NewRecorder()
	engine.ServeHTTP(subW, subReq)
	require.Equal(t, http.StatusCreated, subW.Code)

	var subResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(subW.Body.Bytes(), &subResp))
	subID, err := uuid.Parse(subResp.Data.ID)
	require.NoError(t, err)

	sub, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	sub.NextBillingAt = &now
	sub.CurrentPeriodEnd = now
	require.NoError(t, repos.Subscriptions.Update(ctx, sub))

	require.NoError(t, svcs.Billing.ChargeDueSubscription(ctx, tenantID, subID))
	require.NotEmpty(t, capturedOrderRef)

	inv, err := repos.Invoices.GetByNombaOrderRef(ctx, tenantID, capturedOrderRef)
	require.NoError(t, err)
	require.Equal(t, domain.InvoiceStatusProcessing, inv.Status)

	periodEndBefore := sub.CurrentPeriodEnd
	subAfterCharge, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.True(t, periodEndBefore.Equal(subAfterCharge.CurrentPeriodEnd))

	timestamp := time.Now().UTC().Format(time.RFC3339)
	webhookBody := []byte(`{"event_type":"payment_success","requestId":"` + uuid.NewString() + `","data":{"merchant":{"userId":"u","walletId":"w"},"transaction":{"transactionId":"tx-final","type":"purchase","time":"2026-01-01T00:00:00Z","merchantTxRef":"` + capturedOrderRef + `"}}}`)
	sig, err := nomba.GenerateWebhookSignature(webhookBody, webhookSecret, timestamp)
	require.NoError(t, err)

	hookReq := httptest.NewRequest(http.MethodPost, "/webhooks/nomba/"+tenantID.String(), bytes.NewReader(webhookBody))
	hookReq.Header.Set("Content-Type", "application/json")
	hookReq.Header.Set("nomba-signature", sig)
	hookReq.Header.Set("nomba-timestamp", timestamp)
	hookW := httptest.NewRecorder()
	engine.ServeHTTP(hookW, hookReq)
	require.Equal(t, http.StatusOK, hookW.Code, hookW.Body.String())

	invPaid, err := repos.Invoices.GetByID(ctx, tenantID, inv.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InvoiceStatusPaid, invPaid.Status)
	require.Equal(t, "tx-final", invPaid.NombaTransactionID)

	subFinal, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.True(t, subFinal.CurrentPeriodEnd.After(periodEndBefore))
	require.Equal(t, domain.SubscriptionStateActive, subFinal.State)
}
