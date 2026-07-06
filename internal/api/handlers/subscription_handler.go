package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type SubscriptionHandler struct {
	svc      *service.SubscriptionService
	checkout *service.SubscriptionCheckoutService
}

func NewSubscriptionHandler(svc *service.SubscriptionService, checkout *service.SubscriptionCheckoutService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, checkout: checkout}
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var req struct {
		CustomerID      string `json:"customer_id" binding:"required"`
		PlanID          string `json:"plan_id" binding:"required"`
		PaymentMethodID string `json:"payment_method_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid customer_id"))
		return
	}
	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid plan_id"))
		return
	}
	var pmID *uuid.UUID
	if req.PaymentMethodID != "" {
		id, err := uuid.Parse(req.PaymentMethodID)
		if err != nil {
			dto.RespondError(c, dto.NewBindError("invalid payment_method_id"))
			return
		}
		pmID = &id
	}

	sub, err := h.svc.Create(c.Request.Context(), tenant.ID, service.CreateSubscriptionInput{
		CustomerID: customerID, PlanID: planID, PaymentMethodID: pmID,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondCreated(c, dto.SubscriptionToResponse(sub))
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var pagination dto.PaginationParams
	_ = c.ShouldBindQuery(&pagination)
	pagination.Normalize()

	filter := domain.SubscriptionListFilter{
		State:  c.Query("state"),
		Limit:  pagination.PerPage,
		Offset: (pagination.Page - 1) * pagination.PerPage,
		Sort:   c.Query("sort"),
	}
	if v := c.Query("customer_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			filter.CustomerID = &id
		}
	}
	if v := c.Query("plan_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			filter.PlanID = &id
		}
	}

	subs, total, err := h.svc.List(c.Request.Context(), tenant.ID, filter)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	out := make([]dto.SubscriptionResponse, len(subs))
	for i, s := range subs {
		out[i] = dto.SubscriptionToResponse(s)
	}
	c.JSON(200, dto.Envelope{
		Data: out,
		Meta: dto.Meta{RequestID: c.GetString("request_id"), Page: pagination.Page, PerPage: pagination.PerPage, Total: total},
	})
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	sub, err := h.svc.Get(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.SubscriptionToResponse(sub))
}

func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	var req struct {
		CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
		Reason            string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	sub, err := h.svc.Cancel(c.Request.Context(), tenant.ID, id, service.CancelInput{
		CancelAtPeriodEnd: req.CancelAtPeriodEnd,
		Reason:            req.Reason,
	}, actorFromContext(c))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.SubscriptionToResponse(sub))
}

func (h *SubscriptionHandler) Pause(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	var req struct {
		PauseEndsAt *time.Time `json:"pause_ends_at"`
	}
	_ = c.ShouldBindJSON(&req)

	sub, err := h.svc.Pause(c.Request.Context(), tenant.ID, id, req.PauseEndsAt, actorFromContext(c))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.SubscriptionToResponse(sub))
}

func (h *SubscriptionHandler) Resume(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	sub, err := h.svc.Resume(c.Request.Context(), tenant.ID, id, actorFromContext(c))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.SubscriptionToResponse(sub))
}

func (h *SubscriptionHandler) Upgrade(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	var req struct {
		NewPlanID string `json:"new_plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	planID, err := uuid.Parse(req.NewPlanID)
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid new_plan_id"))
		return
	}

	sub, invoice, err := h.svc.Upgrade(c.Request.Context(), tenant.ID, id, service.UpgradeInput{NewPlanID: planID}, actorFromContext(c))
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, gin.H{
		"subscription": dto.SubscriptionToResponse(sub),
		"invoice":      dto.InvoiceToResponse(invoice),
	})
}

func (h *SubscriptionHandler) PreviewUpgrade(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	planID, err := uuid.Parse(c.Query("new_plan_id"))
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid new_plan_id"))
		return
	}
	result, currency, err := h.svc.PreviewUpgrade(c.Request.Context(), tenant.ID, id, service.UpgradeInput{NewPlanID: planID})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.ProrationToResponse(result, currency))
}

func (h *SubscriptionHandler) Transitions(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	transitions, err := h.svc.ListTransitions(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, transitions)
}

func actorFromContext(c *gin.Context) string {
	if user, ok := middlewareUser(c); ok {
		return user.ID.String()
	}
	return "api_key"
}

// ListForCustomer is used by nested customer routes.
func (h *SubscriptionHandler) ListForCustomer(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	customerID, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	subs, _, err := h.svc.List(c.Request.Context(), tenant.ID, domain.SubscriptionListFilter{
		CustomerID: &customerID,
		Limit:      limit,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	out := make([]dto.SubscriptionResponse, len(subs))
	for i, s := range subs {
		out[i] = dto.SubscriptionToResponse(s)
	}
	dto.RespondOK(c, out)
}

type subscriptionCheckoutRequest struct {
	CustomerID            string   `json:"customer_id" binding:"required"`
	PlanID                string   `json:"plan_id" binding:"required"`
	SuccessURL            string   `json:"success_url"`
	CancelURL             string   `json:"cancel_url"`
	SendCheckoutEmail     bool     `json:"send_checkout_email"`
	CardOnly              bool     `json:"card_only"`
	AllowBankTransfer     bool     `json:"allow_bank_transfer"`
	AllowedPaymentMethods []string `json:"allowed_payment_methods"`
}

type cardCaptureRequest struct {
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
	SendEmail  bool   `json:"send_email"`
}

func (h *SubscriptionHandler) parseCheckoutRequest(c *gin.Context) (uuid.UUID, service.SubscriptionCheckoutInput, bool) {
	var req subscriptionCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return uuid.Nil, service.SubscriptionCheckoutInput{}, false
	}
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid customer_id"))
		return uuid.Nil, service.SubscriptionCheckoutInput{}, false
	}
	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		dto.RespondError(c, dto.NewBindError("invalid plan_id"))
		return uuid.Nil, service.SubscriptionCheckoutInput{}, false
	}
	return customerID, service.SubscriptionCheckoutInput{
		CustomerID:            customerID,
		PlanID:                planID,
		SuccessURL:            req.SuccessURL,
		CancelURL:             req.CancelURL,
		SendCheckoutEmail:     req.SendCheckoutEmail,
		CardOnly:              req.CardOnly,
		AllowBankTransfer:     req.AllowBankTransfer,
		AllowedPaymentMethods: req.AllowedPaymentMethods,
	}, true
}

func (h *SubscriptionHandler) Checkout(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	_, input, ok := h.parseCheckoutRequest(c)
	if !ok {
		return
	}
	result, err := h.checkout.StartCheckout(c.Request.Context(), tenant.ID, input)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondCreated(c, result)
}

func (h *SubscriptionHandler) ResumeCheckout(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	subscriptionID, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	_, input, ok := h.parseCheckoutRequest(c)
	if !ok {
		return
	}
	result, err := h.checkout.ResumeCheckout(c.Request.Context(), tenant.ID, subscriptionID, input)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, result)
}

func (h *SubscriptionHandler) CapturePaymentMethod(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	subscriptionID, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	var req cardCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}
	result, err := h.checkout.StartCardCapture(c.Request.Context(), tenant.ID, subscriptionID, service.CardCaptureInput{
		SuccessURL: req.SuccessURL,
		CancelURL:  req.CancelURL,
		SendEmail:  req.SendEmail,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, result)
}
