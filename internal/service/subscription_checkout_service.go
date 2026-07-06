package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/email"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

const checkoutTokenizationAmount = 100.0

type SubscriptionCheckoutService struct {
	cfg       *config.Config
	clock     clock.Clock
	repos     *db.Repos
	invoices  *InvoiceService
	subs      *SubscriptionService
	nomba     *nomba.Client
	mailer    *email.MailerService
	webhooks  *WebhookService
}

func NewSubscriptionCheckoutService(
	cfg *config.Config,
	clk clock.Clock,
	repos *db.Repos,
	invoices *InvoiceService,
	subs *SubscriptionService,
	nombaClient *nomba.Client,
	mailer *email.MailerService,
	webhooks *WebhookService,
) *SubscriptionCheckoutService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &SubscriptionCheckoutService{
		cfg:      cfg,
		clock:    clk,
		repos:    repos,
		invoices: invoices,
		subs:     subs,
		nomba:    nombaClient,
		mailer:   mailer,
		webhooks: webhooks,
	}
}

type SubscriptionCheckoutInput struct {
	CustomerID            uuid.UUID
	PlanID                uuid.UUID
	SuccessURL            string
	CancelURL             string
	SendCheckoutEmail     bool
	CardOnly              bool // deprecated: checkout is card-only by default
	AllowBankTransfer     bool
	AllowedPaymentMethods []string
}

type CardCaptureInput struct {
	SuccessURL string
	CancelURL  string
	SendEmail  bool
}

type CardCaptureResult struct {
	CheckoutURL    string `json:"checkout_url"`
	OrderReference string `json:"order_reference"`
}

type SubscriptionCheckoutResult struct {
	SubscriptionID uuid.UUID  `json:"subscription_id"`
	InvoiceID      *uuid.UUID `json:"invoice_id,omitempty"`
	CheckoutURL    string     `json:"checkout_url"`
	OrderReference string     `json:"order_reference"`
	Status         string     `json:"status"`
}

func (s *SubscriptionCheckoutService) StartCheckout(ctx context.Context, tenantID uuid.UUID, in SubscriptionCheckoutInput) (*SubscriptionCheckoutResult, error) {
	successURL, err := resolveCheckoutSuccessURL(in.SuccessURL, s.cfg)
	if err != nil {
		return nil, err
	}
	in.SuccessURL = successURL
	if in.CancelURL != "" {
		if err := validateRedirectURL(in.CancelURL, s.cfg); err != nil {
			return nil, err
		}
	}

	plan, err := s.repos.Plans.GetByID(ctx, tenantID, in.PlanID)
	if err != nil {
		return nil, err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, in.CustomerID)
	if err != nil {
		return nil, err
	}

	if err := s.ensureNoIncompleteCheckout(ctx, tenantID, in.CustomerID, in.PlanID); err != nil {
		return nil, err
	}

	sub, err := s.createIncompleteSubscription(ctx, tenantID, in.CustomerID, in.PlanID, plan)
	if err != nil {
		return nil, err
	}

	return s.beginCheckout(ctx, tenantID, sub, plan, customer, in)
}

func (s *SubscriptionCheckoutService) ResumeCheckout(ctx context.Context, tenantID, subscriptionID uuid.UUID, in SubscriptionCheckoutInput) (*SubscriptionCheckoutResult, error) {
	successURL, err := resolveCheckoutSuccessURL(in.SuccessURL, s.cfg)
	if err != nil {
		return nil, err
	}
	in.SuccessURL = successURL
	if in.CancelURL != "" {
		if err := validateRedirectURL(in.CancelURL, s.cfg); err != nil {
			return nil, err
		}
	}

	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if sub.State != domain.SubscriptionStateIncomplete {
		return nil, fmt.Errorf("%w: subscription is not awaiting checkout", domain.ErrValidation)
	}
	if sub.CustomerID != in.CustomerID || sub.PlanID != in.PlanID {
		return nil, fmt.Errorf("%w: customer_id and plan_id must match the subscription", domain.ErrValidation)
	}

	plan, err := s.repos.Plans.GetByID(ctx, tenantID, sub.PlanID)
	if err != nil {
		return nil, err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, sub.CustomerID)
	if err != nil {
		return nil, err
	}

	if err := s.voidOpenCheckoutInvoices(ctx, tenantID, sub.ID); err != nil {
		return nil, err
	}

	return s.beginCheckout(ctx, tenantID, sub, plan, customer, in)
}

