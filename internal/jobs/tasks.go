package jobs

const (
	TaskBillingChargeDue      = "billing:charge_due"
	TaskDunningStep           = "dunning:step"
	TaskTrialConvert          = "trial:convert"
	TaskSubscriptionExpire    = "subscription:expire"
	TaskSubscriptionResume    = "subscription:resume"
	TaskWebhookDeliver        = "webhook:deliver"
	TaskInvoicePDF            = "invoice:pdf"
	TaskBillingReconcile         = "billing:reconcile_processing"
	TaskPaymentMethodReminders   = "billing:payment_method_reminders"
)
