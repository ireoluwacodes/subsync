package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type TenantHandler struct {
	svc *service.TenantService
	cfg *config.Config
}

func NewTenantHandler(svc *service.TenantService, cfg *config.Config) *TenantHandler {
	return &TenantHandler{svc: svc, cfg: cfg}
}

func (h *TenantHandler) Create(c *gin.Context) {
	if h.cfg.BootstrapSecret == "" || c.GetHeader("X-Bootstrap-Secret") != h.cfg.BootstrapSecret {
		c.JSON(http.StatusUnauthorized, dto.Envelope{
			Meta:  dto.Meta{RequestID: c.GetString("request_id")},
			Error: &dto.APIError{Code: "unauthorized", Message: "invalid bootstrap secret"},
		})
		return
	}

	var req dto.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	result, err := h.svc.CreateTenant(c.Request.Context(), service.CreateTenantInput{
		Name:              req.Name,
		Email:             req.Email,
		NombaClientID:     req.NombaClientID,
		NombaClientSecret: req.NombaClientSecret,
		NombaAccountID:    req.NombaAccountID,
		NombaSubAccountID: req.NombaSubAccountID,
		NombaEnv:          req.NombaEnv,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondCreated(c, dto.CreateTenantResponse{
		Tenant: dto.TenantToResponse(result.Tenant, true),
		APIKey: result.APIKey,
	})
}

func (h *TenantHandler) Me(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		c.JSON(401, dto.Envelope{
			Meta:  dto.Meta{RequestID: c.GetString("request_id")},
			Error: &dto.APIError{Code: "unauthorized", Message: "tenant not found in context"},
		})
		return
	}

	fresh, err := h.svc.GetTenant(c.Request.Context(), tenant.ID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondOK(c, dto.TenantToResponse(fresh, false))
}
