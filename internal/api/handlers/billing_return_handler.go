package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/portalpage"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type BillingReturnHandler struct {
	svc      *service.BillingReturnService
	renderer *portalpage.Renderer
}

func NewBillingReturnHandler(svc *service.BillingReturnService, renderer *portalpage.Renderer) *BillingReturnHandler {
	return &BillingReturnHandler{svc: svc, renderer: renderer}
}

func (h *BillingReturnHandler) Success(c *gin.Context) {
	orderReference := strings.TrimSpace(c.Query("orderReference"))
	orderID := strings.TrimSpace(c.Query("orderId"))

	view, err := h.svc.Resolve(c.Request.Context(), orderReference, orderID)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	if wantsJSON(c) {
		dto.RespondOK(c, view)
		return
	}

	title := "Payment status"
	tenantName := view.TenantName
	if tenantName == "" {
		tenantName = "SubSync"
	}
	if view.StatusLabel != "" {
		title = view.StatusLabel
	}

	data := portalpage.BillingSuccessData{
		Title:             title,
		TenantName:        tenantName,
		PlanName:          view.PlanName,
		StatusLabel:       view.StatusLabel,
		StatusMessage:     view.StatusMessage,
		Outcome:           string(view.Outcome),
		OutcomeBadge:      billingOutcomeBadge(view.Outcome),
		OrderReference:    view.OrderReference,
		OrderID:           view.OrderID,
		SubscriptionState: view.SubscriptionState,
		AmountDisplay:     view.AmountDisplay,
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_ = h.renderer.RenderBillingSuccess(c.Writer, data)
}

func billingOutcomeBadge(outcome service.BillingReturnOutcome) string {
	switch outcome {
	case service.BillingReturnSuccess:
		return "ok"
	case service.BillingReturnPending:
		return "warn"
	default:
		return ""
	}
}
