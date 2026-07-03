package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type WebhookHandler struct {
	svc *service.WebhookService
}

func NewWebhookHandler(svc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

func (h *WebhookHandler) CreateEndpoint(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	var req struct {
		URL      string   `json:"url" binding:"required"`
		Events   []string `json:"events"`
		IsActive *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	ep, err := h.svc.CreateEndpoint(c.Request.Context(), tenant.ID, service.CreateWebhookEndpointInput{
		URL: req.URL, Events: req.Events, IsActive: active,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondCreated(c, ep)
}

func (h *WebhookHandler) ListEndpoints(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	eps, err := h.svc.ListEndpoints(c.Request.Context(), tenant.ID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, eps)
}

func (h *WebhookHandler) GetEndpoint(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid id"))
		return
	}
	ep, err := h.svc.GetEndpoint(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, ep)
}

func (h *WebhookHandler) UpdateEndpoint(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid id"))
		return
	}
	var req struct {
		URL      *string  `json:"url"`
		Events   []string `json:"events"`
		IsActive *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	ep, err := h.svc.UpdateEndpoint(c.Request.Context(), tenant.ID, id, service.UpdateWebhookEndpointInput{
		URL: req.URL, Events: req.Events, IsActive: req.IsActive,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, ep)
}

func (h *WebhookHandler) DeleteEndpoint(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid id"))
		return
	}
	if err := h.svc.DeleteEndpoint(c.Request.Context(), tenant.ID, id); err != nil {
		dto.RespondError(c, err)
		return
	}
	c.Status(204)
}

func (h *WebhookHandler) ListDeliveries(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid id"))
		return
	}
	deliveries, err := h.svc.ListDeliveries(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, deliveries)
}
