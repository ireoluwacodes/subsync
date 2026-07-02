package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterWebhookRoutes(rg *gin.RouterGroup, h *handlers.WebhookHandler) {
	rg.POST("/webhook-endpoints", h.CreateEndpoint)
	rg.GET("/webhook-endpoints", h.ListEndpoints)
	rg.GET("/webhook-endpoints/:id", h.GetEndpoint)
	rg.PUT("/webhook-endpoints/:id", h.UpdateEndpoint)
	rg.DELETE("/webhook-endpoints/:id", h.DeleteEndpoint)
}
