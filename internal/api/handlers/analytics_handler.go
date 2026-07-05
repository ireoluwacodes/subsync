package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) MRR(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var params dto.AnalyticsDateRangeParams
	_ = c.ShouldBindQuery(&params)

	result, err := h.svc.MRR(c.Request.Context(), tenant.ID, params.Currency)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.AnalyticsMRRToResponse(result))
}

func (h *AnalyticsHandler) Churn(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var params dto.AnalyticsDateRangeParams
	if err := c.ShouldBindQuery(&params); err != nil {
		dto.RespondError(c, err)
		return
	}
	defFrom, defTo := h.svc.DefaultRange(nil, nil)
	from, to, err := params.ParseRange(defFrom, defTo)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	result, err := h.svc.Churn(c.Request.Context(), tenant.ID, from, to)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.AnalyticsChurnToResponse(result))
}

func (h *AnalyticsHandler) Dunning(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var params dto.AnalyticsDateRangeParams
	if err := c.ShouldBindQuery(&params); err != nil {
		dto.RespondError(c, err)
		return
	}
	defFrom, defTo := h.svc.DefaultRange(nil, nil)
	from, to, err := params.ParseRange(defFrom, defTo)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	result, err := h.svc.Dunning(c.Request.Context(), tenant.ID, from, to)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.AnalyticsDunningToResponse(result))
}

func (h *AnalyticsHandler) Revenue(c *gin.Context) {
	tenant, ok := middlewareTenant(c)
	if !ok {
		return
	}
	var params dto.AnalyticsDateRangeParams
	if err := c.ShouldBindQuery(&params); err != nil {
		dto.RespondError(c, err)
		return
	}
	defFrom, defTo := h.svc.DefaultRange(nil, nil)
	from, to, err := params.ParseRange(defFrom, defTo)
	if err != nil {
		dto.RespondError(c, err)
		return
	}

	result, err := h.svc.Revenue(c.Request.Context(), tenant.ID, from, to, params.Currency)
	if err != nil {
		dto.RespondError(c, err)
		return
	}
	dto.RespondOK(c, dto.AnalyticsRevenueToResponse(result))
}
