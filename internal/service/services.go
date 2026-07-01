package service

import (
	"github.com/ireoluwacodes/subsync/internal/auth"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/nomba"
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

func NewServices(repos *db.Repos, cfg *config.Config, nombaClient *nomba.Client, jwt *auth.JWTService) *Services {
	tenants := NewTenantService(repos.Tenants, nombaClient)
	invoices := NewInvoiceService(repos.Invoices, cfg)
	return &Services{
		Tenants:        tenants,
		Auth:           NewAuthService(repos.Users, repos.Tenants, repos.PasswordResets, jwt, nombaClient, tenants, cfg.PublicBaseURL),
		Settings:       NewSettingsService(repos.Tenants, nombaClient, cfg.PublicBaseURL),
		Plans:          NewPlanService(repos.Plans),
		Customers:      NewCustomerService(repos.Customers),
		PaymentMethods: NewPaymentMethodService(repos.PaymentMethods, repos.Customers),
		Invoices:       invoices,
		Subscriptions:  NewSubscriptionService(repos.Subscriptions, repos.Plans, repos.Customers, invoices),
		Billing:        NewBillingService(repos.Invoices),
		Dunning:        NewDunningService(),
		Webhooks:       NewWebhookService(repos.Webhooks),
		Portal:         NewPortalService(),
		Analytics:      NewAnalyticsService(),
	}
}
