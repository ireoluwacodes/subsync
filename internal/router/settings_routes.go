package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterSettingsRoutes(rg *gin.RouterGroup, h *handlers.SettingsHandler) {
	rg.GET("/settings", h.Get)
	rg.PATCH("/settings/general", h.UpdateGeneral)
	rg.PATCH("/settings/dunning", h.UpdateDunning)
	rg.PATCH("/settings/branding", h.UpdateBranding)
	rg.PATCH("/settings/billing-email", h.UpdateBillingEmail)
	rg.GET("/settings/nomba", h.GetNomba)
	rg.PATCH("/settings/nomba", h.UpdateNomba)
	rg.POST("/settings/api-key/rotate", h.RotateAPIKey)
}
