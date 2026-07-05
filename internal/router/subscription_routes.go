package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterSubscriptionRoutes(rg *gin.RouterGroup, h *handlers.SubscriptionHandler) {
	rg.POST("/subscriptions/checkout", h.Checkout)
	rg.POST("/subscriptions", h.Create)
	rg.GET("/subscriptions", h.List)
	rg.GET("/subscriptions/:id", h.Get)
	rg.POST("/subscriptions/:id/checkout", h.ResumeCheckout)
	rg.POST("/subscriptions/:id/capture-payment-method", h.CapturePaymentMethod)
	rg.POST("/subscriptions/:id/cancel", h.Cancel)
	rg.POST("/subscriptions/:id/pause", h.Pause)
	rg.POST("/subscriptions/:id/resume", h.Resume)
	rg.POST("/subscriptions/:id/upgrade", h.Upgrade)
	rg.GET("/subscriptions/:id/upgrade/preview", h.PreviewUpgrade)
	rg.GET("/subscriptions/:id/transitions", h.Transitions)
}