func (s *SubscriptionCheckoutService) ensureNoIncompleteCheckout(ctx context.Context, tenantID, customerID, planID uuid.UUID) error {
	state := string(domain.SubscriptionStateIncomplete)
	subs, _, err := s.repos.Subscriptions.List(ctx, tenantID, domain.SubscriptionListFilter{
		CustomerID: &customerID,
		PlanID:     &planID,
		State:      state,
		Limit:      1,
	})
	if err != nil {
		return err
	}
	if len(subs) > 0 {
		return fmt.Errorf("%w: incomplete subscription already exists for this customer and plan; use resume checkout", domain.ErrConflict)
	}
	return nil
}

func (s *SubscriptionCheckoutService) createIncompleteSubscription(ctx context.Context, tenantID, customerID, planID uuid.UUID, plan *domain.Plan) (*domain.Subscription, error) {
	now := s.clock.Now().UTC()
	periodEnd := utils.PlanPeriodEnd(now, plan)

	sub := &domain.Subscription{
		TenantID:           tenantID,
		CustomerID:         customerID,
		PlanID:             planID,
		State:              domain.SubscriptionStateIncomplete,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		Metadata:           map[string]any{},
	}
	if plan.TrialDays > 0 {
		trialEnd := now.AddDate(0, 0, plan.TrialDays)
		sub.TrialEndsAt = &trialEnd
	}

	if err := s.repos.Subscriptions.Create(ctx, sub); err != nil {
		return nil, err
	}

	_ = s.repos.Subscriptions.RecordTransition(ctx, &domain.SubscriptionTransition{
		SubscriptionID: sub.ID,
		TenantID:       tenantID,
		FromState:      "",
		ToState:        domain.SubscriptionStateIncomplete,
		Reason:         "checkout_started",
		Actor:          "system",
		Metadata:       map[string]any{},
	})

	if s.webhooks != nil {
		_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventSubscriptionCreated, map[string]any{
			"id":    sub.ID.String(),
			"state": string(sub.State),
		})
	}

	return sub, nil
}

