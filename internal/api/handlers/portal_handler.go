package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/portalpage"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type PortalHandler struct {
	svc      *service.PortalService
	renderer *portalpage.Renderer
}

func NewPortalHandler(svc *service.PortalService, renderer *portalpage.Renderer) *PortalHandler {
	return &PortalHandler{svc: svc, renderer: renderer}
}

func (h *PortalHandler) CreateToken(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	var req struct {
		SubscriptionID uuid.UUID `json:"subscription_id" binding:"required"`
		ExpiresInHours int       `json:"expires_in_hours"`
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
	token := c.Param("token")
	if wantsJSON(c) {
		view, err := h.svc.Home(c.Request.Context(), token)
		if err != nil {
			dto.RespondError(c, err)
			return
		}
		dto.RespondOK(c, view)
		return
	}
	view, err := h.svc.Home(c.Request.Context(), token)
	if err != nil {
		c.String(http.StatusNotFound, "This billing link is invalid or has expired.")
		return
	}
	data := portalpage.HomeData{
		Title:                   "Manage subscription",
		Token:                   token,
		TenantName:              view.TenantName,
		PlanName:                view.PlanName,
		CustomerEmail:           view.CustomerEmail,
		State:                   string(view.Subscription.State),
		CancelAtPeriodEnd:       view.CancelAtPeriodEnd,
		CurrentPeriodStart:      view.CurrentPeriodStart,
		CurrentPeriodEnd:        view.CurrentPeriodEnd,
		CanManagePaymentMethods: view.CanManagePaymentMethods,
		ShowCancelForm:          view.ShowCancelForm,
		AwaitingPaymentMethod:   view.AwaitingPaymentMethod,
		HasCard:                 view.HasCard,
		HasMandate:              view.HasMandate,
		MandateStatus:           view.MandateStatus,
		CanSetupDirectDebit:     view.CanSetupDirectDebit,
		PaymentMethodLast4:      view.PaymentMethodLast4,
		PaymentMethodBrand:      view.PaymentMethodBrand,
		FlashMessage:            c.Query("msg"),
		FlashError:              c.Query("err"),
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_ = h.renderer.RenderHome(c.Writer, data)
}

func (h *PortalHandler) AddCard(c *gin.Context) {
	result, err := h.svc.StartPaymentMethodUpdate(c.Request.Context(), c.Param("token"))
	if err != nil {
		if wantsJSON(c) {
			dto.RespondError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/portal/"+c.Param("token")+"?err=Could+not+start+card+setup")
		return
	}
	c.Redirect(http.StatusFound, result.CheckoutLink)
}

func (h *PortalHandler) UpdatePaymentMethod(c *gin.Context) {
	result, err := h.svc.StartPaymentMethodUpdate(c.Request.Context(), c.Param("token"))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, result)
}

func (h *PortalHandler) DirectDebitForm(c *gin.Context) {
	token := c.Param("token")
	view, err := h.svc.Home(c.Request.Context(), token)
	if err != nil {
		c.String(http.StatusNotFound, "This billing link is invalid or has expired.")
		return
	}
	customerName := ""
	if view.Subscription != nil {
		if customer, cErr := h.svc.CustomerName(c.Request.Context(), view.Subscription.TenantID, view.Subscription.CustomerID); cErr == nil {
			customerName = customer
		}
	}
	banks, banksErr := h.svc.ListBanksForPortal(c.Request.Context(), token)
	banksLoadError := ""
	if banksErr != nil {
		banksLoadError = "Could not load the bank list. Please try again in a moment."
	}
	data := portalpage.DirectDebitFormData{
		Title:          "Set up direct debit",
		Token:          token,
		TenantName:     view.TenantName,
		PlanName:       view.PlanName,
		CustomerEmail:  view.CustomerEmail,
		CustomerName:   customerName,
		Banks:          portalpage.BanksFromNomba(banks),
		BanksLoadError: banksLoadError,
		FlashError:     c.Query("err"),
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = h.renderer.RenderDirectDebitForm(c.Writer, data)
}

func (h *PortalHandler) DirectDebitSubmit(c *gin.Context) {
	token := c.Param("token")
	in := service.DirectDebitSetupInput{
		CustomerAccountNumber: c.PostForm("customer_account_number"),
		BankCode:              c.PostForm("bank_code"),
		CustomerName:          c.PostForm("customer_name"),
		CustomerAccountName:   c.PostForm("customer_account_name"),
		CustomerEmail:         c.PostForm("customer_email"),
		CustomerPhone:         c.PostForm("customer_phone"),
		CustomerAddress:       c.PostForm("customer_address"),
	}
	if in.CustomerAccountNumber == "" || in.BankCode == "" || in.CustomerName == "" {
		c.Redirect(http.StatusSeeOther, "/portal/"+token+"/direct-debit?err=Please+fill+in+all+required+fields")
		return
	}
	_, err := h.svc.StartDirectDebitSetup(c.Request.Context(), token, in)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/portal/"+token+"/direct-debit?err="+url.QueryEscape(err.Error()))
		return
	}
	c.Redirect(http.StatusSeeOther, "/portal/"+token+"/direct-debit/pending")
}

func (h *PortalHandler) DirectDebitPending(c *gin.Context) {
	token := c.Param("token")
	view, err := h.svc.Home(c.Request.Context(), token)
	if err != nil {
		c.String(http.StatusNotFound, "This billing link is invalid or has expired.")
		return
	}
	status, err := h.svc.GetDirectDebitStatus(c.Request.Context(), token)
	instructions := view.MandateInstructions
	ready := false
	mandateStatus := "pending"
	if err == nil && status != nil {
		ready = status.Ready
		mandateStatus = status.MandateStatus
		if status.Instructions != "" {
			instructions = status.Instructions
		}
	}
	if ready {
		c.Redirect(http.StatusSeeOther, "/portal/"+token+"?msg=Direct+debit+is+ready")
		return
	}
	data := portalpage.DirectDebitPendingData{
		Title:         "Complete direct debit",
		Token:         token,
		TenantName:    view.TenantName,
		Instructions:  instructions,
		MandateStatus: mandateStatus,
		Ready:         ready,
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = h.renderer.RenderDirectDebitPending(c.Writer, data)
}

func (h *PortalHandler) Cancel(c *gin.Context) {
	var req service.PortalCancelInput
	if wantsJSON(c) {
		_ = c.ShouldBindJSON(&req)
	}
	token := c.Param("token")
	sub, err := h.svc.Cancel(c.Request.Context(), token, req)
	if err != nil {
		if wantsJSON(c) {
			dto.RespondError(c, err)
			return
		}
		c.Redirect(http.StatusSeeOther, "/portal/"+token+"?err=Could+not+cancel+subscription")
		return
	}
	if wantsJSON(c) {
		dto.RespondOK(c, sub)
		return
	}
	msg := "Your subscription will cancel at the end of your billing period"
	if sub.State == domain.SubscriptionStateCanceled {
		msg = "Subscription canceled"
	}
	c.Redirect(http.StatusSeeOther, "/portal/"+token+"?msg="+url.QueryEscape(msg))
}

func wantsJSON(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "application/json")
}

func boolPtr(v bool) *bool { return &v }
