package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/auth"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type AuthService struct {
	users         domain.UserRepository
	tenants       domain.TenantRepository
	resets        domain.PasswordResetRepository
	jwt           *auth.JWTService
	nomba         *nomba.Client
	tenantSvc     *TenantService
	publicBaseURL string
	mailer        *email.MailerService
	cfg           *config.Config
}

func NewAuthService(
	users domain.UserRepository,
	tenants domain.TenantRepository,
	resets domain.PasswordResetRepository,
	jwt *auth.JWTService,
	nombaClient *nomba.Client,
	tenantSvc *TenantService,
	publicBaseURL string,
	mailer *email.MailerService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		users:         users,
		tenants:       tenants,
		resets:        resets,
		jwt:           jwt,
		nomba:         nombaClient,
		tenantSvc:     tenantSvc,
		publicBaseURL: publicBaseURL,
		mailer:        mailer,
		cfg:           cfg,
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

const (
	passwordResetOTPTTL   = 10 * time.Minute
	passwordResetTokenTTL = 15 * time.Minute
)

func (s *AuthService) ForgotPassword(ctx context.Context, emailAddr string) (string, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(emailAddr))
	if err != nil {
		return "", nil // silent success
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return "", err
	}

	if err := s.resets.InvalidateUnusedForUser(ctx, user.ID); err != nil {
		return "", err
	}

	if err := s.resets.Create(ctx, &domain.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: utils.HashResetSecret(otp),
		ExpiresAt: time.Now().Add(passwordResetOTPTTL),
	}); err != nil {
		return "", err
	}

	if s.mailer != nil && s.mailer.Enabled() {
		subject, html := email.PasswordResetOTPHTML(otp)
		_ = s.mailer.Send(ctx, user.Email, subject, html)
		return "", nil
	}

	return otp, nil
}

func (s *AuthService) ConfirmPasswordOTP(ctx context.Context, emailAddr, otp string) (string, error) {
	if emailAddr == "" || otp == "" {
		return "", fmt.Errorf("%w: email and otp are required", domain.ErrValidation)
	}

	user, err := s.users.GetByEmail(ctx, strings.ToLower(emailAddr))
	if err != nil {
		return "", domain.ErrNotFound
	}

	reset, err := s.resets.GetLatestValidByUserID(ctx, user.ID)
	if err != nil {
		return "", domain.ErrNotFound
	}
	if reset.TokenHash != utils.HashResetSecret(strings.TrimSpace(otp)) {
		return "", domain.ErrNotFound
	}

	resetToken, err := utils.GenerateResetToken()
	if err != nil {
		return "", err
	}

	if err := s.resets.UpdateTokenHash(ctx, reset.ID, utils.HashResetSecret(resetToken), time.Now().Add(passwordResetTokenTTL)); err != nil {
		return "", err
	}

	return resetToken, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if token == "" || newPassword == "" {
		return fmt.Errorf("%w: token and password are required", domain.ErrValidation)
	}
	tokenHash := utils.HashResetSecret(token)

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
	return utils.NombaWebhookURL(s.publicBaseURL, tenantID)
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
