package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterTenantProtectedRoutes(rg *gin.RouterGroup, h *handlers.TenantHandler) {
	rg.GET("/me", h.Me)
}
