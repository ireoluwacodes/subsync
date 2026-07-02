package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterPlanRoutes(rg *gin.RouterGroup, h *handlers.PlanHandler) {
	rg.POST("/plans", h.Create)
	rg.GET("/plans", h.List)
	rg.GET("/plans/:id", h.Get)
	rg.GET("/plans/:id/stats", h.Stats)
	rg.PUT("/plans/:id", h.Update)
	rg.DELETE("/plans/:id", h.Archive)
}
