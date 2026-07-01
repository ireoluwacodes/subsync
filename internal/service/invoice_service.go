package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/pdf"
)

type InvoiceService struct {
	repo   domain.InvoiceRepository
	cfg    *config.Config
	pdf    *pdf.Renderer
}

func NewInvoiceService(repo domain.InvoiceRepository, cfg *config.Config) *InvoiceService {
	return &InvoiceService{repo: repo, cfg: cfg, pdf: pdf.NewRenderer()}
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

func (s *InvoiceService) CreateUpgradeInvoice(ctx context.Context, tenantID uuid.UUID, sub *domain.Subscription, proration domain.ProrationResult, oldPlan, newPlan *domain.Plan) (*domain.Invoice, error) {
	now := time.Now().UTC()
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
	now := time.Now().UTC()
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
	if inv.Status != domain.InvoiceStatusOpen {
		return nil, fmt.Errorf("%w: only open invoices can be charged", domain.ErrValidation)
	}

	now := time.Now().UTC()
	inv.AttemptCount++

	mock := "success"
	if s.cfg != nil && s.cfg.BillingMockResult != "" {
		mock = s.cfg.BillingMockResult
	}

	if mock == "failure" {
		inv.NextAttemptAt = ptrTime(now.Add(24 * time.Hour))
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

func ptrTime(t time.Time) *time.Time { return &t }
