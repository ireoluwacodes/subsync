package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/auth"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
)

type AuthService struct {
	users         domain.UserRepository
	tenants       domain.TenantRepository
	resets        domain.PasswordResetRepository
	jwt           *auth.JWTService
	nomba         *nomba.Client
	tenantSvc     *TenantService
	publicBaseURL string
}

func NewAuthService(
	users domain.UserRepository,
	tenants domain.TenantRepository,
	resets domain.PasswordResetRepository,
	jwt *auth.JWTService,
	nombaClient *nomba.Client,
	tenantSvc *TenantService,
	publicBaseURL string,
) *AuthService {
	return &AuthService{
		users:         users,
		tenants:       tenants,
		resets:        resets,
		jwt:           jwt,
		nomba:         nombaClient,
		tenantSvc:     tenantSvc,
		publicBaseURL: publicBaseURL,
	}
}

type RegisterInput struct {
	Email             string
	Password          string
	Name              string
	NombaClientID     string
	NombaClientSecret string
	NombaAccountID    string
	NombaSubAccountID  string
	NombaEnv           string
	NombaWebhookSecret string
}

type AuthResult struct {
	User         *domain.User
	Tenant       *domain.Tenant
	APIKey       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	if in.Email == "" || in.Password == "" || in.Name == "" {
		return nil, fmt.Errorf("%w: email, password, and name are required", domain.ErrValidation)
	}

	if _, err := s.users.GetByEmail(ctx, strings.ToLower(in.Email)); err == nil {
		return nil, domain.ErrConflict
	} else if err != nil && err != domain.ErrNotFound {
		return nil, err
	}

	result, err := s.tenantSvc.CreateTenant(ctx, CreateTenantInput{
		Name:              in.Name,
		Email:             strings.ToLower(in.Email),
		NombaClientID:     in.NombaClientID,
		NombaClientSecret: in.NombaClientSecret,
		NombaAccountID:    in.NombaAccountID,
		NombaSubAccountID: in.NombaSubAccountID,
		NombaEnv:          in.NombaEnv,
		NombaWebhookSecret: in.NombaWebhookSecret,
	})
	if err != nil {
		return nil, err
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		TenantID:     result.Tenant.ID,
		Email:        strings.ToLower(in.Email),
		PasswordHash: hash,
		Name:         in.Name,
		TokenVersion: 1,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := s.jwt.IssuePair(user, result.Tenant.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		Tenant:       result.Tenant,
		APIKey:       result.APIKey,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, domain.ErrNotFound
	}

	tenant, err := s.tenants.GetByID(ctx, user.TenantID)
	if err != nil {
		return nil, err
	}

	tokens, err := s.jwt.IssuePair(user, tenant.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		Tenant:       tenant,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := s.jwt.Parse(refreshToken)
	if err != nil || claims.Type != "refresh" {
		return nil, domain.ErrNotFound
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TokenVersion != claims.TokenVersion {
		return nil, domain.ErrNotFound
	}

	tenant, err := s.tenants.GetByID(ctx, user.TenantID)
	if err != nil {
		return nil, err
	}

	tokens, err := s.jwt.IssuePair(user, tenant.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		Tenant:       tenant,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.users.BumpTokenVersion(ctx, userID)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return "", nil // silent success
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	if err := s.resets.Create(ctx, &domain.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}); err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if token == "" || newPassword == "" {
		return fmt.Errorf("%w: token and password are required", domain.ErrValidation)
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	reset, err := s.resets.GetValidByTokenHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrNotFound
	}

	pwHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, reset.UserID, pwHash); err != nil {
		return err
	}
	if err := s.users.BumpTokenVersion(ctx, reset.UserID); err != nil {
		return err
	}
	return s.resets.MarkUsed(ctx, reset.ID)
}

func (s *AuthService) NombaWebhookURL(tenantID uuid.UUID) string {
	return NombaWebhookURL(s.publicBaseURL, tenantID)
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, token string) (*domain.User, *domain.Tenant, error) {
	claims, err := s.jwt.Parse(token)
	if err != nil || claims.Type != "access" {
		return nil, nil, domain.ErrNotFound
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, nil, domain.ErrNotFound
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user.TokenVersion != claims.TokenVersion {
		return nil, nil, domain.ErrNotFound
	}

	tenant, err := s.tenants.GetByID(ctx, user.TenantID)
	if err != nil {
		return nil, nil, err
	}

	return user, tenant, nil
}
