package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type PortalHandler struct {
	svc *service.PortalService
}

func NewPortalHandler(svc *service.PortalService) *PortalHandler {
	return &PortalHandler{svc: svc}
}

func (h *PortalHandler) CreateToken(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	var req struct {
		SubscriptionID   uuid.UUID `json:"subscription_id" binding:"required"`
		ExpiresInHours   int       `json:"expires_in_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	result, err := h.svc.CreateToken(c.Request.Context(), tenant.ID, service.CreatePortalTokenInput{
		SubscriptionID: req.SubscriptionID,
		ExpiresInHours: req.ExpiresInHours,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondCreated(c, result)
}

func (h *PortalHandler) Home(c *gin.Context) {
	view, err := h.svc.Home(c.Request.Context(), c.Param("token"))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, view)
}

func (h *PortalHandler) Cancel(c *gin.Context) {
	var req service.PortalCancelInput
	_ = c.ShouldBindJSON(&req)
	sub, err := h.svc.Cancel(c.Request.Context(), c.Param("token"), req)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, sub)
}

func (h *PortalHandler) UpdatePaymentMethod(c *gin.Context) {
	result, err := h.svc.StartPaymentMethodUpdate(c.Request.Context(), c.Param("token"))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, result)
}
