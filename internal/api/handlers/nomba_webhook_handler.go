package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type NombaWebhookHandler struct {
	cfg     *config.Config
	tenants *service.TenantService
	events  *service.NombaEventService
}

func NewNombaWebhookHandler(cfg *config.Config, tenants *service.TenantService, events *service.NombaEventService) *NombaWebhookHandler {
	return &NombaWebhookHandler{cfg: cfg, tenants: tenants, events: events}
}

func (h *NombaWebhookHandler) Receive(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("tenant_id"))
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid tenant_id"))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	tenant, err := h.tenants.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	if err := h.tenants.LoadNombaWebhookSecret(c.Request.Context(), tenant); err != nil {
		dto.RespondError(c, err)
		return
	}

	signature := c.GetHeader("nomba-signature")
	if signature == "" {
		signature = c.GetHeader("nomba-sig-value")
	}
	timestamp := c.GetHeader("nomba-timestamp")

	secret := tenant.NombaWebhookSecret
	if secret == "" && h.cfg != nil {
		secret = h.cfg.NombaWebhookSigningKey
	}
	if secret == "" {
		dto.RespondError(c, domain.ErrUnauthorized)
		return
	}

	if timestamp != "" {
		if err := nomba.ValidateWebhookTimestamp(timestamp, time.Now().UTC()); err != nil {
			dto.RespondError(c, domain.ErrUnauthorized)
			return
		}
	}

	if err := nomba.VerifyWebhookSignature(body, signature, secret, timestamp); err != nil {
		dto.RespondError(c, domain.ErrUnauthorized)
		return
	}

	if err := h.events.ProcessInbound(c.Request.Context(), tenantID, body); err != nil {
		dto.RespondError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
