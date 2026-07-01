package nomba

import (
	"context"
	"sync"
	"time"
)

const tokenRefreshSkew = 5 * time.Minute

type tokenManager struct {
	client       *Client
	baseURL      string
	accountID    string
	clientID     string
	clientSecret string

	mu           sync.Mutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

func newTokenManager(client *Client, baseURL, accountID, clientID, clientSecret string) *tokenManager {
	return &tokenManager{
		client:       client,
		baseURL:      baseURL,
		accountID:    accountID,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (tm *tokenManager) token(ctx context.Context) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.accessToken != "" && time.Now().Add(tokenRefreshSkew).Before(tm.expiresAt) {
		return tm.accessToken, nil
	}

	if tm.refreshToken != "" {
		resp, err := tm.client.refreshToken(ctx, tm.baseURL, tm.accountID, tm.refreshToken)
		if err == nil {
			tm.store(resp)
			return tm.accessToken, nil
		}
	}

	resp, err := tm.client.issueToken(ctx, tm.baseURL, tm.accountID, tm.clientID, tm.clientSecret)
	if err != nil {
		return "", err
	}
	tm.store(resp)
	return tm.accessToken, nil
}

func (tm *tokenManager) store(resp TokenResponse) {
	tm.accessToken = resp.AccessToken
	tm.refreshToken = resp.RefreshToken
	if t, err := time.Parse(time.RFC3339, resp.ExpiresAt); err == nil {
		tm.expiresAt = t
	} else {
		tm.expiresAt = time.Now().Add(30 * time.Minute)
	}
}
