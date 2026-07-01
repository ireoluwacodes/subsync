package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type Claims struct {
	jwt.RegisteredClaims
	TenantID     string `json:"tenant_id,omitempty"`
	Email        string `json:"email,omitempty"`
	TokenVersion int    `json:"token_version"`
	Type         string `json:"type"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type JWTService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		secret:     []byte(cfg.JWTSecret),
		accessTTL:  cfg.JWTAccessTTL,
		refreshTTL: cfg.JWTRefreshTTL,
	}
}

func (j *JWTService) IssuePair(user *domain.User, tenantID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(j.accessTTL)
	refreshExp := now.Add(j.refreshTTL)

	access, err := j.sign(Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		TenantID:     tenantID.String(),
		Email:        user.Email,
		TokenVersion: user.TokenVersion,
		Type:         tokenTypeAccess,
	})
	if err != nil {
		return nil, err
	}

	refresh, err := j.sign(Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		TokenVersion: user.TokenVersion,
		Type:         tokenTypeRefresh,
	})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    accessExp,
	}, nil
}

func (j *JWTService) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (j *JWTService) sign(claims Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.secret)
}
