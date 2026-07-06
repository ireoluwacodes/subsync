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
	"github.com/ireoluwacodes/subsync/internal/nomba"
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
	cfg            *config.Config
	clock          clock.Clock
	repos          *db.Repos
	invoices       *InvoiceService
	subs           *SubscriptionService
	paymentMethods *PaymentMethodService
	mailer         *email.MailerService
	publisher      TaskPublisher
	webhooks       *WebhookService
	portal         *PortalService
	pmResolver     *PaymentMethodResolver
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
	svc := &BillingService{
		cfg:       cfg,
		clock:     clk,
		repos:     repos,
		invoices:  invoices,
		subs:      subs,
		mailer:    mailer,
		publisher: publisher,
		webhooks:  webhooks,
	}
	if repos != nil {
		svc.pmResolver = NewPaymentMethodResolver(repos.PaymentMethods)
	}
	return svc
}

func (s *BillingService) SetPaymentMethods(paymentMethods *PaymentMethodService) {
	s.paymentMethods = paymentMethods
}

func (s *BillingService) SetPortal(portal *PortalService) {
	s.portal = portal
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

func (s *BillingService) ProcessPaymentMethodReminders(ctx context.Context, limit int) (int, error) {
	subs, err := s.repos.Subscriptions.ListAwaitingPaymentMethodBeforeBilling(ctx, s.clock.Now().UTC(), limit)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, sub := range subs {
		if s.pmResolver != nil && s.pmResolver.HasChargeablePM(ctx, sub.TenantID, sub) {
			continue
		}
		billingAt := sub.NextBillingAt
		if billingAt == nil {
			continue
		}
		now := s.clock.Now().UTC()
		if !billingAt.After(now) {
			continue
		}
		hoursUntil := billingAt.Sub(now)
		due := pmRemindersDue(sub, hoursUntil)
		if len(due) == 0 {
			continue
		}

		tenant, err := s.repos.Tenants.GetByID(ctx, sub.TenantID)
		if err != nil {
			continue
		}
		customer, err := s.repos.Customers.GetByID(ctx, sub.TenantID, sub.CustomerID)
		if err != nil {
			continue
		}
		plan, err := s.repos.Plans.GetByID(ctx, sub.TenantID, sub.PlanID)
		if err != nil {
			continue
		}

		daysUntil := int(hoursUntil.Hours() / 24)
		if daysUntil < 1 {
			daysUntil = 1
		}
		s.sendPaymentMethodCaptureReminder(ctx, tenant, customer, sub, plan, daysUntil)
		for _, key := range due {
			markPMReminderSent(sub, key)
		}
		if err := s.repos.Subscriptions.Update(ctx, sub); err != nil {
			continue
		}
		sent++
	}
	return sent, nil
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

	if s.pmResolver == nil {
		return fmt.Errorf("%w: payment method resolver not configured", domain.ErrValidation)
	}
	pm, err := s.pmResolver.ResolvePrimaryPM(ctx, tenantID, sub)
	if err != nil {
		pm, err = s.pmResolver.ResolveMandatePM(ctx, tenantID, sub)
		if err != nil {
			tenant, tErr := s.repos.Tenants.GetByID(ctx, tenantID)
			if tErr != nil {
				return tErr
			}
			return s.handleRenewalWithoutPaymentMethod(ctx, tenant, customer, sub, plan, inv)
		}
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return err
	}

	charged, chargeErr := s.invoices.ChargeWithPayment(ctx, tenant, pm, inv, customer.Email)
	return s.ApplyChargeOutcome(ctx, tenant, customer, sub, plan, charged, chargeErr)
}

// ApplyChargeOutcome handles sync mock success, async processing, or charge failure.
func (s *BillingService) ApplyChargeOutcome(ctx context.Context, tenant *domain.Tenant, customer *domain.Customer, sub *domain.Subscription, plan *domain.Plan, charged *domain.Invoice, chargeErr error) error {
	if chargeErr != nil {
		return s.handleChargeFailure(ctx, tenant, customer, sub, charged, chargeErr)
	}
	if charged != nil && charged.Status == domain.InvoiceStatusProcessing {
		return nil
	}
	return s.CompleteSuccessfulCharge(ctx, tenant, customer, sub, plan, charged)
}

// CompleteSuccessfulCharge advances the subscription after a confirmed (sync or webhook) payment.
func (s *BillingService) CompleteSuccessfulCharge(ctx context.Context, tenant *domain.Tenant, customer *domain.Customer, sub *domain.Subscription, plan *domain.Plan, inv *domain.Invoice) error {
	return s.handleChargeSuccess(ctx, tenant, customer, sub, plan, inv)
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

	subject, html := email.SubscriptionConfirmedHTML(tenant.Name, inv.AmountPaid, inv.Currency, "")
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

// CompleteCheckoutFromWebhook activates an incomplete subscription after Nomba checkout payment.
func (s *BillingService) CompleteCheckoutFromWebhook(
	ctx context.Context,
	tenantID uuid.UUID,
	inv *domain.Invoice,
	tokenKey, transactionID string,
	tx nomba.WebhookTransaction,
	order *nomba.WebhookOrder,
) error {
	if inv == nil {
		return fmt.Errorf("%w: checkout invoice required", domain.ErrValidation)
	}
	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, inv.SubscriptionID)
	if err != nil {
		return err
	}
	if sub.State != domain.SubscriptionStateIncomplete {
		return nil
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

	if err := s.activateCheckoutSubscription(ctx, tenant, customer, sub, plan, inv, &tokenKey, transactionID, tx, order); err != nil {
		return err
	}

	if s.webhooks != nil {
		_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventInvoicePaid, map[string]any{
			"id":          inv.ID.String(),
			"status":      string(inv.Status),
			"amount_paid": inv.AmountPaid,
			"currency":    inv.Currency,
		})
	}
	return nil
}

// CompleteTrialCheckoutFromWebhook tokenizes a card and starts a trialing subscription.
func (s *BillingService) CompleteTrialCheckoutFromWebhook(
	ctx context.Context,
	tenantID, subscriptionID uuid.UUID,
	tokenKey, transactionID string,
	tx nomba.WebhookTransaction,
	order *nomba.WebhookOrder,
) error {
	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}
	if sub.State != domain.SubscriptionStateIncomplete {
		return nil
	}
	plan, err := s.repos.Plans.GetByID(ctx, tenantID, sub.PlanID)
	if err != nil {
		return err
	}
	if plan.TrialDays <= 0 {
		return fmt.Errorf("%w: subscription is not a trial checkout", domain.ErrValidation)
	}
	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, sub.CustomerID)
	if err != nil {
		return err
	}

	return s.activateCheckoutSubscription(ctx, tenant, customer, sub, plan, nil, &tokenKey, transactionID, tx, order)
}

