package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type BillingReturnOutcome string

const (
	BillingReturnSuccess BillingReturnOutcome = "success"
	BillingReturnPending BillingReturnOutcome = "pending"
	BillingReturnFailed  BillingReturnOutcome = "failed"
	BillingReturnUnknown BillingReturnOutcome = "unknown"
)

type BillingReturnView struct {
	Outcome          BillingReturnOutcome `json:"outcome"`
	StatusLabel      string               `json:"status_label"`
	StatusMessage    string               `json:"status_message"`
	OrderReference   string               `json:"order_reference,omitempty"`
	OrderID          string               `json:"order_id,omitempty"`
	TenantName       string               `json:"tenant_name,omitempty"`
	PlanName         string               `json:"plan_name,omitempty"`
	SubscriptionID   string               `json:"subscription_id,omitempty"`
	SubscriptionState string              `json:"subscription_state,omitempty"`
	InvoiceStatus    string               `json:"invoice_status,omitempty"`
	AmountDisplay    string               `json:"amount_display,omitempty"`
	Purpose          string               `json:"purpose,omitempty"`
}

type BillingReturnService struct {
	cfg   *config.Config
	repos *db.Repos
	nomba *nomba.Client
}

func NewBillingReturnService(cfg *config.Config, repos *db.Repos, nombaClient *nomba.Client) *BillingReturnService {
	return &BillingReturnService{cfg: cfg, repos: repos, nomba: nombaClient}
}

func (s *BillingReturnService) Resolve(ctx context.Context, orderReference, orderID string) (*BillingReturnView, error) {
	orderRef := strings.TrimSpace(orderReference)
	if orderRef == "" {
		orderRef = strings.TrimSpace(orderID)
	}

	view := &BillingReturnView{
		OrderReference: orderRef,
		OrderID:        strings.TrimSpace(orderID),
		Outcome:        BillingReturnUnknown,
		StatusLabel:    "Payment status unavailable",
		StatusMessage:  "We could not find this checkout. If you just paid, confirmation may take a moment — check your email for a receipt.",
	}

	if orderRef == "" {
		return view, nil
	}

	var (
		inv *domain.Invoice
		sub *domain.Subscription
		err error
	)

	inv, err = s.repos.Invoices.FindByNombaOrderRef(ctx, orderRef)
	if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	if inv != nil {
		sub, err = s.repos.Subscriptions.GetByID(ctx, inv.TenantID, inv.SubscriptionID)
		if err != nil {
			return nil, err
		}
		view.Purpose = invoicePurposeLabel(inv)
	} else {
		sub, err = s.subscriptionFromOrderRef(ctx, orderRef)
		if err != nil {
			return nil, err
		}
		if sub != nil {
			view.Purpose = orderRefPurpose(orderRef)
		}
	}

	if sub == nil {
		return view, nil
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, sub.TenantID)
	if err != nil {
		return nil, err
	}
	plan, err := s.repos.Plans.GetByID(ctx, sub.TenantID, sub.PlanID)
	if err != nil {
		return nil, err
	}

	view.TenantName = tenant.Name
	view.PlanName = plan.Name
	view.SubscriptionID = sub.ID.String()
	view.SubscriptionState = string(sub.State)

	if inv != nil {
		view.InvoiceStatus = string(inv.Status)
		view.AmountDisplay = utils.FormatMoneyDisplay(inv.AmountDue, inv.Currency)
	}

	nombaOK := false
	if s.nomba != nil && orderRef != "" {
		if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err == nil {
			if result, err := s.nomba.VerifyCheckoutTransaction(ctx, tenant, orderRef); err == nil && result.Status {
				nombaOK = true
			}
		}
	}

	outcome, label, message := billingReturnPresentation(inv, sub, nombaOK)
	view.Outcome = outcome
	view.StatusLabel = label
	view.StatusMessage = message
	return view, nil
}

func (s *BillingReturnService) subscriptionFromOrderRef(ctx context.Context, orderRef string) (*domain.Subscription, error) {
	if subID, ok := ParseCheckoutSubscriptionID(orderRef); ok {
		return s.repos.Subscriptions.GetByIDGlobal(ctx, subID)
	}
	if subID, ok := ParseCardCaptureSubscriptionID(orderRef); ok {
		return s.repos.Subscriptions.GetByIDGlobal(ctx, subID)
	}
	return nil, nil
}

func billingReturnPresentation(inv *domain.Invoice, sub *domain.Subscription, nombaOK bool) (BillingReturnOutcome, string, string) {
	if inv != nil {
		switch inv.Status {
		case domain.InvoiceStatusPaid:
			return BillingReturnSuccess, "Payment successful", "Your payment was received. You can close this page and return to the app."
		case domain.InvoiceStatusProcessing:
			return BillingReturnPending, "Payment processing", "Your payment is being confirmed. This page will not update automatically — check your email for confirmation."
		case domain.InvoiceStatusOpen:
			if nombaOK {
				return BillingReturnSuccess, "Payment successful", "Your payment was received. Confirmation may still be processing on our side."
			}
			return BillingReturnPending, "Payment pending", "We have not confirmed your payment yet. If you completed checkout, wait a moment and refresh this page."
		case domain.InvoiceStatusVoid:
			return BillingReturnFailed, "Payment not completed", "This checkout was canceled or expired before payment completed."
		case domain.InvoiceStatusUncollectible:
			return BillingReturnFailed, "Payment failed", "We could not collect payment for this checkout."
		}
	}

	if sub != nil {
		switch sub.State {
		case domain.SubscriptionStateActive:
			return BillingReturnSuccess, "Subscription active", "Your subscription is active. You can close this page and return to the app."
		case domain.SubscriptionStateTrialing:
			return BillingReturnSuccess, "Trial started", "Your card was verified and your trial has started."
		case domain.SubscriptionStateIncomplete:
			return BillingReturnPending, "Checkout incomplete", "Payment is not complete yet. Return to checkout to finish, or refresh after paying."
		case domain.SubscriptionStatePastDue:
			return BillingReturnFailed, "Payment issue", "There is a billing issue on this subscription. Use the billing portal link from your email to update payment."
		}
	}

	if nombaOK {
		return BillingReturnSuccess, "Payment successful", "Nomba confirmed your payment. Subscription details may still be updating."
	}

	return BillingReturnUnknown, "Payment status unavailable", "We could not determine the payment outcome yet. If you just paid, wait a moment and refresh."
}

func invoicePurposeLabel(inv *domain.Invoice) string {
	if inv == nil || inv.Metadata == nil {
		return "checkout"
	}
	switch inv.Metadata["purpose"] {
	case domain.InvoicePurposeCardCapture:
		return "card_capture"
	case domain.InvoicePurposeSubscriptionCheckout:
		return "subscription_checkout"
	default:
		return "checkout"
	}
}

func orderRefPurpose(orderRef string) string {
	if strings.HasPrefix(orderRef, domain.CardCaptureOrderRefPrefix) {
		return "card_capture"
	}
	if strings.HasPrefix(orderRef, domain.CheckoutOrderRefPrefix) {
		return "trial_verification"
	}
	return "checkout"
}

func resolveCheckoutSuccessURL(raw string, cfg *config.Config) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if cfg == nil {
			return "", fmt.Errorf("%w: success_url is required", domain.ErrValidation)
		}
		return cfg.DefaultCheckoutSuccessURL(), nil
	}
	return raw, validateRedirectURL(raw, cfg)
}
