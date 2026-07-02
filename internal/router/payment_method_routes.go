package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterPaymentMethodRoutes(rg *gin.RouterGroup, h *handlers.PaymentMethodHandler) {
	rg.POST("/payment-methods", h.Create)
	rg.GET("/payment-methods/:id", h.Get)
	rg.DELETE("/payment-methods/:id", h.Delete)
	rg.POST("/payment-methods/:id/set-default", h.SetDefault)
}