func (s *BillingService) activateCheckoutSubscription(
	ctx context.Context,
	tenant *domain.Tenant,
	customer *domain.Customer,
	sub *domain.Subscription,
	plan *domain.Plan,
	inv *domain.Invoice,
	tokenKey *string,
	transactionID string,
	tx nomba.WebhookTransaction,
	order *nomba.WebhookOrder,
) error {
	key := ""
	if tokenKey != nil {
		key = *tokenKey
	}
	if nomba.IsPlaceholderToken(key) {
		key = ""
	}
	if key == "" {
		key = nomba.EffectiveTokenKey(tx, nil)
	}

	transferPaid := key == "" && nomba.IsTransferPayment(tx, order)
	if key == "" && !transferPaid {
		return fmt.Errorf("%w: missing tokenKey for card checkout", domain.ErrValidation)
	}

	from := sub.State
	sub.DunningStep = 0
	sub.DunningStartedAt = nil

	if key != "" {
		if s.paymentMethods == nil {
			return fmt.Errorf("%w: payment method service not configured", domain.ErrValidation)
		}
		cardLast4, cardBrand := cardDetailsFromTransaction(tx)
		setDefault := true
		if s.pmResolver != nil {
			setDefault = !s.pmResolver.CustomerHasDefaultCard(ctx, sub.TenantID, sub.CustomerID)
		}
		pm, err := s.paymentMethods.Create(ctx, sub.TenantID, CreatePaymentMethodInput{
			CustomerID: sub.CustomerID,
			Type:       domain.PaymentMethodTokenizedCard,
			TokenKey:   key,
			CardLast4:  cardLast4,
			CardBrand:  cardBrand,
			IsDefault:  setDefault,
		})
		if err != nil {
			return err
		}
		sub.PaymentMethodID = &pm.ID
		setSubscriptionMeta(sub, domain.SubscriptionMetaAwaitingPaymentMethod, nil)
		clearPMReminderMetadata(sub)
	} else {
		setSubscriptionMeta(sub, domain.SubscriptionMetaAwaitingPaymentMethod, true)
	}

	if plan.TrialDays > 0 {
		sub.State = domain.SubscriptionStateTrialing
		if sub.TrialEndsAt == nil {
			trialEnd := s.clock.Now().UTC().AddDate(0, 0, plan.TrialDays)
			sub.TrialEndsAt = &trialEnd
		}
		sub.NextBillingAt = sub.TrialEndsAt
	} else {
		sub.State = domain.SubscriptionStateActive
		sub.NextBillingAt = &sub.CurrentPeriodEnd
	}

	if err := s.repos.Subscriptions.Update(ctx, sub); err != nil {
		return err
	}

	reason := "checkout_completed"
	if transferPaid {
		reason = "checkout_completed_transfer"
	}
	_ = s.repos.Subscriptions.RecordTransition(ctx, &domain.SubscriptionTransition{
		SubscriptionID: sub.ID,
		TenantID:       sub.TenantID,
		FromState:      from,
		ToState:        sub.State,
		Reason:         reason,
		Actor:          "system",
		Metadata:       map[string]any{},
	})

	if s.webhooks != nil {
		_ = s.webhooks.Emit(ctx, sub.TenantID, domain.WebhookEventSubscriptionUpdated, map[string]any{
			"id":    sub.ID.String(),
			"state": string(sub.State),
		})
	}

	if inv != nil && inv.Status == domain.InvoiceStatusPaid {
		s.sendSubscriptionConfirmedEmail(ctx, tenant, customer, sub, inv)
		s.enqueueInvoicePDF(ctx, tenant.ID, inv.ID)
	} else if !transferPaid {
		s.sendSubscriptionConfirmedEmail(ctx, tenant, customer, sub, nil)
	}

	if transferPaid {
		s.sendPaymentMethodCaptureRequiredEmail(ctx, tenant, customer, sub, plan)
	}

	_ = transactionID
	return nil
}

