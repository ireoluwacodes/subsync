package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterTenantPublicRoutes(api *gin.RouterGroup, h *handlers.TenantHandler) {
	api.POST("/tenants", h.Create)
}

func RegisterTenantProtectedRoutes(rg *gin.RouterGroup, h *handlers.TenantHandler) {
	rg.GET("/me", h.Me)
}
