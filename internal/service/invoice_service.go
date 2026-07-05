package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/pdf"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type InvoiceService struct {
	repo     domain.InvoiceRepository
	cfg      *config.Config
	pdf      *pdf.Renderer
	nomba    *nomba.Client
	clock    clock.Clock
	webhooks *WebhookService
}

func NewInvoiceService(repo domain.InvoiceRepository, cfg *config.Config, nombaClient *nomba.Client, clk clock.Clock, webhooks *WebhookService) *InvoiceService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &InvoiceService{repo: repo, cfg: cfg, pdf: pdf.NewRenderer(), nomba: nombaClient, clock: clk, webhooks: webhooks}
}

func (s *InvoiceService) Get(ctx context.Context, tenantID, id uuid.UUID) (*domain.Invoice, []*domain.InvoiceLineItem, error) {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.repo.ListLineItems(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	return inv, items, nil
}

func (s *InvoiceService) List(ctx context.Context, tenantID uuid.UUID, filter domain.InvoiceListFilter) ([]*domain.Invoice, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *InvoiceService) CreateSubscriptionInvoice(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription, plan *domain.Plan) (*domain.Invoice, error) {
	return s.createSubscriptionInvoice(ctx, tenantID, sub, plan, nil)
}

func (s *InvoiceService) CreateSubscriptionCheckoutInvoice(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription, plan *domain.Plan) (*domain.Invoice, error) {
	return s.createSubscriptionInvoice(ctx, tenantID, sub, plan, map[string]any{
		"purpose": domain.InvoicePurposeSubscriptionCheckout,
	})
}

func (s *InvoiceService) createSubscriptionInvoice(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription, plan *domain.Plan, metadata map[string]any) (*domain.Invoice, error) {
	now := s.clock.Now().UTC()
	meta := map[string]any{}
	for k, v := range metadata {
		meta[k] = v
	}
	inv := &domain.Invoice{
		TenantID:       tenantID,
		SubscriptionID: sub.ID,
		CustomerID:     sub.CustomerID,
		Status:         domain.InvoiceStatusOpen,
		AmountDue:      plan.Amount,
		Currency:       plan.Currency,
		PeriodStart:    sub.CurrentPeriodStart,
		PeriodEnd:      sub.CurrentPeriodEnd,
		DueDate:        &now,
		NombaOrderRef:  uuid.New().String(),
		Metadata:       meta,
	}
	if err := s.repo.Create(ctx, inv); err != nil {
		return nil, err
	}
	_ = s.repo.CreateLineItem(ctx, &domain.InvoiceLineItem{
		InvoiceID:   inv.ID,
		TenantID:    tenantID,
		Type:        domain.LineItemSubscription,
		Description: plan.Name,
		Amount:      plan.Amount,
		Currency:    plan.Currency,
	})

	s.emitInvoiceCreated(ctx, tenantID, inv)
	return inv, nil
}

func (s *InvoiceService) CreateUpgradeInvoice(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription, proration domain.ProrationResult, oldPlan, newPlan *domain.Plan) (*domain.Invoice, error) {
	now := s.clock.Now().UTC()
	inv := &domain.Invoice{
		TenantID:       tenantID,
		SubscriptionID: sub.ID,
		CustomerID:     sub.CustomerID,
		Status:         domain.InvoiceStatusOpen,
		AmountDue:      proration.NetAmount,
		Currency:       newPlan.Currency,
		PeriodStart:    sub.CurrentPeriodStart,
		PeriodEnd:      sub.CurrentPeriodEnd,
		DueDate:        &now,
		NombaOrderRef:  uuid.New().String(),
		Metadata:       map[string]any{},
	}
	if inv.AmountDue < 0 {
		inv.AmountDue = 0
	}

	if err := s.repo.Create(ctx, inv); err != nil {
		return nil, err
	}

	if proration.CreditAmount > 0 {
		_ = s.repo.CreateLineItem(ctx, &domain.InvoiceLineItem{
			InvoiceID:   inv.ID,
			TenantID:    tenantID,
			Type:        domain.LineItemProrationCredit,
			Description: fmt.Sprintf("Unused %s", oldPlan.Name),
			Amount:      -proration.CreditAmount,
			Currency:    newPlan.Currency,
		})
	}
	if proration.DebitAmount > 0 {
		_ = s.repo.CreateLineItem(ctx, &domain.InvoiceLineItem{
			InvoiceID:   inv.ID,
			TenantID:    tenantID,
			Type:        domain.LineItemProrationDebit,
			Description: fmt.Sprintf("Prorated %s", newPlan.Name),
			Amount:      proration.DebitAmount,
			Currency:    newPlan.Currency,
		})
	}

	s.emitInvoiceCreated(ctx, tenantID, inv)
	return inv, nil
}

func (s *InvoiceService) Void(ctx context.Context, tenantID, id uuid.UUID) (*domain.Invoice, error) {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.InvoiceStatusOpen {
		return nil, fmt.Errorf("%w: only open invoices can be voided", domain.ErrValidation)
	}
	now := s.clock.Now().UTC()
	inv.Status = domain.InvoiceStatusVoid
	inv.VoidedAt = &now
	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *InvoiceService) Charge(ctx context.Context, tenantID, id uuid.UUID) (*domain.Invoice, error) {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return s.chargeInvoice(ctx, nil, nil, inv, "")
}

func (s *InvoiceService) ChargeWithPayment(ctx context.Context, tenant *domain.Tenant, pm *domain.PaymentMethod, inv *domain.Invoice, customerEmail string) (*domain.Invoice, error) {
	return s.chargeInvoice(ctx, tenant, pm, inv, customerEmail)
}

func (s *InvoiceService) chargeInvoice(ctx context.Context, tenant *domain.Tenant, pm *domain.PaymentMethod, inv *domain.Invoice, customerEmail string) (*domain.Invoice, error) {
	if inv.Status != domain.InvoiceStatusOpen {
		return nil, fmt.Errorf("%w: only open invoices can be charged", domain.ErrValidation)
	}
	now := s.clock.Now().UTC()
	inv.AttemptCount++

	useMock := s.cfg == nil || s.cfg.BillingMockResult != ""
	if useMock {
		mock := "success"
		if s.cfg != nil && s.cfg.BillingMockResult != "" {
			mock = s.cfg.BillingMockResult
		}
		if mock == "failure" {
			inv.NextAttemptAt = utils.PtrTime(now.Add(24 * time.Hour))
			if err := s.repo.Update(ctx, inv); err != nil {
				return nil, err
			}
			return inv, fmt.Errorf("%w: mock charge failed", domain.ErrValidation)
		}
		inv.Status = domain.InvoiceStatusPaid
		inv.AmountPaid = inv.AmountDue
		inv.PaidAt = &now
		if err := s.repo.Update(ctx, inv); err != nil {
			return nil, err
		}
		return inv, nil
	}

	if tenant == nil || pm == nil || s.nomba == nil {
		return nil, fmt.Errorf("%w: tenant and payment method required for live charge", domain.ErrValidation)
	}
	if pm.TokenKey == "" {
		return nil, fmt.Errorf("%w: payment method missing token", domain.ErrValidation)
	}

	callbackURL := billingCallbackURL(s.cfg)
	result, err := s.nomba.TokenizedCardPayment(ctx, tenant, nomba.TokenizedCardPaymentRequest{
		TokenKey: pm.TokenKey,
		Order: nomba.Order{
			OrderReference: inv.NombaOrderRef,
			CustomerEmail:  customerEmail,
			Amount:         float64(inv.AmountDue) / 100.0,
			Currency:       nomba.Currency(inv.Currency),
			AccountID:      tenant.NombaOrderAccountID(),
			CallbackURL:    callbackURL,
			OrderMetaData: map[string]string{
				"invoice_id": inv.ID.String(),
				"purpose":    "billing_charge",
			},
		},
	})
	if err != nil {
		inv.NextAttemptAt = utils.PtrTime(now.Add(24 * time.Hour))
		_ = s.repo.Update(ctx, inv)
		return inv, err
	}
	if !result.Status {
		inv.NextAttemptAt = utils.PtrTime(now.Add(24 * time.Hour))
		_ = s.repo.Update(ctx, inv)
		return inv, fmt.Errorf("%w: nomba charge declined: %s", domain.ErrValidation, result.Message)
	}

	inv.Status = domain.InvoiceStatusProcessing
	inv.NombaTransactionID = result.Message
	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *InvoiceService) SetPDFURL(ctx context.Context, tenantID, id uuid.UUID, pdfURL string) error {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if inv.Metadata == nil {
		inv.Metadata = map[string]any{}
	}
	inv.Metadata["pdf_url"] = pdfURL
	return s.repo.Update(ctx, inv)
}

func (s *InvoiceService) RenderPDF(ctx context.Context, tenantID, id uuid.UUID, tenant *domain.Tenant) ([]byte, error) {
	inv, items, err := s.Get(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return s.pdf.RenderInvoice(tenant, inv, items)
}

func (s *InvoiceService) CustomerPaidTotal(ctx context.Context, tenantID, customerID uuid.UUID) (int64, error) {
	return s.repo.SumPaidByCustomer(ctx, tenantID, customerID)
}

func (s *InvoiceService) SetWebhooks(webhooks *WebhookService) {
	s.webhooks = webhooks
}

func billingCallbackURL(cfg *config.Config) string {
	if cfg == nil || cfg.PublicBaseURL == "" {
		return "https://subsync.io/billing/callback"
	}
	return strings.TrimRight(cfg.PublicBaseURL, "/") + "/billing/callback"
}

func (s *InvoiceService) emitInvoiceCreated(ctx context.Context, tenantID uuid.UUID, inv *domain.Invoice) {
	if s.webhooks == nil {
		return
	}
	_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventInvoiceCreated, map[string]any{
		"id":     inv.ID.String(),
		"status": string(inv.Status),
		"amount_due": inv.AmountDue,
		"currency": inv.Currency,
	})
}
