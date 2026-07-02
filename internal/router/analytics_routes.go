package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterAnalyticsRoutes(rg *gin.RouterGroup, h *handlers.AnalyticsHandler) {
	rg.GET("/analytics/mrr", h.MRR)
	rg.GET("/analytics/churn", h.Churn)
	rg.GET("/analytics/dunning", h.Dunning)
	rg.GET("/analytics/revenue", h.Revenue)
}
