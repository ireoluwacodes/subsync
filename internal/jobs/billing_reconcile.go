package jobs

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func (r *Registry) handleBillingReconcileProcessing(ctx context.Context, _ *asynq.Task) error {
	cutoff := r.handlers.Clock.Now().UTC().Add(-15 * time.Minute)
	invoices, err := r.handlers.Repos.Invoices.ListProcessingBefore(ctx, cutoff, 50)
	if err != nil {
		return err
	}
	reconciled := 0
	for _, inv := range invoices {
		tenant, err := r.handlers.Repos.Tenants.GetByID(ctx, inv.TenantID)
		if err != nil {
			continue
		}
		if err := r.handlers.Repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
			continue
		}
		if inv.NombaOrderRef == "" {
			continue
		}
		result, err := r.handlers.Nomba.VerifyCheckoutTransaction(ctx, tenant, inv.NombaOrderRef)
		if err != nil || !result.Status {
			continue
		}
		txID := result.Message
		if err := r.handlers.Billing.FinalizePaidInvoice(ctx, inv.TenantID, inv.ID, txID); err != nil {
			continue
		}
		reconciled++
	}
	zap.L().Info("billing:reconcile_processing complete", zap.Int("reconciled", reconciled))
	return nil
}
