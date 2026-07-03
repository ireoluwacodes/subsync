package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type AuthHandler struct {
	cfg *config.Config
	svc *service.AuthService
}

func NewAuthHandler(cfg *config.Config, svc *service.AuthService) *AuthHandler {
	return &AuthHandler{cfg: cfg, svc: svc}
}

type registerRequest struct {
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password" binding:"required,min=8"`
	Name              string `json:"name" binding:"required"`
	NombaClientID     string `json:"nomba_client_id" binding:"required"`
	NombaClientSecret string `json:"nomba_client_secret" binding:"required"`
	NombaAccountID    string `json:"nomba_account_id" binding:"required"`
	NombaSubAccountID  string `json:"nomba_sub_account_id"`
	NombaEnv           string `json:"nomba_env" binding:"required"`
	NombaWebhookSecret string `json:"nomba_webhook_secret"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	result, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Email:             req.Email,
		Password:          req.Password,
		Name:              req.Name,
		NombaClientID:     req.NombaClientID,
		NombaClientSecret: req.NombaClientSecret,
		NombaAccountID:    req.NombaAccountID,
		NombaSubAccountID:  req.NombaSubAccountID,
		NombaEnv:           req.NombaEnv,
		NombaWebhookSecret: req.NombaWebhookSecret,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	setRefreshTokenCookie(c, h.cfg, result.RefreshToken)

	dto.RespondCreated(c, gin.H{
		"user":          dto.UserToResponse(result.User),
		"tenant":        dto.TenantToResponse(result.Tenant, false),
		"api_key":       result.APIKey,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_at":    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"nomba": dto.NombaOnboardingFromTenant(result.Tenant, h.svc.NombaWebhookURL(result.Tenant.ID)),
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	setRefreshTokenCookie(c, h.cfg, result.RefreshToken)

	dto.RespondOK(c, gin.H{
		"user":          dto.UserToResponse(result.User),
		"tenant":        dto.TenantToResponse(result.Tenant, false),
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_at":    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"nomba": dto.NombaOnboardingFromTenant(result.Tenant, h.svc.NombaWebhookURL(result.Tenant.ID)),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken := refreshTokenFromRequest(c)
	if refreshToken == "" {
		dto.RespondError(c, dto.NewBindError("refresh_token required (cookie or query)"))
		return
	}

	result, err := h.svc.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	setRefreshTokenCookie(c, h.cfg, result.RefreshToken)

	dto.RespondOK(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_at":    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	user, ok := middlewareUser(c)
	if !ok {
		return
	}
	if err := h.svc.Logout(c.Request.Context(), user.ID); err != nil {
		dto.RespondError(c, err)
		return
	}
	clearRefreshTokenCookie(c)
	dto.RespondOK(c, gin.H{"ok": true})
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	otp, _ := h.svc.ForgotPassword(c.Request.Context(), req.Email)
	resp := gin.H{"ok": true}
	if otp != "" {
		resp["otp"] = otp // dev stub when email is disabled
	}
	dto.RespondOK(c, resp)
}

type confirmPasswordOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

func (h *AuthHandler) ConfirmPasswordOTP(c *gin.Context) {
	var req confirmPasswordOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	resetToken, err := h.svc.ConfirmPasswordOTP(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, gin.H{
		"ok":          true,
		"reset_token": resetToken,
	})
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, gin.H{"ok": true})
}
