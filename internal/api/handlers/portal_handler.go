package handlers

import "github.com/gin-gonic/gin"

type PortalHandler struct{}

func NewPortalHandler() *PortalHandler { return &PortalHandler{} }

func (h *PortalHandler) Register(r *gin.Engine) {
	r.GET("/portal/:token", NotImplemented)
	r.POST("/portal/:token/cancel", NotImplemented)
	r.POST("/portal/:token/update-payment-method", NotImplemented)
}

func (h *PortalHandler) RegisterAPI(rg *gin.RouterGroup) {
	rg.POST("/portal/token", NotImplemented)
}
