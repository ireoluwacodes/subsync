package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
)

type NombaEventService struct {
	clock    clock.Clock
	repos    *db.Repos
	billing  *BillingService
	webhooks *WebhookService
	portal   *PortalService
}

func NewNombaEventService(
	clk clock.Clock,
	repos *db.Repos,
	billing *BillingService,
	webhooks *WebhookService,
	portal *PortalService,
) *NombaEventService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &NombaEventService{
		clock:    clk,
		repos:    repos,
		billing:  billing,
		webhooks: webhooks,
		portal:   portal,
	}
}

func (s *NombaEventService) ProcessInbound(ctx context.Context, tenantID uuid.UUID, rawBody []byte) error {
	var event nomba.WebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return fmt.Errorf("parse webhook: %w", err)
	}
	if event.RequestID == "" {
		return fmt.Errorf("%w: missing requestId", domain.ErrValidation)
	}

	existing, err := s.repos.NombaEvents.GetByEventID(ctx, tenantID, event.RequestID)
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && err != domain.ErrNotFound {
		return err
	}

	var payload map[string]any
	_ = json.Unmarshal(rawBody, &payload)

	record := &domain.NombaEvent{
		TenantID:  tenantID,
		EventID:   event.RequestID,
		EventType: string(event.EventType),
		Payload:   payload,
	}
	if err := s.repos.NombaEvents.Create(ctx, record); err != nil {
		if err == domain.ErrConflict {
			return nil
		}
		return err
	}

	processErr := s.dispatch(ctx, tenantID, event)
	if processErr != nil {
		_ = s.repos.NombaEvents.MarkFailed(ctx, record.ID, processErr.Error())
		return processErr
	}
	return s.repos.NombaEvents.MarkProcessed(ctx, record.ID)
}

func (s *NombaEventService) dispatch(ctx context.Context, tenantID uuid.UUID, event nomba.WebhookEvent) error {
	switch event.EventType {
	case nomba.WebhookEventPaymentSuccess:
		return s.handlePaymentSuccess(ctx, tenantID, event)
	case nomba.WebhookEventPaymentFailed:
		return s.handlePaymentFailed(ctx, tenantID, event)
	case nomba.WebhookEventPaymentReversal:
		return s.handlePaymentReversal(ctx, tenantID, event)
	default:
		return nil
	}
}

func (s *NombaEventService) handlePaymentSuccess(ctx context.Context, tenantID uuid.UUID, event nomba.WebhookEvent) error {
	tx := event.Data.Transaction
	orderRef := tx.MerchantTxRef
	if orderRef == "" {
		orderRef = tx.AliasAccountReference
	}

	if strings.HasPrefix(orderRef, "portal-") {
		return s.portal.HandlePaymentSuccess(ctx, tenantID, orderRef, tx.TokenKey, tx)
	}

	inv, err := s.matchInvoice(ctx, tenantID, orderRef, tx.TransactionID)
	if err != nil {
		return err
	}

	if inv != nil && IsSubscriptionCheckoutInvoice(inv) {
		return s.billing.CompleteCheckoutFromWebhook(ctx, tenantID, inv, tx.TokenKey, tx.TransactionID, tx)
	}

	if subID, ok := ParseCheckoutSubscriptionID(orderRef); ok {
		return s.billing.CompleteTrialCheckoutFromWebhook(ctx, tenantID, subID, tx.TokenKey, tx.TransactionID, tx)
	}

	if subID, ok := ParseCardCaptureSubscriptionID(orderRef); ok {
		return s.billing.CompleteCardCaptureFromWebhook(ctx, tenantID, subID, tx.TokenKey, tx)
	}

	if inv == nil && orderRef == "" {
		return nil
	}
	if inv == nil {
		return nil
	}
	return s.billing.FinalizePaidInvoice(ctx, tenantID, inv.ID, tx.TransactionID)
}

func (s *NombaEventService) handlePaymentFailed(ctx context.Context, tenantID uuid.UUID, event nomba.WebhookEvent) error {
	tx := event.Data.Transaction
	if strings.HasPrefix(tx.MerchantTxRef, "portal-") {
		return nil
	}
	inv, err := s.matchInvoice(ctx, tenantID, tx.MerchantTxRef, tx.TransactionID)
	if err != nil {
		return err
	}
	if inv == nil {
		return nil
	}
	return s.billing.HandleWebhookPaymentFailure(ctx, tenantID, inv.ID)
}

func (s *NombaEventService) handlePaymentReversal(ctx context.Context, tenantID uuid.UUID, event nomba.WebhookEvent) error {
	tx := event.Data.Transaction
	inv, err := s.matchInvoice(ctx, tenantID, tx.MerchantTxRef, tx.TransactionID)
	if err != nil {
		return err
	}
	if inv == nil {
		return nil
	}
	if inv.Status != domain.InvoiceStatusPaid {
		return nil
	}
	now := s.clock.Now().UTC()
	inv.Status = domain.InvoiceStatusVoid
	inv.VoidedAt = &now
	return s.repos.Invoices.Update(ctx, inv)
}

func (s *NombaEventService) matchInvoice(ctx context.Context, tenantID uuid.UUID, orderRef, transactionID string) (*domain.Invoice, error) {
	if orderRef != "" {
		inv, err := s.repos.Invoices.GetByNombaOrderRef(ctx, tenantID, orderRef)
		if err == nil {
			return inv, nil
		}
		if err != domain.ErrNotFound {
			return nil, err
		}
	}
	if transactionID != "" {
		inv, err := s.repos.Invoices.GetByNombaTransactionID(ctx, tenantID, transactionID)
		if err == nil {
			return inv, nil
		}
		if err != domain.ErrNotFound {
			return nil, err
		}
	}
	return nil, nil
}
