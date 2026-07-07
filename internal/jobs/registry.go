package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/ireoluwacodes/subsync/internal/observability"
)

type Registry struct {
	mux      *asynq.ServeMux
	handlers *Handlers
}

func NewRegistry(h *Handlers) *Registry {
	mux := asynq.NewServeMux()
	mux.Use(sentryMiddleware)
	return &Registry{mux: mux, handlers: h}
}

// sentryMiddleware recovers panics and forwards job failures to Sentry. Returned
// errors are only reported on the final attempt (once asynq retries are exhausted)
// to avoid flooding Sentry with transient, auto-retried failures.
func sentryMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("panic in task %s: %v", t.Type(), rec)
				observability.CaptureJobError(t.Type(), err, nil)
				zap.L().Error("panic recovered in task", zap.String("task", t.Type()), zap.Any("panic", rec))
			}
		}()

		err = next.ProcessTask(ctx, t)
		if err != nil {
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, ok := asynq.GetMaxRetry(ctx)
			if !ok || retried >= maxRetry {
				observability.CaptureJobError(t.Type(), err, map[string]any{
					"retry_count": retried,
					"max_retry":   maxRetry,
				})
			}
		}
		return err
	})
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
	r.mux.HandleFunc(TaskBillingReconcile, r.handleBillingReconcileProcessing)
	r.mux.HandleFunc(TaskPaymentMethodReminders, r.handlePaymentMethodReminders)
	r.mux.HandleFunc(TaskMandatePollStatus, r.handleMandatePollStatus)
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

func (r *Registry) handlePaymentMethodReminders(ctx context.Context, t *asynq.Task) error {
	n, err := r.handlers.Billing.ProcessPaymentMethodReminders(ctx, 50)
	if err != nil {
		return err
	}
	zap.L().Info("billing:payment_method_reminders complete", zap.Int("sent", n))
	return nil
}

func (r *Registry) handleMandatePollStatus(ctx context.Context, t *asynq.Task) error {
	if r.handlers.Mandates == nil {
		return nil
	}
	n, err := r.handlers.Mandates.ProcessPendingMandates(ctx, 50)
	if err != nil {
		return err
	}
	zap.L().Info("mandate:poll_status complete", zap.Int("processed", n))
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
