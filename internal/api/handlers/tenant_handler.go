package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type TenantHandler struct {
	svc *service.TenantService
}

func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
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
