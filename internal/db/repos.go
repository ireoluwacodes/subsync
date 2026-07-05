package db

import (
	"github.com/ireoluwacodes/subsync/internal/crypto"
)

type Repos struct {
	Tenants        *TenantRepo
	Users          *UserRepo
	PasswordResets *PasswordResetRepo
	Plans          *PlanRepo
	Customers      *CustomerRepo
	PaymentMethods *PaymentMethodRepo
	Subscriptions  *SubscriptionRepo
	Invoices       *InvoiceRepo
	Webhooks       *WebhookRepo
	NombaEvents    *NombaEventRepo
	PortalTokens   *PortalTokenRepo
	Analytics      *AnalyticsRepo
}

func NewRepos(database *DB, enc *crypto.CredentialEncryptor) *Repos {
	return &Repos{
		Tenants:        NewTenantRepo(database, enc),
		Users:          NewUserRepo(database),
		PasswordResets: NewPasswordResetRepo(database),
		Plans:          NewPlanRepo(database),
		Customers:      NewCustomerRepo(database),
		PaymentMethods: NewPaymentMethodRepo(database),
		Subscriptions:  NewSubscriptionRepo(database),
		Invoices:       NewInvoiceRepo(database),
		Webhooks:       NewWebhookRepo(database),
		NombaEvents:    NewNombaEventRepo(database),
		PortalTokens:   NewPortalTokenRepo(database),
		Analytics:      NewAnalyticsRepo(database),
	}
}
