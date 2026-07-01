package handlers

import "github.com/gin-gonic/gin"

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler { return &WebhookHandler{} }

func (h *WebhookHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/webhook-endpoints", NotImplemented)
	rg.GET("/webhook-endpoints", NotImplemented)
	rg.GET("/webhook-endpoints/:id", NotImplemented)
	rg.PUT("/webhook-endpoints/:id", NotImplemented)
	rg.DELETE("/webhook-endpoints/:id", NotImplemented)
}
