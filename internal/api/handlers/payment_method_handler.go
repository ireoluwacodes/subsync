package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type PaymentMethodHandler struct {
	svc *service.PaymentMethodService
}

func NewPaymentMethodHandler(svc *service.PaymentMethodService) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: svc}
}

func (h *PaymentMethodHandler) Create(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	var req dto.CreatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	input, err := req.ToInput()
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	pm, err := h.svc.Create(c.Request.Context(), tenant.ID, input)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondCreated(c, dto.PaymentMethodToResponse(pm))
}

func (h *PaymentMethodHandler) Get(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	pm, err := h.svc.Get(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondOK(c, dto.PaymentMethodToResponse(pm))
}

func (h *PaymentMethodHandler) Delete(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), tenant.ID, id); err != nil {
		dto.RespondError(c, err)
		return
	}

	c.Status(204)
}

func (h *PaymentMethodHandler) SetDefault(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	pm, err := h.svc.SetDefault(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondOK(c, dto.PaymentMethodToResponse(pm))
}
