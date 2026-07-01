package nomba

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"go.uber.org/zap"
)

func testTenant() *domain.Tenant {
	return &domain.Tenant{
		ID:                uuid.New(),
		NombaClientID:     "client-id",
		NombaClientSecret: "client-secret",
		NombaAccountID:    "acct-merchant",
		NombaEnv:          domain.NombaEnvSandbox,
	}
}

func testClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c := NewClient(zap.NewNop(), srv.Client())
	c.SetTestBaseURL(srv.URL)
	return c
}

func TestClient_CreateOrder(t *testing.T) {
	var authCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PathAuthTokenIssue:
			authCalls.Add(1)
			require.Equal(t, "acct-merchant", r.Header.Get(HeaderAccountID))
			_ = json.NewEncoder(w).Encode(APIResponse[TokenResponse]{
				Code: ResponseCodeSuccess,
				Data: TokenResponse{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresAt:    time.Now().Add(30 * time.Minute).Format(time.RFC3339),
				},
			})
		case PathCheckoutOrder:
			require.Equal(t, "Bearer access-token", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode(APIResponse[CreateOrderResult]{
				Code: ResponseCodeSuccess,
				Data: CreateOrderResult{OrderReference: "ord-123"},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := testClient(t, srv)
	result, err := client.CreateOrder(context.Background(), testTenant(), CreateOrderRequest{
		Order: Order{CustomerEmail: "user@example.com", Amount: 1000, Currency: CurrencyNGN},
	})
	require.NoError(t, err)
	require.Equal(t, "ord-123", result.OrderReference)
	require.Equal(t, int32(1), authCalls.Load())
}

func TestClient_PerTenantTokenCache(t *testing.T) {
	var authCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == PathAuthTokenIssue {
			authCalls.Add(1)
			_ = json.NewEncoder(w).Encode(APIResponse[TokenResponse]{
				Code: ResponseCodeSuccess,
				Data: TokenResponse{
					AccessToken: "token",
					ExpiresAt:   time.Now().Add(30 * time.Minute).Format(time.RFC3339),
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(APIResponse[CreateOrderResult]{
			Code: ResponseCodeSuccess,
			Data: CreateOrderResult{OrderReference: "ord"},
		})
	}))
	defer srv.Close()

	client := testClient(t, srv)
	order := CreateOrderRequest{Order: Order{CustomerEmail: "a@b.com", Amount: 1, Currency: CurrencyNGN}}
	tenant := testTenant()
	_, err := client.CreateOrder(context.Background(), tenant, order)
	require.NoError(t, err)
	_, err = client.CreateOrder(context.Background(), tenant, order)
	require.NoError(t, err)

	other := testTenant()
	other.NombaClientID = "other"
	_, err = client.CreateOrder(context.Background(), other, order)
	require.NoError(t, err)
	require.Equal(t, int32(2), authCalls.Load())
}

func TestClient_RetryOn429(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PathAuthTokenIssue:
			_ = json.NewEncoder(w).Encode(APIResponse[TokenResponse]{
				Code: ResponseCodeSuccess,
				Data: TokenResponse{AccessToken: "token", ExpiresAt: time.Now().Add(30 * time.Minute).Format(time.RFC3339)},
			})
		case PathCheckoutOrder:
			n := attempts.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			_ = json.NewEncoder(w).Encode(APIResponse[CreateOrderResult]{
				Code: ResponseCodeSuccess,
				Data: CreateOrderResult{OrderReference: "ord-retry"},
			})
		}
	}))
	defer srv.Close()

	client := testClient(t, srv)
	result, err := client.CreateOrder(context.Background(), testTenant(), CreateOrderRequest{
		Order: Order{CustomerEmail: "user@example.com", Amount: 500, Currency: CurrencyNGN},
	})
	require.NoError(t, err)
	require.Equal(t, "ord-retry", result.OrderReference)
}

func TestHTTPError_Unwrap(t *testing.T) {
	err := NewHTTPError(http.StatusTooManyRequests, APIError{Code: "99"})
	require.ErrorIs(t, err, ErrRetryable)
}
