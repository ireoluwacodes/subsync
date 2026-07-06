package jobs

import (
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/service"
	"github.com/ireoluwacodes/subsync/internal/storage"
)

type Handlers struct {
	Config   *config.Config
	Clock    clock.Clock
	Billing  *service.BillingService
	Dunning  *service.DunningService
	Invoices *service.InvoiceService
	Subs     *service.SubscriptionService
	Tenants  *service.TenantService
	Nomba    *nomba.Client
	Email    *email.MailerService
	Mandates *service.MandateService
	Storage  *storage.StorageService
	Repos    *db.Repos
	Queue    *queue.Queue
}
