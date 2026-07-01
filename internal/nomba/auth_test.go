package nomba

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"go.uber.org/zap"
)

func TestTokenManager_RefreshBeforeExpiry(t *testing.T) {
	var issueCalls atomic.Int32
	var refreshCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PathAuthTokenIssue:
			issueCalls.Add(1)
			_ = json.NewEncoder(w).Encode(APIResponse[TokenResponse]{
				Code: ResponseCodeSuccess,
				Data: TokenResponse{
					AccessToken:  "access-1",
					RefreshToken: "refresh-1",
					ExpiresAt:    time.Now().Add(6 * time.Minute).Format(time.RFC3339),
				},
			})
		case PathAuthTokenRefresh:
			refreshCalls.Add(1)
			_ = json.NewEncoder(w).Encode(APIResponse[TokenResponse]{
				Code: ResponseCodeSuccess,
				Data: TokenResponse{
					AccessToken:  "access-2",
					RefreshToken: "refresh-2",
					ExpiresAt:    time.Now().Add(30 * time.Minute).Format(time.RFC3339),
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClient(zap.NewNop(), srv.Client())
	client.SetTestBaseURL(srv.URL)
	tm := newTokenManager(client, srv.URL, "acct", "client", "secret")

	token1, err := tm.token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "access-1", token1)
	require.Equal(t, int32(1), issueCalls.Load())

	token1b, err := tm.token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "access-1", token1b)
	require.Equal(t, int32(1), issueCalls.Load())

	tm.mu.Lock()
	tm.expiresAt = time.Now().Add(2 * time.Minute)
	tm.mu.Unlock()

	token2, err := tm.token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "access-2", token2)
	require.Equal(t, int32(1), issueCalls.Load())
	require.Equal(t, int32(1), refreshCalls.Load())
}

func TestValidateCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(APIResponse[TokenResponse]{
			Code: ResponseCodeSuccess,
			Data: TokenResponse{AccessToken: "ok", ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339)},
		})
	}))
	defer srv.Close()

	client := NewClient(zap.NewNop(), srv.Client())
	client.SetTestBaseURL(srv.URL)
	err := client.ValidateCredentials(context.Background(), domain.NombaEnvSandbox, "c", "s", "acct")
	require.NoError(t, err)
}
