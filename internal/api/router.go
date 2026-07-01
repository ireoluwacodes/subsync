package api

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/handlers"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type RouterDeps struct {
	Config *config.Config
	DB     *db.DB
	Queue  *queue.Queue
	Repos  *db.Repos
	Svcs   *service.Services
	Nomba  *nomba.Client
}

func NewRouter(deps RouterDeps) *gin.Engine {
	if gin.Mode() == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	health := handlers.NewHealthHandler(deps.DB, deps.Queue)
	r.GET("/health", health.Liveness)
	r.GET("/ready", health.Readiness)

	tenantHandler := handlers.NewTenantHandler(deps.Svcs.Tenants, deps.Config)
	tenantHandler.RegisterPublic(r)

	authHandler := handlers.NewAuthHandler(deps.Svcs.Auth)
	authHandler.RegisterPublic(r)

	var rdb = deps.Queue.Redis()
	v1 := r.Group("/v1", middleware.Auth(deps.Repos.Tenants, deps.Svcs.Auth, rdb), middleware.Tenant())

	authHandler.RegisterAuthenticated(v1)
	tenantHandler.RegisterAuthenticated(v1)
	handlers.NewSettingsHandler(deps.Svcs.Settings).Register(v1)
	handlers.NewPlanHandler(deps.Svcs.Plans, deps.Svcs.Subscriptions).Register(v1)
	handlers.NewCustomerHandler(deps.Svcs.Customers).Register(v1)

	subHandler := handlers.NewSubscriptionHandler(deps.Svcs.Subscriptions)
	invHandler := handlers.NewInvoiceHandler(deps.Svcs.Invoices)
	subHandler.Register(v1)
	invHandler.Register(v1)

	v1.GET("/customers/:id/subscriptions", subHandler.ListForCustomer)
	v1.GET("/customers/:id/invoices", invHandler.ListForCustomer)

	handlers.NewPaymentMethodHandler(deps.Svcs.PaymentMethods).Register(v1)
	handlers.NewWebhookHandler().Register(v1)
	handlers.NewPortalHandler().RegisterAPI(v1)
	handlers.NewAnalyticsHandler().Register(v1)

	handlers.NewPortalHandler().Register(r)

	return r
}
