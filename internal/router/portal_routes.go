package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterPortalAPIRoutes(rg *gin.RouterGroup, h *handlers.PortalHandler) {
	rg.POST("/portal/token", h.CreateToken)
}

func RegisterPortalPublicRoutes(r *gin.Engine, h *handlers.PortalHandler) {
	r.GET("/portal/:token", h.Home)
	r.POST("/portal/:token/cancel", h.Cancel)
	r.POST("/portal/:token/update-payment-method", h.UpdatePaymentMethod)
}
