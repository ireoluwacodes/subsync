package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type SettingsHandler struct {
	svc *service.SettingsService
}

func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	fresh, err := h.svc.Get(c.Request.Context(), tenant.ID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.TenantToResponse(fresh, false))
}

func (h *SettingsHandler) UpdateGeneral(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Website string `json:"website"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	updated, err := h.svc.UpdateGeneral(c.Request.Context(), tenant.ID, service.UpdateGeneralInput{
		Name: req.Name, Email: req.Email, Website: req.Website,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.TenantToResponse(updated, false))
}

func (h *SettingsHandler) UpdateDunning(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var req struct {
		Steps []map[string]any `json:"steps"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	updated, err := h.svc.UpdateDunning(c.Request.Context(), tenant.ID, map[string]any{"steps": req.Steps})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.TenantToResponse(updated, false))
}

func (h *SettingsHandler) UpdateBranding(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	updated, err := h.svc.UpdateBranding(c.Request.Context(), tenant.ID, req)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.TenantToResponse(updated, false))
}

func (h *SettingsHandler) UpdateBillingEmail(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	updated, err := h.svc.UpdateBillingEmail(c.Request.Context(), tenant.ID, req)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.TenantToResponse(updated, false))
}

func nombaSettingsResponse(v *service.NombaSettingsView) dto.NombaSettingsResponse {
	return dto.NombaSettingsResponse{
		WebhookURL:                   v.WebhookURL,
		NombaClientID:                v.NombaClientID,
		NombaAccountID:               v.NombaAccountID,
		NombaSubAccountID:            v.NombaSubAccountID,
		NombaEnv:                     v.NombaEnv,
		NombaWebhookSecretConfigured: v.NombaWebhookSecretConfigured,
	}
}

func (h *SettingsHandler) GetNomba(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	view, err := h.svc.GetNomba(c.Request.Context(), tenant.ID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, nombaSettingsResponse(view))
}

func (h *SettingsHandler) UpdateNomba(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var req service.UpdateNombaInput
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	updated, err := h.svc.UpdateNomba(c.Request.Context(), tenant.ID, req)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, nombaSettingsResponse(updated))
}

func (h *SettingsHandler) RotateAPIKey(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	apiKey, updated, err := h.svc.RotateAPIKey(c.Request.Context(), tenant.ID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, gin.H{
		"api_key": apiKey,
		"tenant":  dto.TenantToResponse(updated, false),
	})
}
