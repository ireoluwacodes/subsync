package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type Registry struct {
	mux *asynq.ServeMux
}

func NewRegistry() *Registry {
	return &Registry{mux: asynq.NewServeMux()}
}

func (r *Registry) Mux() *asynq.ServeMux {
	return r.mux
}

func (r *Registry) RegisterAll() {
	r.mux.HandleFunc(TaskBillingChargeDue, handleBillingChargeDue)
	r.mux.HandleFunc(TaskDunningStep, handleDunningStep)
	r.mux.HandleFunc(TaskTrialConvert, handleTrialConvert)
	r.mux.HandleFunc(TaskSubscriptionExpire, handleSubscriptionExpire)
	r.mux.HandleFunc(TaskSubscriptionResume, handleSubscriptionResume)
	r.mux.HandleFunc(TaskWebhookDeliver, handleWebhookDeliver)
	r.mux.HandleFunc(TaskInvoicePDF, handleInvoicePDF)
}

func noopHandler(name string) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		zap.L().Info("job stub executed", zap.String("task", name))
		return nil
	}
}

func handleBillingChargeDue(ctx context.Context, t *asynq.Task) error {
	return noopHandler(TaskBillingChargeDue)(ctx, t)
}

func handleDunningStep(ctx context.Context, t *asynq.Task) error {
	return noopHandler(TaskDunningStep)(ctx, t)
}

func handleTrialConvert(ctx context.Context, t *asynq.Task) error {
	return noopHandler(TaskTrialConvert)(ctx, t)
}

func handleSubscriptionExpire(ctx context.Context, t *asynq.Task) error {
	return noopHandler(TaskSubscriptionExpire)(ctx, t)
}

func handleSubscriptionResume(ctx context.Context, t *asynq.Task) error {
	return noopHandler(TaskSubscriptionResume)(ctx, t)
}

func handleWebhookDeliver(ctx context.Context, t *asynq.Task) error {
	return noopHandler(TaskWebhookDeliver)(ctx, t)
}

func handleInvoicePDF(ctx context.Context, t *asynq.Task) error {
	return noopHandler(TaskInvoicePDF)(ctx, t)
}
