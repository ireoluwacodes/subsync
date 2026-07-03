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
	NombaEvents    *NombaEventService
	Analytics      *AnalyticsService
}

func NewServices(repos *db.Repos, cfg *config.Config, nombaClient *nomba.Client, jwt *auth.JWTService, q *queue.Queue) *Services {
	clk := clock.RealClock{}
	mailer := email.NewMailerService(cfg)
	invoices := NewInvoiceService(repos.Invoices, cfg, nombaClient, clk, nil)

	var publisher TaskPublisher
	if q != nil {
		publisher = q.Client
	}

	webhooks := NewWebhookService(repos.Webhooks, repos.Tenants, publisher, cfg)
	invoices.SetWebhooks(webhooks)
	subs := NewSubscriptionService(repos.Subscriptions, repos.Plans, repos.Customers, invoices, webhooks)
	paymentMethods := NewPaymentMethodService(repos.PaymentMethods, repos.Customers, webhooks)

	tenants := NewTenantService(repos.Tenants, nombaClient)
	billing := NewBillingService(cfg, clk, repos, invoices, subs, mailer, publisher, webhooks)
	portal := NewPortalService(clk, repos, subs, paymentMethods, nombaClient, cfg, cfg.PublicBaseURL, webhooks)
	nombaEvents := NewNombaEventService(clk, repos, billing, webhooks, portal)

	return &Services{
		Tenants:        tenants,
		Auth:           NewAuthService(repos.Users, repos.Tenants, repos.PasswordResets, jwt, nombaClient, tenants, cfg.PublicBaseURL, mailer, cfg),
		Settings:       NewSettingsService(repos.Tenants, nombaClient, cfg.PublicBaseURL),
		Plans:          NewPlanService(repos.Plans),
		Customers:      NewCustomerService(repos.Customers),
		PaymentMethods: paymentMethods,
		Invoices:       invoices,
		Subscriptions:  subs,
		Billing:        billing,
		Dunning:        NewDunningService(clk, repos, invoices, subs, nombaClient, mailer, publisher),
		Webhooks:       webhooks,
		Portal:         portal,
		NombaEvents:    nombaEvents,
		Analytics:      NewAnalyticsService(),
	}
}
