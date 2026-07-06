package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type PlanHandler struct {
	svc     *service.PlanService
	subSvc  *service.SubscriptionService
}

func NewPlanHandler(svc *service.PlanService, subSvc *service.SubscriptionService) *PlanHandler {
	return &PlanHandler{svc: svc, subSvc: subSvc}
}

func (h *PlanHandler) Create(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	var req dto.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	plan, err := h.svc.Create(c.Request.Context(), tenant.ID, req.ToInput())
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondCreated(c, dto.PlanToResponse(plan))
}

func (h *PlanHandler) List(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	var pagination dto.PaginationParams
	_ = c.ShouldBindQuery(&pagination)
	pagination.Normalize()

	activeOnly := false
	if v := c.Query("is_active"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			dto.RespondError(c, dto.NewBindError("invalid is_active query param"))
			return
		}
		activeOnly = parsed
	}

	offset := (pagination.Page - 1) * pagination.PerPage
	plans, total, err := h.svc.List(c.Request.Context(), tenant.ID, activeOnly, pagination.PerPage, offset)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	c.JSON(200, dto.Envelope{
		Data: dto.PlansToResponse(plans),
		Meta: dto.Meta{
			RequestID: c.GetString("request_id"),
			Page:      pagination.Page,
			PerPage:   pagination.PerPage,
			Total:     total,
		},
	})
}

func (h *PlanHandler) Get(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	plan, err := h.svc.Get(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondOK(c, dto.PlanToResponse(plan))
}

func (h *PlanHandler) Stats(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}
	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	plan, err := h.svc.Get(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	count, err := h.subSvc.PlanStats(c.Request.Context(), tenant.ID, id)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.PlanStatsToResponse(count, plan.Amount*count, plan.Currency))
}

func (h *PlanHandler) Update(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	var req dto.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondError(c, dto.NewBindError("invalid request body"))
		return
	}

	plan, err := h.svc.Update(c.Request.Context(), tenant.ID, id, req.ToInput())
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	dto.RespondOK(c, dto.PlanToResponse(plan))
}

func (h *PlanHandler) Archive(c *gin.Context) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		return
	}

	id, err := dto.IDParam(c, "id")
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	if err := h.svc.Archive(c.Request.Context(), tenant.ID, id); err != nil {
		dto.RespondError(c, err)
		return
	}

	c.Status(204)
}
