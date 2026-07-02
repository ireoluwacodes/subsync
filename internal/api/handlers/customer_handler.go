package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) Create(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	customer, err := h.svc.Create(c.Request.Context(), tenant.ID, req.ToInput())
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondCreated(c, dto.CustomerToResponse(customer))
}

func (h *CustomerHandler) List(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	var pagination dto.PaginationParams
	_ = c.ShouldBindQuery(&pagination)
	pagination.Normalize()

	offset := (pagination.Page - 1) * pagination.PerPage
	customers, total, err := h.svc.List(c.Request.Context(), tenant.ID, pagination.PerPage, offset)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	c.JSON(200, dto.Envelope{
		Data: dto.CustomersToResponse(customers),
		Meta: dto.Meta{
			RequestID: c.GetString("request_id"),
			Page:      pagination.Page,
			PerPage:   pagination.PerPage,
			Total:     total,
		},
	})
}

func (h *CustomerHandler) Get(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	customer, err := h.svc.Get(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondOK(c, dto.CustomerToResponse(customer))
}

func (h *CustomerHandler) Update(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	customer, err := h.svc.Update(c.Request.Context(), tenant.ID, id, req.ToInput())
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondOK(c, dto.CustomerToResponse(customer))
}
