package nomba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"go.uber.org/zap"
)

type Client struct {
	httpClient *http.Client
	log        *zap.Logger
	testBaseURL string

	mu     sync.RWMutex
	tokens map[string]*tokenManager
}

func NewClient(log *zap.Logger, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &Client{
		httpClient: httpClient,
		log:        log,
		tokens:     make(map[string]*tokenManager),
	}
}

func (c *Client) InvalidateTenant(tenantID uuid.UUID) {
	c.mu.Lock()
	delete(c.tokens, tenantID.String())
	c.mu.Unlock()
}

func (c *Client) SetTestBaseURL(url string) {
	c.testBaseURL = url
}

func (c *Client) baseURLFor(tenant *domain.Tenant) string {
	if c.testBaseURL != "" {
		return c.testBaseURL
	}
	return tenant.NombaBaseURL()
}

func (c *Client) tokenFor(ctx context.Context, tenant *domain.Tenant) (string, error) {
	if tenant == nil {
		return "", fmt.Errorf("tenant is required")
	}
	if tenant.NombaClientSecret == "" {
		return "", fmt.Errorf("tenant nomba client secret not loaded")
	}
	key := tenant.ID.String()

	c.mu.RLock()
	tm, ok := c.tokens[key]
	c.mu.RUnlock()
	if !ok {
		c.mu.Lock()
		tm, ok = c.tokens[key]
		if !ok {
			tm = newTokenManager(c, c.baseURLFor(tenant), tenant.NombaAccountID, tenant.NombaClientID, tenant.NombaClientSecret)
			c.tokens[key] = tm
		}
		c.mu.Unlock()
	}
	return tm.token(ctx)
}

// ValidateCredentials attempts a token issue without caching.
func (c *Client) ValidateCredentials(ctx context.Context, nombaEnv, clientID, clientSecret, accountID string) error {
	baseURL := sandboxBaseURL
	if nombaEnv == domain.NombaEnvProduction {
		baseURL = productionBaseURL
	}
	if c.testBaseURL != "" {
		baseURL = c.testBaseURL
	}
	_, err := c.issueToken(ctx, baseURL, accountID, clientID, clientSecret)
	return err
}

const sandboxBaseURL = "https://sandbox.nomba.com"
const productionBaseURL = "https://api.nomba.com"

func (c *Client) do(ctx context.Context, tenant *domain.Tenant, method, path string, body any, out any) error {
	token, err := c.tokenFor(ctx, tenant)
	if err != nil {
		return err
	}
	return c.doWithToken(ctx, tenant, method, path, body, out, token, true)
}

func (c *Client) doUnauthenticated(ctx context.Context, baseURL, accountID, method, path string, body any, out any) error {
	return c.doWithTokenOnBase(ctx, baseURL, accountID, method, path, body, out, "", false)
}

func (c *Client) doWithToken(ctx context.Context, tenant *domain.Tenant, method, path string, body any, out any, token string, withAuth bool) error {
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		err = c.roundTrip(ctx, c.baseURLFor(tenant), tenant.NombaAccountID, method, path, body, out, token, withAuth)
		if err == nil {
			return nil
		}
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			time.Sleep(parseRetryAfterFromError(httpErr))
			continue
		}
		return err
	}
	return err
}

func (c *Client) doWithTokenOnBase(ctx context.Context, baseURL, accountID, method, path string, body any, out any, token string, withAuth bool) error {
	return c.roundTrip(ctx, baseURL, accountID, method, path, body, out, token, withAuth)
}

func parseRetryAfterFromError(_ *HTTPError) time.Duration {
	return 1 * time.Second
}

func (c *Client) roundTrip(ctx context.Context, baseURL, accountID, method, path string, body any, out any, token string, withAuth bool) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if accountID != "" {
		req.Header.Set(HeaderAccountID, accountID)
	}
	if withAuth {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nomba request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil {
			return nil
		}
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		return nil
	}

	if c.log != nil {
		c.log.Warn("nomba api http error",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", resp.StatusCode),
			zap.ByteString("response_body", respBody),
		)
	}

	var apiErr APIError
	if err := json.Unmarshal(respBody, &apiErr); err != nil || apiErr.Code == "" {
		return HTTPErrorFromNombaBody(resp.StatusCode, respBody)
	}
	return NewHTTPError(resp.StatusCode, apiErr)
}

func doData[T any](c *Client, ctx context.Context, tenant *domain.Tenant, method, path string, body any) (T, error) {
	var zero T
	var resp APIResponse[T]
	if err := c.do(ctx, tenant, method, path, body, &resp); err != nil {
		return zero, err
	}
	if !resp.OK() {
		return zero, NewHTTPError(http.StatusOK, APIError{Code: resp.Code, Description: resp.Description})
	}
	return resp.Data, nil
}

func doUnauthenticatedData[T any](c *Client, ctx context.Context, baseURL, accountID, method, path string, body any) (T, error) {
	var zero T
	var resp APIResponse[T]
	if err := c.doUnauthenticated(ctx, baseURL, accountID, method, path, body, &resp); err != nil {
		return zero, err
	}
	if !resp.OK() {
		return zero, NewHTTPError(http.StatusOK, APIError{Code: resp.Code, Description: resp.Description})
	}
	return resp.Data, nil
}

func (c *Client) issueToken(ctx context.Context, baseURL, accountID, clientID, clientSecret string) (TokenResponse, error) {
	return doUnauthenticatedData[TokenResponse](c, ctx, baseURL, accountID, "POST", PathAuthTokenIssue, IssueTokenRequest{
		GrantType:    GrantTypeClientCredentials,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
}

func (c *Client) refreshToken(ctx context.Context, baseURL, accountID, refreshToken string) (TokenResponse, error) {
	return doUnauthenticatedData[TokenResponse](c, ctx, baseURL, accountID, "POST", PathAuthTokenRefresh, RefreshTokenRequest{
		GrantType:    GrantTypeRefreshToken,
		RefreshToken: refreshToken,
	})
}
