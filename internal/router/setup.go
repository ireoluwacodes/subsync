package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/portalpage"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/service"
)

// Setup wires handlers and returns a configured Gin engine.
func Setup(cfg *config.Config, database *db.DB, q *queue.Queue, repos *db.Repos, svcs *service.Services) *gin.Engine {
	portalRenderer, err := portalpage.NewRenderer()
	if err != nil {
		panic("portal renderer: " + err.Error())
	}

	deps := Dependencies{
		Config:               cfg,
		Repos:                repos,
		Queue:                q,
		AuthService:          svcs.Auth,
		Services:             svcs,
		HealthHandler:        handlers.NewHealthHandler(database, q),
		AuthHandler:          handlers.NewAuthHandler(cfg, svcs.Auth),
		TenantHandler:        handlers.NewTenantHandler(svcs.Tenants),
		SettingsHandler:      handlers.NewSettingsHandler(svcs.Settings),
		PlanHandler:          handlers.NewPlanHandler(svcs.Plans, svcs.Subscriptions),
		CustomerHandler:      handlers.NewCustomerHandler(svcs.Customers),
		SubscriptionHandler:  handlers.NewSubscriptionHandler(svcs.Subscriptions, svcs.Checkout),
		InvoiceHandler:       handlers.NewInvoiceHandler(svcs.Invoices),
		PaymentMethodHandler: handlers.NewPaymentMethodHandler(svcs.PaymentMethods),
		WebhookHandler:       handlers.NewWebhookHandler(svcs.Webhooks),
		PortalHandler:        handlers.NewPortalHandler(svcs.Portal, portalRenderer),
		BillingReturnHandler: handlers.NewBillingReturnHandler(svcs.BillingReturn, portalRenderer),
		AnalyticsHandler:     handlers.NewAnalyticsHandler(svcs.Analytics),
		NombaWebhookHandler:  handlers.NewNombaWebhookHandler(cfg, svcs.Tenants, svcs.NombaEvents),
	}

	return SetupRouter(deps)
}
