package service

import (
	"github.com/ireoluwacodes/subsync/internal/auth"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/queue"
)

type Services struct {
	Tenants        *TenantService
	Auth           *AuthService
	Settings       *SettingsService
	Plans          *PlanService
	Customers      *CustomerService
	PaymentMethods *PaymentMethodService
	Subscriptions  *SubscriptionService
	Billing        *BillingService
	Dunning        *DunningService
	Invoices       *InvoiceService
	Webhooks       *WebhookService
	Portal         *PortalService
	Analytics      *AnalyticsService
}

func NewServices(repos *db.Repos, cfg *config.Config, nombaClient *nomba.Client, jwt *auth.JWTService, q *queue.Queue) *Services {
	clk := clock.RealClock{}
	mailer := email.NewMailerService(cfg)
	invoices := NewInvoiceService(repos.Invoices, cfg, nombaClient, clk)
	subs := NewSubscriptionService(repos.Subscriptions, repos.Plans, repos.Customers, invoices)

	var publisher TaskPublisher
	if q != nil {
		publisher = q.Client
	}

	tenants := NewTenantService(repos.Tenants, nombaClient)
	return &Services{
		Tenants:        tenants,
		Auth:           NewAuthService(repos.Users, repos.Tenants, repos.PasswordResets, jwt, nombaClient, tenants, cfg.PublicBaseURL, mailer, cfg),
		Settings:       NewSettingsService(repos.Tenants, nombaClient, cfg.PublicBaseURL),
		Plans:          NewPlanService(repos.Plans),
		Customers:      NewCustomerService(repos.Customers),
		PaymentMethods: NewPaymentMethodService(repos.PaymentMethods, repos.Customers),
		Invoices:       invoices,
		Subscriptions:  subs,
		Billing:        NewBillingService(cfg, clk, repos, invoices, subs, mailer, publisher),
		Dunning:        NewDunningService(clk, repos, invoices, subs, nombaClient, mailer, publisher),
		Webhooks:       NewWebhookService(repos.Webhooks),
		Portal:         NewPortalService(),
		Analytics:      NewAnalyticsService(),
	}
}
