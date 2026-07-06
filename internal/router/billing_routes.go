package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterBillingReturnRoutes(r *gin.Engine, h *handlers.BillingReturnHandler) {
	r.GET("/billing/success", h.Success)
}
