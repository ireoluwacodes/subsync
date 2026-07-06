package jobs

import (
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func RegisterPeriodicTasks(scheduler *asynq.Scheduler) error {
	entries := []struct {
		cron string
		task string
	}{
		{"*/5 * * * *", TaskBillingChargeDue},
		{"*/15 * * * *", TaskBillingReconcile},
		{"*/15 * * * *", TaskMandatePollStatus},
		{"0 * * * *", TaskTrialConvert},
		{"0 * * * *", TaskPaymentMethodReminders},
		{"0 * * * *", TaskSubscriptionExpire},
		{"0 * * * *", TaskSubscriptionResume},
	}

	for _, e := range entries {
		task := asynq.NewTask(e.task, nil)
		if _, err := scheduler.Register(e.cron, task); err != nil {
			return err
		}
		zap.L().Info("registered periodic task",
			zap.String("task", e.task),
			zap.String("cron", e.cron),
		)
	}

	return nil
}
