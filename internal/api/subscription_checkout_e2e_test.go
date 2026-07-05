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

func TestSubscriptionCheckout_NoTrial_E2E(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	webhookSecret := "checkout-webhook-secret"
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
	require.NoError(t, db.Migrate(ctx, database))

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
		case "/v1/checkout/order":
			var body nomba.CreateOrderRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			capturedOrderRef = body.Order.OrderReference
			_ = json.NewEncoder(w).Encode(nomba.APIResponse[nomba.CreateOrderResult]{
				Code: nomba.ResponseCodeSuccess,
				Data: nomba.CreateOrderResult{
					CheckoutLink:   "https://checkout.nomba.test/pay",
					OrderReference: body.Order.OrderReference,
				},
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

	apiKey, tenantID := registerTenant(t, engine)

	planBody, _ := json.Marshal(map[string]any{
		"name": "Pro", "amount": 100000, "currency": "NGN", "interval": "monthly", "trial_days": 0,
	})
	planID := postJSON(t, engine, "/api/v1/plans", apiKey, planBody, http.StatusCreated)

	custBody, _ := json.Marshal(map[string]any{
		"email": "checkout-" + uuid.NewString() + "@example.com",
		"name":  "Jane Doe",
	})
	customerID := postJSON(t, engine, "/api/v1/customers", apiKey, custBody, http.StatusCreated)

	checkoutBody, _ := json.Marshal(map[string]any{
		"customer_id":  customerID,
		"plan_id":      planID,
		"success_url":  "http://localhost:3000/billing/success",
		"cancel_url":   "http://localhost:3000/pricing",
	})
	checkoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/checkout", bytes.NewReader(checkoutBody))
	checkoutReq.Header.Set("Authorization", "Bearer "+apiKey)
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutW := httptest.NewRecorder()
	engine.ServeHTTP(checkoutW, checkoutReq)
	require.Equal(t, http.StatusCreated, checkoutW.Code, checkoutW.Body.String())

	var checkoutResp struct {
		Data struct {
			SubscriptionID string `json:"subscription_id"`
			CheckoutURL    string `json:"checkout_url"`
			Status         string `json:"status"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(checkoutW.Body.Bytes(), &checkoutResp))
	require.Equal(t, "incomplete", checkoutResp.Data.Status)
	require.NotEmpty(t, checkoutResp.Data.CheckoutURL)
	require.NotEmpty(t, capturedOrderRef)

	subID, err := uuid.Parse(checkoutResp.Data.SubscriptionID)
	require.NoError(t, err)

	subBefore, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.Equal(t, domain.SubscriptionStateIncomplete, subBefore.State)
	periodEndBefore := subBefore.CurrentPeriodEnd

	inv, err := repos.Invoices.GetByNombaOrderRef(ctx, tenantID, capturedOrderRef)
	require.NoError(t, err)
	require.Equal(t, domain.InvoiceStatusOpen, inv.Status)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	webhookBody := []byte(`{"event_type":"payment_success","requestId":"` + uuid.NewString() + `","data":{"merchant":{"userId":"u","walletId":"w"},"transaction":{"transactionId":"tx-checkout","type":"purchase","time":"2026-01-01T00:00:00Z","merchantTxRef":"` + capturedOrderRef + `","tokenKey":"tok_checkout_card"}}}`)
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

	subFinal, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.Equal(t, domain.SubscriptionStateActive, subFinal.State)
	require.NotNil(t, subFinal.PaymentMethodID)
	require.NotNil(t, subFinal.NextBillingAt)
	require.Equal(t, periodEndBefore, subFinal.CurrentPeriodEnd)
	require.Equal(t, periodEndBefore, *subFinal.NextBillingAt)
}

func TestSubscriptionCheckout_Transfer_E2E(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	webhookSecret := "checkout-transfer-secret"
	cfg := &config.Config{
		AppEnv:                        "development",
		PostgresDSN:                   dsn,
		RedisURL:                      redisURL,
		JWTSecret:                     "test-jwt-secret",
		NombaCredentialsEncryptionKey: "0123456789abcdef0123456789abcdef",
		NombaWebhookSigningKey:        webhookSecret,
		PublicBaseURL:                 "http://localhost:8080",
	}
	cfg.JWTAccessTTL = 24 * time.Hour
	cfg.JWTRefreshTTL = 168 * time.Hour

	ctx := context.Background()
	database, err := db.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(database.Close)
	require.NoError(t, db.Migrate(ctx, database))

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
		case "/v1/checkout/order":
			var body nomba.CreateOrderRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			capturedOrderRef = body.Order.OrderReference
			_ = json.NewEncoder(w).Encode(nomba.APIResponse[nomba.CreateOrderResult]{
				Code: nomba.ResponseCodeSuccess,
				Data: nomba.CreateOrderResult{CheckoutLink: "https://checkout.nomba.test/transfer"},
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

	apiKey, tenantID := registerTenant(t, engine)

	planBody, _ := json.Marshal(map[string]any{
		"name": "Transfer Plan", "amount": 100000, "currency": "NGN", "interval": "monthly", "trial_days": 0,
	})
	planID := postJSON(t, engine, "/api/v1/plans", apiKey, planBody, http.StatusCreated)

	custBody, _ := json.Marshal(map[string]any{
		"email": "transfer-" + uuid.NewString() + "@example.com",
		"name":  "Transfer User",
	})
	customerID := postJSON(t, engine, "/api/v1/customers", apiKey, custBody, http.StatusCreated)

	checkoutBody, _ := json.Marshal(map[string]any{
		"customer_id":         customerID,
		"plan_id":             planID,
		"success_url":         "http://localhost:3000/billing/success",
		"allow_bank_transfer": true,
	})
	checkoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/checkout", bytes.NewReader(checkoutBody))
	checkoutReq.Header.Set("Authorization", "Bearer "+apiKey)
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutW := httptest.NewRecorder()
	engine.ServeHTTP(checkoutW, checkoutReq)
	require.Equal(t, http.StatusCreated, checkoutW.Code)

	var checkoutResp struct {
		Data struct {
			SubscriptionID string `json:"subscription_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(checkoutW.Body.Bytes(), &checkoutResp))
	subID, err := uuid.Parse(checkoutResp.Data.SubscriptionID)
	require.NoError(t, err)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	webhookBody := []byte(`{"event_type":"payment_success","requestId":"` + uuid.NewString() + `","data":{"merchant":{"userId":"u"},"transaction":{"transactionId":"tx-transfer","type":"vact_transfer","time":"2026-01-01T00:00:00Z","merchantTxRef":"` + capturedOrderRef + `","aliasAccountReference":"` + capturedOrderRef + `"}}}`)
	sig, err := nomba.GenerateWebhookSignature(webhookBody, webhookSecret, timestamp)
	require.NoError(t, err)

	hookReq := httptest.NewRequest(http.MethodPost, "/webhooks/nomba/"+tenantID.String(), bytes.NewReader(webhookBody))
	hookReq.Header.Set("Content-Type", "application/json")
	hookReq.Header.Set("nomba-signature", sig)
	hookReq.Header.Set("nomba-timestamp", timestamp)
	hookW := httptest.NewRecorder()
	engine.ServeHTTP(hookW, hookReq)
	require.Equal(t, http.StatusOK, hookW.Code, hookW.Body.String())

	subFinal, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.Equal(t, domain.SubscriptionStateActive, subFinal.State)
	require.Nil(t, subFinal.PaymentMethodID)
	require.True(t, subFinal.Metadata[domain.SubscriptionMetaAwaitingPaymentMethod].(bool))

	captureBody, _ := json.Marshal(map[string]any{
		"success_url": "http://localhost:3000/billing/card-added",
	})
	captureReq := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/"+subID.String()+"/capture-payment-method", bytes.NewReader(captureBody))
	captureReq.Header.Set("Authorization", "Bearer "+apiKey)
	captureReq.Header.Set("Content-Type", "application/json")
	captureW := httptest.NewRecorder()
	engine.ServeHTTP(captureW, captureReq)
	require.Equal(t, http.StatusOK, captureW.Code, captureW.Body.String())

	var captureResp struct {
		Data struct {
			OrderReference string `json:"order_reference"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(captureW.Body.Bytes(), &captureResp))

	captureWebhookBody := []byte(`{"event_type":"payment_success","requestId":"` + uuid.NewString() + `","data":{"merchant":{"userId":"u"},"transaction":{"transactionId":"tx-capture","type":"purchase","time":"2026-01-01T00:00:00Z","merchantTxRef":"` + captureResp.Data.OrderReference + `","tokenKey":"tok_capture_card"}}}`)
	captureSig, err := nomba.GenerateWebhookSignature(captureWebhookBody, webhookSecret, timestamp)
	require.NoError(t, err)

	captureHookReq := httptest.NewRequest(http.MethodPost, "/webhooks/nomba/"+tenantID.String(), bytes.NewReader(captureWebhookBody))
	captureHookReq.Header.Set("Content-Type", "application/json")
	captureHookReq.Header.Set("nomba-signature", captureSig)
	captureHookReq.Header.Set("nomba-timestamp", timestamp)
	captureHookW := httptest.NewRecorder()
	engine.ServeHTTP(captureHookW, captureHookReq)
	require.Equal(t, http.StatusOK, captureHookW.Code, captureHookW.Body.String())

	subWithCard, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.NotNil(t, subWithCard.PaymentMethodID)
	_, hasFlag := subWithCard.Metadata[domain.SubscriptionMetaAwaitingPaymentMethod]
	require.False(t, hasFlag)
}

func TestSubscriptionCheckout_TransferCancelAtRenewal_E2E(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	webhookSecret := "checkout-transfer-cancel-secret"
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
	require.NoError(t, db.Migrate(ctx, database))

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
		case "/v1/checkout/order":
			var body nomba.CreateOrderRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			capturedOrderRef = body.Order.OrderReference
			_ = json.NewEncoder(w).Encode(nomba.APIResponse[nomba.CreateOrderResult]{
				Code: nomba.ResponseCodeSuccess,
				Data: nomba.CreateOrderResult{CheckoutLink: "https://checkout.nomba.test/transfer"},
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

	apiKey, tenantID := registerTenant(t, engine)

	planBody, _ := json.Marshal(map[string]any{
		"name": "Cancel Plan", "amount": 100000, "currency": "NGN", "interval": "monthly", "trial_days": 0,
	})
	planID := postJSON(t, engine, "/api/v1/plans", apiKey, planBody, http.StatusCreated)

	custBody, _ := json.Marshal(map[string]any{
		"email": "transfer-cancel-" + uuid.NewString() + "@example.com",
		"name":  "Cancel User",
	})
	customerID := postJSON(t, engine, "/api/v1/customers", apiKey, custBody, http.StatusCreated)

	checkoutBody, _ := json.Marshal(map[string]any{
		"customer_id":         customerID,
		"plan_id":             planID,
		"success_url":         "http://localhost:3000/billing/success",
		"allow_bank_transfer": true,
	})
	checkoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/checkout", bytes.NewReader(checkoutBody))
	checkoutReq.Header.Set("Authorization", "Bearer "+apiKey)
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutW := httptest.NewRecorder()
	engine.ServeHTTP(checkoutW, checkoutReq)
	require.Equal(t, http.StatusCreated, checkoutW.Code)

	var checkoutResp struct {
		Data struct {
			SubscriptionID string `json:"subscription_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(checkoutW.Body.Bytes(), &checkoutResp))
	subID, err := uuid.Parse(checkoutResp.Data.SubscriptionID)
	require.NoError(t, err)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	webhookBody := []byte(`{"event_type":"payment_success","requestId":"` + uuid.NewString() + `","data":{"merchant":{"userId":"u"},"transaction":{"transactionId":"tx-transfer","type":"vact_transfer","time":"2026-01-01T00:00:00Z","merchantTxRef":"` + capturedOrderRef + `","aliasAccountReference":"` + capturedOrderRef + `"}}}`)
	sig, err := nomba.GenerateWebhookSignature(webhookBody, webhookSecret, timestamp)
	require.NoError(t, err)

	hookReq := httptest.NewRequest(http.MethodPost, "/webhooks/nomba/"+tenantID.String(), bytes.NewReader(webhookBody))
	hookReq.Header.Set("Content-Type", "application/json")
	hookReq.Header.Set("nomba-signature", sig)
	hookReq.Header.Set("nomba-timestamp", timestamp)
	hookW := httptest.NewRecorder()
	engine.ServeHTTP(hookW, hookReq)
	require.Equal(t, http.StatusOK, hookW.Code)

	sub, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.Nil(t, sub.PaymentMethodID)

	past := time.Now().UTC().Add(-time.Hour)
	sub.NextBillingAt = &past
	sub.CurrentPeriodEnd = past
	require.NoError(t, repos.Subscriptions.Update(ctx, sub))

	_, err = svcs.Billing.ProcessDueSubscriptions(ctx, 10)
	require.NoError(t, err)

	subAfter, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.Equal(t, domain.SubscriptionStateCanceled, subAfter.State)
}

func TestSubscriptionCheckout_Trial_E2E(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	webhookSecret := "checkout-trial-secret"
	cfg := &config.Config{
		AppEnv:                        "development",
		PostgresDSN:                   dsn,
		RedisURL:                      redisURL,
		JWTSecret:                     "test-jwt-secret",
		NombaCredentialsEncryptionKey: "0123456789abcdef0123456789abcdef",
		NombaWebhookSigningKey:        webhookSecret,
		PublicBaseURL:                 "http://localhost:8080",
	}
	cfg.JWTAccessTTL = 24 * time.Hour
	cfg.JWTRefreshTTL = 168 * time.Hour

	ctx := context.Background()
	database, err := db.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(database.Close)
	require.NoError(t, db.Migrate(ctx, database))

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
		case "/v1/checkout/order":
			var body nomba.CreateOrderRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			capturedOrderRef = body.Order.OrderReference
			_ = json.NewEncoder(w).Encode(nomba.APIResponse[nomba.CreateOrderResult]{
				Code: nomba.ResponseCodeSuccess,
				Data: nomba.CreateOrderResult{CheckoutLink: "https://checkout.nomba.test/trial"},
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

	apiKey, tenantID := registerTenant(t, engine)

	planBody, _ := json.Marshal(map[string]any{
		"name": "Trial Pro", "amount": 100000, "currency": "NGN", "interval": "monthly", "trial_days": 7,
	})
	planID := postJSON(t, engine, "/api/v1/plans", apiKey, planBody, http.StatusCreated)

	custBody, _ := json.Marshal(map[string]any{
		"email": "trial-" + uuid.NewString() + "@example.com",
		"name":  "Trial User",
	})
	customerID := postJSON(t, engine, "/api/v1/customers", apiKey, custBody, http.StatusCreated)

	checkoutBody, _ := json.Marshal(map[string]any{
		"customer_id": customerID,
		"plan_id":     planID,
		"success_url": "http://localhost:3000/billing/success",
	})
	checkoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/checkout", bytes.NewReader(checkoutBody))
	checkoutReq.Header.Set("Authorization", "Bearer "+apiKey)
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutW := httptest.NewRecorder()
	engine.ServeHTTP(checkoutW, checkoutReq)
	require.Equal(t, http.StatusCreated, checkoutW.Code)

	var checkoutResp struct {
		Data struct {
			SubscriptionID string `json:"subscription_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(checkoutW.Body.Bytes(), &checkoutResp))
	subID, err := uuid.Parse(checkoutResp.Data.SubscriptionID)
	require.NoError(t, err)
	require.True(t, len(capturedOrderRef) > 0)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	webhookBody := []byte(`{"event_type":"payment_success","requestId":"` + uuid.NewString() + `","data":{"merchant":{"userId":"u"},"transaction":{"transactionId":"tx-trial","type":"purchase","time":"2026-01-01T00:00:00Z","merchantTxRef":"` + capturedOrderRef + `","tokenKey":"tok_trial_card"}}}`)
	sig, err := nomba.GenerateWebhookSignature(webhookBody, webhookSecret, timestamp)
	require.NoError(t, err)

	hookReq := httptest.NewRequest(http.MethodPost, "/webhooks/nomba/"+tenantID.String(), bytes.NewReader(webhookBody))
	hookReq.Header.Set("Content-Type", "application/json")
	hookReq.Header.Set("nomba-signature", sig)
	hookReq.Header.Set("nomba-timestamp", timestamp)
	hookW := httptest.NewRecorder()
	engine.ServeHTTP(hookW, hookReq)
	require.Equal(t, http.StatusOK, hookW.Code, hookW.Body.String())

	subFinal, err := repos.Subscriptions.GetByID(ctx, tenantID, subID)
	require.NoError(t, err)
	require.Equal(t, domain.SubscriptionStateTrialing, subFinal.State)
	require.NotNil(t, subFinal.PaymentMethodID)
	require.NotNil(t, subFinal.TrialEndsAt)
	require.NotNil(t, subFinal.NextBillingAt)
}

func registerTenant(t *testing.T, engine http.Handler) (apiKey string, tenantID uuid.UUID) {
	t.Helper()
	email := "checkout-tenant-" + uuid.NewString() + "@example.com"
	registerBody, _ := json.Marshal(map[string]string{
		"email":               email,
		"password":            "securepass123",
		"name":                "Checkout Tenant",
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
	id, err := uuid.Parse(registerResp.Data.Tenant.ID)
	require.NoError(t, err)
	return registerResp.Data.APIKey, id
}

func postJSON(t *testing.T, engine http.Handler, path, apiKey string, body []byte, wantStatus int) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, wantStatus, w.Code, w.Body.String())

	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp.Data.ID
}
