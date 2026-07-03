package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type Registry struct {
	mux      *asynq.ServeMux
	handlers *Handlers
}

func NewRegistry(h *Handlers) *Registry {
	return &Registry{mux: asynq.NewServeMux(), handlers: h}
}

func (r *Registry) Mux() *asynq.ServeMux {
	return r.mux
}

func (r *Registry) RegisterAll() {
	r.mux.HandleFunc(TaskBillingChargeDue, r.handleBillingChargeDue)
	r.mux.HandleFunc(TaskDunningStep, r.handleDunningStep)
	r.mux.HandleFunc(TaskTrialConvert, r.handleTrialConvert)
	r.mux.HandleFunc(TaskSubscriptionExpire, r.handleSubscriptionExpire)
	r.mux.HandleFunc(TaskSubscriptionResume, r.handleSubscriptionResume)
	r.mux.HandleFunc(TaskWebhookDeliver, r.handleWebhookDeliverJob)
	r.mux.HandleFunc(TaskInvoicePDF, r.handleInvoicePDF)
}

type taskPayload struct {
	TenantID       string `json:"tenant_id"`
	SubscriptionID string `json:"subscription_id"`
	InvoiceID      string `json:"invoice_id"`
}

func (r *Registry) handleBillingChargeDue(ctx context.Context, t *asynq.Task) error {
	n, err := r.handlers.Billing.ProcessDueSubscriptions(ctx, 50)
	if err != nil {
		return err
	}
	zap.L().Info("billing:charge_due complete", zap.Int("processed", n))
	return nil
}

func (r *Registry) handleDunningStep(ctx context.Context, t *asynq.Task) error {
	var p taskPayload
	if len(t.Payload()) > 0 {
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}
	}
	tenantID, subID, err := parseIDs(p.TenantID, p.SubscriptionID)
	if err != nil {
		return err
	}
	return r.handlers.Dunning.ProcessStep(ctx, tenantID, subID)
}

func (r *Registry) handleTrialConvert(ctx context.Context, t *asynq.Task) error {
	n, err := r.handlers.Subs.ConvertTrialsEnding(ctx, r.handlers.Clock.Now().UTC(), 50)
	if err != nil {
		return err
	}
	zap.L().Info("trial:convert complete", zap.Int("processed", n))
	return nil
}

func (r *Registry) handleSubscriptionExpire(ctx context.Context, t *asynq.Task) error {
	n, err := r.handlers.Subs.ExpireCancelAtPeriodEnd(ctx, r.handlers.Clock.Now().UTC(), 50)
	if err != nil {
		return err
	}
	zap.L().Info("subscription:expire complete", zap.Int("processed", n))
	return nil
}

func (r *Registry) handleSubscriptionResume(ctx context.Context, t *asynq.Task) error {
	n, err := r.handlers.Subs.ResumePausedEnding(ctx, r.handlers.Clock.Now().UTC(), 50)
	if err != nil {
		return err
	}
	zap.L().Info("subscription:resume complete", zap.Int("processed", n))
	return nil
}

func (r *Registry) handleInvoicePDF(ctx context.Context, t *asynq.Task) error {
	var p taskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	tenantID, invoiceID, err := parseIDs(p.TenantID, p.InvoiceID)
	if err != nil {
		return err
	}
	tenant, err := r.handlers.Repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	pdfBytes, err := r.handlers.Invoices.RenderPDF(ctx, tenantID, invoiceID, tenant)
	if err != nil {
		return err
	}
	key := "invoices/" + tenantID.String() + "/" + invoiceID.String() + ".pdf"
	url, err := r.handlers.Storage.Upload(ctx, key, pdfBytes, "application/pdf")
	if err != nil {
		return err
	}
	if url != "" {
		return r.handlers.Invoices.SetPDFURL(ctx, tenantID, invoiceID, url)
	}
	return nil
}

func (r *Registry) handleWebhookDeliverJob(ctx context.Context, t *asynq.Task) error {
	return r.handlers.handleWebhookDeliver(ctx, t.Payload())
}

func noopHandler(name string) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		zap.L().Info("job stub executed", zap.String("task", name))
		return nil
	}
}
