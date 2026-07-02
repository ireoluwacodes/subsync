package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterHealthRoutes(r *gin.Engine, h *handlers.HealthHandler) {
	r.GET("/health", h.Liveness)
	r.GET("/ready", h.Readiness)
}
