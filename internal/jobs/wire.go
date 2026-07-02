package jobs

import (
	"context"

	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/crypto"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/service"
	"github.com/ireoluwacodes/subsync/internal/storage"
	"go.uber.org/zap"
)

func NewHandlers(ctx context.Context, cfg *config.Config, log *zap.Logger, database *db.DB, q *queue.Queue) (*Handlers, error) {
	key, err := crypto.ParseKey(cfg.DevEncryptionKey())
	if err != nil {
		return nil, err
	}
	enc, err := crypto.NewCredentialEncryptor(key)
	if err != nil {
		return nil, err
	}

	repos := db.NewRepos(database, enc)
	nombaClient := nomba.NewClient(log, nil)
	clk := clock.RealClock{}
	mailer := email.NewMailerService(cfg)
	store := storage.NewStorageService(cfg)

	invoices := service.NewInvoiceService(repos.Invoices, cfg, nombaClient, clk)
	subs := service.NewSubscriptionService(repos.Subscriptions, repos.Plans, repos.Customers, invoices)
	billing := service.NewBillingService(cfg, clk, repos, invoices, subs, mailer, q.Client)
	dunning := service.NewDunningService(clk, repos, invoices, subs, nombaClient, mailer, q.Client)

	return &Handlers{
		Config:   cfg,
		Clock:    clk,
		Billing:  billing,
		Dunning:  dunning,
		Invoices: invoices,
		Subs:     subs,
		Tenants:  service.NewTenantService(repos.Tenants, nombaClient),
		Nomba:    nombaClient,
		Email:    mailer,
		Storage:  store,
		Repos:    repos,
		Queue:    q,
	}, nil
}
