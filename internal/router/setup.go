package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/service"
)

// Setup wires handlers and returns a configured Gin engine.
func Setup(cfg *config.Config, database *db.DB, q *queue.Queue, repos *db.Repos, svcs *service.Services) *gin.Engine {
	deps := Dependencies{
		Config:               cfg,
		Repos:                repos,
		Queue:                q,
		AuthService:          svcs.Auth,
		HealthHandler:        handlers.NewHealthHandler(database, q),
		AuthHandler:          handlers.NewAuthHandler(svcs.Auth),
		TenantHandler:        handlers.NewTenantHandler(svcs.Tenants, cfg),
		SettingsHandler:      handlers.NewSettingsHandler(svcs.Settings),
		PlanHandler:          handlers.NewPlanHandler(svcs.Plans, svcs.Subscriptions),
		CustomerHandler:      handlers.NewCustomerHandler(svcs.Customers),
		SubscriptionHandler:  handlers.NewSubscriptionHandler(svcs.Subscriptions),
		InvoiceHandler:       handlers.NewInvoiceHandler(svcs.Invoices),
		PaymentMethodHandler: handlers.NewPaymentMethodHandler(svcs.PaymentMethods),
		WebhookHandler:       handlers.NewWebhookHandler(),
		PortalHandler:        handlers.NewPortalHandler(),
		AnalyticsHandler:     handlers.NewAnalyticsHandler(),
	}

	return SetupRouter(deps)
}