func (s *SubscriptionCheckoutService) beginCheckout(
	ctx context.Context,
	tenantID uuid.UUID,
	sub *domain.Subscription,
	plan *domain.Plan,
	customer *domain.Customer,
	in SubscriptionCheckoutInput,
) (*SubscriptionCheckoutResult, error) {
	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return nil, err
	}

	var inv *domain.Invoice
	var orderRef string
	var amount float64

	if plan.TrialDays > 0 {
		orderRef = domain.CheckoutOrderRefPrefix + sub.ID.String() + "-" + uuid.New().String()
		amount = checkoutTokenizationAmount
	} else {
		inv, err = s.invoices.CreateSubscriptionCheckoutInvoice(ctx, tenantID, sub, plan)
		if err != nil {
			return nil, err
		}
		orderRef = inv.NombaOrderRef
		amount = float64(plan.Amount) / 100.0
	}

	result, err := s.nomba.CreateOrder(ctx, tenant, nomba.CreateOrderRequest{
		TokenizeCard: true,
		Order: nomba.Order{
			OrderReference:        orderRef,
			CustomerEmail:         customer.Email,
			Amount:                amount,
			Currency:              nomba.Currency(plan.Currency),
			AccountID:             tenant.NombaOrderAccountID(),
			CallbackURL:           strings.TrimSpace(in.SuccessURL),
			AllowedPaymentMethods: checkoutAllowedPaymentMethods(in),
			OrderMetaData: map[string]string{
				"purpose":         domain.InvoicePurposeSubscriptionCheckout,
				"subscription_id": sub.ID.String(),
				"customer_id":     customer.ID.String(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if in.SendCheckoutEmail && s.mailer != nil {
		subject, html := email.CheckoutLinkHTML(tenant.Name, plan.Name, amount, string(plan.Currency), result.CheckoutLink)
		_ = s.mailer.Send(ctx, customer.Email, subject, html)
	}

	out := &SubscriptionCheckoutResult{
		SubscriptionID: sub.ID,
		CheckoutURL:    result.CheckoutLink,
		OrderReference: orderRef,
		Status:         string(domain.SubscriptionStateIncomplete),
	}
	if inv != nil {
		out.InvoiceID = &inv.ID
	}
	return out, nil
}

func (s *SubscriptionCheckoutService) voidOpenCheckoutInvoices(ctx context.Context, tenantID, subscriptionID uuid.UUID) error {
	invoices, _, err := s.repos.Invoices.List(ctx, tenantID, domain.InvoiceListFilter{
		SubscriptionID: &subscriptionID,
		Status:         string(domain.InvoiceStatusOpen),
		Limit:          50,
	})
	if err != nil {
		return err
	}
	for _, inv := range invoices {
		if inv.Metadata == nil {
			continue
		}
		if inv.Metadata["purpose"] != domain.InvoicePurposeSubscriptionCheckout {
			continue
		}
		if _, err := s.invoices.Void(ctx, tenantID, inv.ID); err != nil {
			return err
		}
	}
	return nil
}

func validateRedirectURL(raw string, cfg *config.Config) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%w: invalid redirect url", domain.ErrValidation)
	}
	if cfg != nil && !cfg.IsDevelopment() && u.Scheme != "https" {
		return fmt.Errorf("%w: redirect url must use https in production", domain.ErrValidation)
	}
	return nil
}

func IsSubscriptionCheckoutInvoice(inv *domain.Invoice) bool {
	if inv == nil || inv.Metadata == nil {
		return false
	}
	purpose, _ := inv.Metadata["purpose"].(string)
	return purpose == domain.InvoicePurposeSubscriptionCheckout
}

func ParseCheckoutSubscriptionID(orderRef string) (uuid.UUID, bool) {
	return parsePrefixedSubscriptionID(orderRef, domain.CheckoutOrderRefPrefix)
}

func checkoutAllowedPaymentMethods(in SubscriptionCheckoutInput) []nomba.PaymentMethod {
	if methods := nomba.ParsePaymentMethods(in.AllowedPaymentMethods); len(methods) > 0 {
		return methods
	}
	if in.AllowBankTransfer {
		return []nomba.PaymentMethod{nomba.PaymentMethodCard, nomba.PaymentMethodTransfer}
	}
	return []nomba.PaymentMethod{nomba.PaymentMethodCard}
}

// StartCardCapture creates a card-only Nomba checkout to save a payment method before renewal.
func (s *SubscriptionCheckoutService) StartCardCapture(
	ctx context.Context,
	tenantID, subscriptionID uuid.UUID,
	in CardCaptureInput,
) (*CardCaptureResult, error) {
	successURL, err := resolveCheckoutSuccessURL(in.SuccessURL, s.cfg)
	if err != nil {
		return nil, err
	}
	in.SuccessURL = successURL
	if in.CancelURL != "" {
		if err := validateRedirectURL(in.CancelURL, s.cfg); err != nil {
			return nil, err
		}
	}

	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, subscriptionID)
	if err != nil {
		return nil, err
	}
	switch sub.State {
	case domain.SubscriptionStateActive, domain.SubscriptionStateTrialing, domain.SubscriptionStatePastDue, domain.SubscriptionStateIncomplete:
	default:
		return nil, fmt.Errorf("%w: subscription cannot capture a payment method in state %s", domain.ErrValidation, sub.State)
	}
	if sub.PaymentMethodID != nil && !subscriptionAwaitingPaymentMethod(sub) {
		return nil, fmt.Errorf("%w: subscription already has a payment method", domain.ErrConflict)
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return nil, err
	}
	customer, err := s.repos.Customers.GetByID(ctx, tenantID, sub.CustomerID)
	if err != nil {
		return nil, err
	}

	orderRef := domain.CardCaptureOrderRefPrefix + sub.ID.String()
	result, err := s.nomba.CreateOrder(ctx, tenant, nomba.CreateOrderRequest{
		TokenizeCard: true,
		Order: nomba.Order{
			OrderReference: orderRef,
			CustomerEmail:  customer.Email,
			Amount:         checkoutTokenizationAmount,
			Currency:       nomba.CurrencyNGN,
			AccountID:      tenant.NombaOrderAccountID(),
			CallbackURL:    strings.TrimSpace(in.SuccessURL),
			AllowedPaymentMethods: []nomba.PaymentMethod{
				nomba.PaymentMethodCard,
			},
			OrderMetaData: map[string]string{
				"purpose":         domain.InvoicePurposeCardCapture,
				"subscription_id": sub.ID.String(),
				"customer_id":     customer.ID.String(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if in.SendEmail && s.mailer != nil {
		subject, html := email.PaymentMethodCaptureRequiredHTML(tenant.Name, "", result.CheckoutLink)
		_ = s.mailer.Send(ctx, customer.Email, subject, html)
	}

	return &CardCaptureResult{
		CheckoutURL:    result.CheckoutLink,
		OrderReference: orderRef,
	}, nil
}
