package nomba

const (
	GrantTypeClientCredentials = "client_credentials"
	GrantTypeRefreshToken      = "refresh_token"
)

// IssueTokenRequest is the body for POST /v1/auth/token/issue.
type IssueTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// RefreshTokenRequest is the body for POST /v1/auth/token/refresh.
type RefreshTokenRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// RevokeTokenRequest is the body for POST /v1/auth/token/revoke.
type RevokeTokenRequest struct {
	ClientID    string `json:"clientId"`
	AccessToken string `json:"access_token"`
}

// TokenResponse is the data payload from token issue/refresh.
type TokenResponse struct {
	BusinessID   string `json:"businessId"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expiresAt"`
}
