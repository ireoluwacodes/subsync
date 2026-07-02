package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type InvoiceHandler struct {
	svc *service.InvoiceService
}

func NewInvoiceHandler(svc *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

func (h *InvoiceHandler) List(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var pagination dto.PaginationParams
	_ = c.ShouldBindQuery(&pagination)
	pagination.Normalize()

	filter := domain.InvoiceListFilter{
		Status: c.Query("status"),
		Limit:  pagination.PerPage,
		Offset: (pagination.Page - 1) * pagination.PerPage,
	}
	if v := c.Query("customer_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.CustomerID = &id
		}
	}
	if v := c.Query("subscription_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.SubscriptionID = &id
		}
	}

	invoices, total, err := h.svc.List(c.Request.Context(), tenant.ID, filter)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	out := make([]dto.InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		out[i] = dto.InvoiceToResponse(inv)
	}
	c.JSON(200, dto.Envelope{
		Data: out,
		Meta: dto.Meta{RequestID: c.GetString("request_id"), Page: pagination.Page, PerPage: pagination.PerPage, Total: total},
	})
}

func (h *InvoiceHandler) Get(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	inv, items, err := h.svc.Get(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, gin.H{
		"invoice":    dto.InvoiceToResponse(inv),
		"line_items": items,
	})
}

func (h *InvoiceHandler) PDF(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	inv, _, err := h.svc.Get(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	if inv.Metadata != nil {
		if pdfURL, ok := inv.Metadata["pdf_url"].(string); ok && pdfURL != "" {
			c.Redirect(http.StatusFound, pdfURL)
			return
		}
	}
	data, err := h.svc.RenderPDF(c.Request.Context(), tenant.ID, id, tenant)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	c.Data(http.StatusOK, "application/pdf", data)
}

func (h *InvoiceHandler) Void(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	inv, err := h.svc.Void(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.InvoiceToResponse(inv))
}

func (h *InvoiceHandler) Retry(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	inv, err := h.svc.Charge(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.InvoiceToResponse(inv))
}

func (h *InvoiceHandler) ListForCustomer(c *gin.Context) {
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
	invoices, _, err := h.svc.List(c.Request.Context(), tenant.ID, domain.InvoiceListFilter{
		CustomerID: &customerID,
		Limit:      limit,
	})
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	out := make([]dto.InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		out[i] = dto.InvoiceToResponse(inv)
	}
	dto.RespondOK(c, out)
}