func cardDetailsFromTransaction(tx nomba.WebhookTransaction) (last4, brand string) {
	_ = tx
	return "", ""
}

func (s *BillingService) handleRenewalWithoutPaymentMethod(
	ctx context.Context,
	tenant *domain.Tenant,
	customer *domain.Customer,
	sub *domain.Subscription,
	plan *domain.Plan,
	inv *domain.Invoice,
) error {
	if s.pmResolver != nil && s.pmResolver.HasPendingMandate(ctx, tenant.ID, sub) {
		s.sendPaymentMethodCaptureReminder(ctx, tenant, customer, sub, plan, 0)
		return nil
	}

	if inv != nil && inv.Status == domain.InvoiceStatusOpen {
		if _, err := s.invoices.Void(ctx, tenant.ID, inv.ID); err != nil {
			return err
		}
	}

	if _, err := s.subs.Cancel(ctx, tenant.ID, sub.ID, CancelInput{
		CancelAtPeriodEnd: false,
		Reason:            "no_payment_method_at_renewal",
	}, "system"); err != nil {
		return err
	}

	_ = customer
	_ = plan
	return nil
}

func (s *BillingService) sendSubscriptionConfirmedEmail(
	ctx context.Context,
	tenant *domain.Tenant,
	customer *domain.Customer,
	sub *domain.Subscription,
	inv *domain.Invoice,
) {
	if s.mailer == nil {
		return
	}
	portalURL := ""
	if s.portal != nil {
		portalURL, _ = s.portal.CreatePaymentMethodCaptureLink(ctx, tenant.ID, sub.ID)
	}
	var amount int64
	currency := "NGN"
	if inv != nil {
		amount = inv.AmountPaid
		currency = inv.Currency
	}
	subject, htmlBody := email.SubscriptionConfirmedHTML(tenant.Name, amount, currency, portalURL)
	_ = s.mailer.Send(ctx, customer.Email, subject, htmlBody)
}

