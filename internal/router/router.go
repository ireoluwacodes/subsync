package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
)

const V1Prefix = "/api/v1"

// SetupRouter configures the Gin engine with middleware and routes.
func SetupRouter(deps Dependencies) *gin.Engine {
	if deps.Config.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.RedirectTrailingSlash = false

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(deps.Config.CORSAllowedOrigins))

	RegisterHealthRoutes(r, deps.HealthHandler)

	api := r.Group(V1Prefix)

	RegisterAuthPublicRoutes(api, deps.AuthHandler)

	protected := api.Group("", middleware.Auth(deps.Repos.Tenants, deps.AuthService, deps.Queue.Redis()), middleware.Tenant())

	RegisterAuthProtectedRoutes(protected, deps.AuthHandler)
	RegisterTenantProtectedRoutes(protected, deps.TenantHandler)
	RegisterSettingsRoutes(protected, deps.SettingsHandler)
	RegisterPlanRoutes(protected, deps.PlanHandler)
	RegisterCustomerRoutes(protected, deps.CustomerHandler)
	RegisterSubscriptionRoutes(protected, deps.SubscriptionHandler)
	RegisterInvoiceRoutes(protected, deps.InvoiceHandler)
	RegisterCustomerNestedRoutes(protected, deps.SubscriptionHandler, deps.InvoiceHandler)
	RegisterPaymentMethodRoutes(protected, deps.PaymentMethodHandler)
	RegisterWebhookRoutes(protected, deps.WebhookHandler)
	RegisterAnalyticsRoutes(protected, deps.AnalyticsHandler)
	RegisterPortalAPIRoutes(protected, deps.PortalHandler)

	RegisterNombaWebhookRoutes(r, deps.NombaWebhookHandler)
	RegisterPortalPublicRoutes(r, deps.PortalHandler)

	return r
}

func RegisterNombaWebhookRoutes(r *gin.Engine, h *handlers.NombaWebhookHandler) {
	r.POST("/webhooks/nomba/:tenant_id", h.Receive)
}
