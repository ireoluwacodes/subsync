package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

const (
	jobDunningStep = "dunning:step"
	jobInvoicePDF  = "invoice:pdf"
)

type TaskPublisher interface {
	EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type BillingService struct {
	cfg       *config.Config
	clock     clock.Clock
	repos     *db.Repos
	invoices  *InvoiceService
	subs      *SubscriptionService
	mailer    *email.MailerService
	publisher TaskPublisher
	webhooks  *WebhookService
}

func NewBillingService(
	cfg *config.Config,
	clk clock.Clock,
	repos *db.Repos,
	invoices *InvoiceService,
	subs *SubscriptionService,
	mailer *email.MailerService,
	publisher TaskPublisher,
	webhooks *WebhookService,
) *BillingService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &BillingService{
		cfg:       cfg,
		clock:     clk,
		repos:     repos,
		invoices:  invoices,
		subs:      subs,
		mailer:    mailer,
		publisher: publisher,
		webhooks:  webhooks,
	}
}

type jobPayload struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	SubscriptionID uuid.UUID `json:"subscription_id,omitempty"`
	InvoiceID      uuid.UUID `json:"invoice_id,omitempty"`
}

func (s *BillingService) ProcessDueSubscriptions(ctx context.Context, limit int) (int, error) {
	subs, err := s.repos.Subscriptions.ListDueForBilling(ctx, s.clock.Now().UTC(), limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, sub := range subs {
		if err := s.ChargeDueSubscription(ctx, sub.TenantID, sub.ID); err != nil {
			continue
		}
		processed++
	}
	return processed, nil
}

func (s *BillingService) ChargeDueSubscription(ctx context.Context, tenantID, subscriptionID uuid.UUID) error {
	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}
	plan, err := s.repos.Plans.GetByID(ctx, tenantID, sub.PlanID)
	if err != nil {
		return err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, sub.CustomerID)
	if err != nil {
		return err
	}

	inv, err := s.repos.Invoices.FindForBillingPeriod(ctx, tenantID, sub.ID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if err == domain.ErrNotFound {
		inv, err = s.invoices.CreateSubscriptionInvoice(ctx, tenantID, sub, plan)
		if err != nil {
			return err
		}
	}
	if inv.Status == domain.InvoiceStatusPaid {
		return s.advanceSubscriptionPeriod(ctx, sub, plan)
	}
	if inv.Status == domain.InvoiceStatusProcessing {
		return nil
	}

	pm, err := s.resolvePaymentMethod(ctx, tenantID, sub, customer.ID)
	if err != nil {
		return err
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return err
	}

	charged, chargeErr := s.invoices.ChargeWithPayment(ctx, tenant, pm, inv)
	if chargeErr != nil {
		return s.handleChargeFailure(ctx, tenant, customer, sub, charged, chargeErr)
	}

	if charged.Status == domain.InvoiceStatusProcessing {
		return nil
	}

	if err := s.handleChargeSuccess(ctx, tenant, customer, sub, plan, charged); err != nil {
		return err
	}
	return nil
}

func (s *BillingService) resolvePaymentMethod(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription, customerID uuid.UUID) (*domain.PaymentMethod, error) {
	if sub.PaymentMethodID != nil {
		return s.repos.PaymentMethods.GetByID(ctx, tenantID, *sub.PaymentMethodID)
	}
	return s.repos.PaymentMethods.GetDefaultForCustomer(ctx, tenantID, customerID)
}

func (s *BillingService) handleChargeSuccess(ctx context.Context, tenant *domain.Tenant, customer *domain.Customer, sub *domain.Subscription, plan *domain.Plan, inv *domain.Invoice) error {
	from := sub.State
	if sub.State == domain.SubscriptionStatePastDue {
		sub.State = domain.SubscriptionStateActive
	}
	sub.DunningStep = 0
	sub.DunningStartedAt = nil

	if err := s.advanceSubscriptionPeriod(ctx, sub, plan); err != nil {
		return err
	}

	if from != sub.State {
		_ = s.repos.Subscriptions.RecordTransition(ctx, &domain.SubscriptionTransition{
			SubscriptionID: sub.ID,
			TenantID:       sub.TenantID,
			FromState:      from,
			ToState:        sub.State,
			Reason:         "payment_succeeded",
			Actor:          "system",
		})
	}

	subject, html := email.SubscriptionConfirmedHTML(tenant.Name, inv.AmountPaid, inv.Currency)
	_ = s.mailer.Send(ctx, customer.Email, subject, html)

	s.enqueueInvoicePDF(ctx, tenant.ID, inv.ID)
	return nil
}

func (s *BillingService) handleChargeFailure(ctx context.Context, tenant *domain.Tenant, customer *domain.Customer, sub *domain.Subscription, inv *domain.Invoice, chargeErr error) error {
	now := s.clock.Now().UTC()
	if sub.State != domain.SubscriptionStatePastDue {
		from := sub.State
		sub.State = domain.SubscriptionStatePastDue
		_ = s.repos.Subscriptions.RecordTransition(ctx, &domain.SubscriptionTransition{
			SubscriptionID: sub.ID,
			TenantID:       sub.TenantID,
			FromState:      from,
			ToState:        sub.State,
			Reason:         "payment_failed",
			Actor:          "system",
		})
	}
	if sub.DunningStartedAt == nil {
		sub.DunningStartedAt = &now
		sub.DunningStep = 0
	}
	_ = s.repos.Subscriptions.Update(ctx, sub)

	amount := int64(0)
	currency := "NGN"
	if inv != nil {
		amount = inv.AmountDue
		currency = inv.Currency
	}
	subject, html := email.PaymentFailedHTML(tenant.Name, amount, currency)
	_ = s.mailer.Send(ctx, customer.Email, subject, html)

	s.enqueueDunningStep(ctx, tenant.ID, sub.ID, s.delayForDunningStep(tenant, 0))
	return chargeErr
}

func (s *BillingService) advanceSubscriptionPeriod(ctx context.Context, sub *domain.Subscription, plan *domain.Plan) error {
	start := sub.CurrentPeriodEnd
	end := utils.PlanPeriodEnd(start, plan)
	sub.CurrentPeriodStart = start
	sub.CurrentPeriodEnd = end
	sub.NextBillingAt = &end
	return s.repos.Subscriptions.Update(ctx, sub)
}

func (s *BillingService) enqueueDunningStep(ctx context.Context, tenantID, subscriptionID uuid.UUID, delay time.Duration) {
	if s.publisher == nil {
		return
	}
	raw, _ := json.Marshal(jobPayload{TenantID: tenantID, SubscriptionID: subscriptionID})
	task := asynq.NewTask(jobDunningStep, raw)
	_, _ = s.publisher.EnqueueContext(ctx, task, asynq.ProcessIn(delay))
}

func (s *BillingService) enqueueInvoicePDF(ctx context.Context, tenantID, invoiceID uuid.UUID) {
	if s.publisher == nil {
		return
	}
	raw, _ := json.Marshal(jobPayload{TenantID: tenantID, InvoiceID: invoiceID})
	task := asynq.NewTask(jobInvoicePDF, raw)
	_, _ = s.publisher.EnqueueContext(ctx, task)
}

func (s *BillingService) delayForDunningStep(tenant *domain.Tenant, stepIndex int) time.Duration {
	steps, err := utils.ParseDunningSteps(tenant.DunningConfig)
	if err != nil || stepIndex >= len(steps) {
		return 24 * time.Hour
	}
	return time.Duration(steps[stepIndex].DelayDays) * 24 * time.Hour
}

// EnqueueInitialDunning schedules the first dunning step using tenant config.
func (s *BillingService) EnqueueInitialDunning(ctx context.Context, tenant *domain.Tenant, subscriptionID uuid.UUID) {
	delay := s.delayForDunningStep(tenant, 0)
	s.enqueueDunningStep(ctx, tenant.ID, subscriptionID, delay)
}

// FinalizePaidInvoice marks an invoice paid and advances the subscription after webhook confirmation.
func (s *BillingService) FinalizePaidInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID, transactionID string) error {
	inv, err := s.repos.Invoices.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return err
	}
	if inv.Status == domain.InvoiceStatusPaid {
		return nil
	}
	if inv.Status != domain.InvoiceStatusOpen && inv.Status != domain.InvoiceStatusProcessing {
		return nil
	}

	now := s.clock.Now().UTC()
	inv.Status = domain.InvoiceStatusPaid
	inv.AmountPaid = inv.AmountDue
	inv.PaidAt = &now
	if transactionID != "" {
		inv.NombaTransactionID = transactionID
	}
	if err := s.repos.Invoices.Update(ctx, inv); err != nil {
		return err
	}

	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, inv.SubscriptionID)
	if err != nil {
		return err
	}
	plan, err := s.repos.Plans.GetByID(ctx, tenantID, sub.PlanID)
	if err != nil {
		return err
	}
	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, inv.CustomerID)
	if err != nil {
		return err
	}

	if err := s.handleChargeSuccess(ctx, tenant, customer, sub, plan, inv); err != nil {
		return err
	}

	if s.webhooks != nil {
		_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventInvoicePaid, map[string]any{
			"id":     inv.ID.String(),
			"status": string(inv.Status),
			"amount_paid": inv.AmountPaid,
			"currency": inv.Currency,
		})
	}
	return nil
}

// HandleWebhookPaymentFailure processes a Nomba payment_failed webhook for an invoice.
func (s *BillingService) HandleWebhookPaymentFailure(ctx context.Context, tenantID, invoiceID uuid.UUID) error {
	inv, err := s.repos.Invoices.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return err
	}
	if inv.Status == domain.InvoiceStatusPaid {
		return nil
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, inv.SubscriptionID)
	if err != nil {
		return err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, inv.CustomerID)
	if err != nil {
		return err
	}

	inv.Status = domain.InvoiceStatusOpen
	now := s.clock.Now().UTC()
	inv.NextAttemptAt = utils.PtrTime(now.Add(24 * time.Hour))
	_ = s.repos.Invoices.Update(ctx, inv)

	err = s.handleChargeFailure(ctx, tenant, customer, sub, inv, fmt.Errorf("%w: nomba payment failed", domain.ErrValidation))
	if s.webhooks != nil {
		_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventInvoicePaymentFailed, map[string]any{
			"id":     inv.ID.String(),
			"status": string(inv.Status),
		})
	}
	return err
}