func (s *BillingService) sendPaymentMethodCaptureRequiredEmail(
	ctx context.Context,
	tenant *domain.Tenant,
	customer *domain.Customer,
	sub *domain.Subscription,
	plan *domain.Plan,
) {
	if s.mailer == nil {
		return
	}
	captureURL := ""
	if s.portal != nil {
		if link, err := s.portal.CreatePaymentMethodCaptureLink(ctx, tenant.ID, sub.ID); err == nil {
			captureURL = link
		}
	}
	subject, html := email.PaymentMethodCaptureRequiredHTML(tenant.Name, plan.Name, captureURL)
	_ = s.mailer.Send(ctx, customer.Email, subject, html)
}

func (s *BillingService) sendPaymentMethodCaptureReminder(
	ctx context.Context,
	tenant *domain.Tenant,
	customer *domain.Customer,
	sub *domain.Subscription,
	plan *domain.Plan,
	daysUntilBilling int,
) {
	if s.mailer == nil {
		return
	}
	captureURL := ""
	if s.portal != nil {
		if link, err := s.portal.CreatePaymentMethodCaptureLink(ctx, tenant.ID, sub.ID); err == nil {
			captureURL = link
		}
	}
	subject, html := email.PaymentMethodCaptureReminderHTML(tenant.Name, plan.Name, captureURL, daysUntilBilling)
	_ = s.mailer.Send(ctx, customer.Email, subject, html)
}

// CompleteCardCaptureFromWebhook attaches a tokenized card after a capture-{subscriptionID} checkout.
func (s *BillingService) CompleteCardCaptureFromWebhook(
	ctx context.Context,
	tenantID, subscriptionID uuid.UUID,
	tokenKey string,
	tx nomba.WebhookTransaction,
) error {
	if tokenKey == "" {
		tokenKey = tx.TokenKey
	}
	if tokenKey == "" {
		return fmt.Errorf("%w: missing tokenKey for card capture", domain.ErrValidation)
	}
	if s.paymentMethods == nil {
		return fmt.Errorf("%w: payment method service not configured", domain.ErrValidation)
	}

	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}

	cardLast4, cardBrand := cardDetailsFromTransaction(tx)
	setDefault := true
	if s.pmResolver != nil {
		setDefault = !s.pmResolver.CustomerHasDefaultCard(ctx, tenantID, sub.CustomerID)
	}
	pm, err := s.paymentMethods.Create(ctx, tenantID, CreatePaymentMethodInput{
		CustomerID: sub.CustomerID,
		Type:       domain.PaymentMethodTokenizedCard,
		TokenKey:   tokenKey,
		CardLast4:  cardLast4,
		CardBrand:  cardBrand,
		IsDefault:  setDefault,
	})
	if err != nil {
		return err
	}

	sub.PaymentMethodID = &pm.ID
	setSubscriptionMeta(sub, domain.SubscriptionMetaAwaitingPaymentMethod, nil)
	clearPMReminderMetadata(sub)
	if err := s.repos.Subscriptions.Update(ctx, sub); err != nil {
		return err
	}

	if s.webhooks != nil {
		_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventPaymentMethodAttached, map[string]any{
			"id":          pm.ID.String(),
			"customer_id": pm.CustomerID.String(),
		})
		_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventSubscriptionUpdated, map[string]any{
			"id":    sub.ID.String(),
			"state": string(sub.State),
		})
	}

	// Attempt renewal immediately if billing was waiting on a card.
	if sub.State == domain.SubscriptionStateActive || sub.State == domain.SubscriptionStatePastDue {
		_ = s.ChargeDueSubscription(ctx, tenantID, subscriptionID)
	}
	return nil
}
