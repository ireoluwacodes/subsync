package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type DunningService struct {
	clock     clock.Clock
	repos     *db.Repos
	invoices  *InvoiceService
	subs      *SubscriptionService
	nomba     *nomba.Client
	mailer    *email.MailerService
	publisher TaskPublisher
}

func NewDunningService(
	clk clock.Clock,
	repos *db.Repos,
	invoices *InvoiceService,
	subs *SubscriptionService,
	nombaClient *nomba.Client,
	mailer *email.MailerService,
	publisher TaskPublisher,
) *DunningService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &DunningService{
		clock:     clk,
		repos:     repos,
		invoices:  invoices,
		subs:      subs,
		nomba:     nombaClient,
		mailer:    mailer,
		publisher: publisher,
	}
}

func (s *DunningService) ProcessStep(ctx context.Context, tenantID, subscriptionID uuid.UUID) error {
	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return err
	}
	if sub.State != domain.SubscriptionStatePastDue {
		return nil
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, sub.CustomerID)
	if err != nil {
		return err
	}

	steps, err := utils.ParseDunningSteps(tenant.DunningConfig)
	if err != nil {
		return err
	}
	if sub.DunningStep >= len(steps) {
		return s.cancelSubscription(ctx, tenant, customer, sub)
	}

	step := steps[sub.DunningStep]
	var stepErr error
	switch step.Action {
	case "retry":
		stepErr = s.retryCharge(ctx, tenant, sub, customer)
	case "retry_and_notify":
		stepErr = s.retryCharge(ctx, tenant, sub, customer)
		subject, html := email.DunningWarningHTML(tenant.Name, sub.DunningStep+1)
		_ = s.mailer.Send(ctx, customer.Email, subject, html)
	case "mandate_fallback":
		stepErr = s.mandateFallback(ctx, tenant, sub)
	case "cancel":
		return s.cancelSubscription(ctx, tenant, customer, sub)
	default:
		stepErr = s.retryCharge(ctx, tenant, sub, customer)
	}

	sub.DunningStep++
	if sub.DunningStartedAt == nil {
		now := s.clock.Now().UTC()
		sub.DunningStartedAt = &now
	}
	_ = s.repos.Subscriptions.Update(ctx, sub)

	if stepErr == nil {
		return nil
	}

	if sub.DunningStep < len(steps) {
		delay := time.Duration(steps[sub.DunningStep].DelayDays) * 24 * time.Hour
		s.enqueueDunning(ctx, tenantID, subscriptionID, delay)
	}
	return stepErr
}

func (s *DunningService) retryCharge(ctx context.Context, tenant *domain.Tenant, sub *domain.Subscription, customer *domain.Customer) error {
	inv, err := s.repos.Invoices.LatestOpenForSubscription(ctx, tenant.ID, sub.ID)
	if err != nil {
		return err
	}
	pm, err := s.resolvePaymentMethod(ctx, tenant.ID, sub, customer.ID)
	if err != nil {
		return err
	}
	charged, chargeErr := s.invoices.ChargeWithPayment(ctx, tenant, pm, inv)
	if chargeErr != nil {
		return chargeErr
	}
	plan, err := s.repos.Plans.GetByID(ctx, tenant.ID, sub.PlanID)
	if err != nil {
		return err
	}
	sub.State = domain.SubscriptionStateActive
	sub.DunningStep = 0
	sub.DunningStartedAt = nil
	start := sub.CurrentPeriodEnd
	end := utils.PlanPeriodEnd(start, plan)
	sub.CurrentPeriodStart = start
	sub.CurrentPeriodEnd = end
	sub.NextBillingAt = &end
	_ = s.repos.Subscriptions.Update(ctx, sub)
	subject, html := email.SubscriptionConfirmedHTML(tenant.Name, charged.AmountPaid, charged.Currency)
	_ = s.mailer.Send(ctx, customer.Email, subject, html)
	return nil
}

func (s *DunningService) mandateFallback(ctx context.Context, tenant *domain.Tenant, sub *domain.Subscription) error {
	pm, err := s.resolvePaymentMethod(ctx, tenant.ID, sub, sub.CustomerID)
	if err != nil {
		return err
	}
	if pm.Type != domain.PaymentMethodDirectDebit || pm.MandateID == "" {
		return fmt.Errorf("%w: no direct debit mandate", domain.ErrValidation)
	}
	inv, err := s.repos.Invoices.LatestOpenForSubscription(ctx, tenant.ID, sub.ID)
	if err != nil {
		return err
	}
	status, err := s.nomba.GetMandateStatus(ctx, tenant, pm.MandateID)
	if err != nil {
		return err
	}
	if !status.MandateReadyForDebit() {
		return fmt.Errorf("%w: mandate not ready", domain.ErrValidation)
	}
	amount := fmt.Sprintf("%.2f", float64(inv.AmountDue)/100.0)
	_, err = s.nomba.DebitMandate(ctx, tenant, nomba.DebitMandateRequest{
		MandateID: pm.MandateID,
		Amount:    amount,
	})
	return err
}

func (s *DunningService) cancelSubscription(ctx context.Context, tenant *domain.Tenant, customer *domain.Customer, sub *domain.Subscription) error {
	_, err := s.subs.Cancel(ctx, tenant.ID, sub.ID, CancelInput{Reason: "dunning_exhausted"}, "system")
	if err != nil {
		return err
	}
	subject, html := email.DunningFinalHTML(tenant.Name)
	_ = s.mailer.Send(ctx, customer.Email, subject, html)
	return nil
}

func (s *DunningService) resolvePaymentMethod(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription, customerID uuid.UUID) (*domain.PaymentMethod, error) {
	if sub.PaymentMethodID != nil {
		return s.repos.PaymentMethods.GetByID(ctx, tenantID, *sub.PaymentMethodID)
	}
	return s.repos.PaymentMethods.GetDefaultForCustomer(ctx, tenantID, customerID)
}

func (s *DunningService) enqueueDunning(ctx context.Context, tenantID, subscriptionID uuid.UUID, delay time.Duration) {
	if s.publisher == nil {
		return
	}
	raw, _ := json.Marshal(jobPayload{TenantID: tenantID, SubscriptionID: subscriptionID})
	task := asynq.NewTask(jobDunningStep, raw)
	_, _ = s.publisher.EnqueueContext(ctx, task, asynq.ProcessIn(delay))
}
