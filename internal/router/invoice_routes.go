package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterInvoiceRoutes(rg *gin.RouterGroup, h *handlers.InvoiceHandler) {
	rg.GET("/invoices", h.List)
	rg.GET("/invoices/:id", h.Get)
	rg.GET("/invoices/:id/pdf", h.PDF)
	rg.POST("/invoices/:id/void", h.Void)
	rg.POST("/invoices/:id/retry", h.Retry)
}
