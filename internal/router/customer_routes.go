package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
)

func RegisterCustomerRoutes(rg *gin.RouterGroup, h *handlers.CustomerHandler) {
	rg.POST("/customers", h.Create)
	rg.GET("/customers", h.List)
	rg.GET("/customers/:id", h.Get)
	rg.PUT("/customers/:id", h.Update)
}

func RegisterCustomerNestedRoutes(rg *gin.RouterGroup, sub *handlers.SubscriptionHandler, inv *handlers.InvoiceHandler) {
	rg.GET("/customers/:id/subscriptions", sub.ListForCustomer)
	rg.GET("/customers/:id/invoices", inv.ListForCustomer)
}
