package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterAuthPublicRoutes(api *gin.RouterGroup, h *handlers.AuthHandler) {
	g := api.Group("/auth")
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.POST("/forgot-password", h.ForgotPassword)
	g.POST("/reset-password", h.ResetPassword)
}

func RegisterAuthProtectedRoutes(rg *gin.RouterGroup, h *handlers.AuthHandler) {
	rg.POST("/auth/logout", h.Logout)
}
