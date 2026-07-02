package router

import (
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type Dependencies struct {
	Config      *config.Config
	Repos       *db.Repos
	Queue       *queue.Queue
	AuthService *service.AuthService

	HealthHandler        *handlers.HealthHandler
	AuthHandler          *handlers.AuthHandler
	TenantHandler        *handlers.TenantHandler
	SettingsHandler      *handlers.SettingsHandler
	PlanHandler          *handlers.PlanHandler
	CustomerHandler      *handlers.CustomerHandler
	SubscriptionHandler  *handlers.SubscriptionHandler
	InvoiceHandler       *handlers.InvoiceHandler
	PaymentMethodHandler *handlers.PaymentMethodHandler
	WebhookHandler       *handlers.WebhookHandler
	PortalHandler        *handlers.PortalHandler
	AnalyticsHandler     *handlers.AnalyticsHandler
}
