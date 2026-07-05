package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

const (
	portalTokenDefaultTTL = 72 * time.Hour
	portalCheckoutAmount  = 100.0
)

type PortalService struct {
	clock         clock.Clock
	repos         *db.Repos
	subs          *SubscriptionService
	paymentMethod *PaymentMethodService
	nomba         *nomba.Client
	cfg           *config.Config
	publicBaseURL string
	webhooks      *WebhookService
}

func NewPortalService(
	clk clock.Clock,
	repos *db.Repos,
	subs *SubscriptionService,
	paymentMethods *PaymentMethodService,
	nombaClient *nomba.Client,
	cfg *config.Config,
	publicBaseURL string,
	webhooks *WebhookService,
) *PortalService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &PortalService{
		clock:         clk,
		repos:         repos,
		subs:          subs,
		paymentMethod: paymentMethods,
		nomba:         nombaClient,
		cfg:           cfg,
		publicBaseURL: publicBaseURL,
		webhooks:      webhooks,
	}
}

type CreatePortalTokenInput struct {
	SubscriptionID   uuid.UUID
	ExpiresInHours   int
}

type CreatePortalTokenResult struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

func (s *PortalService) CreateToken(ctx context.Context, tenantID uuid.UUID, in CreatePortalTokenInput) (*CreatePortalTokenResult, error) {
	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, in.SubscriptionID)
	if err != nil {
		return nil, err
	}

	ttl := portalTokenDefaultTTL
	if in.ExpiresInHours > 0 {
		ttl = time.Duration(in.ExpiresInHours) * time.Hour
	}

	raw, err := utils.GeneratePortalToken()
	if err != nil {
		return nil, err
	}

	token := &domain.PortalToken{
		TenantID:       tenantID,
		SubscriptionID: sub.ID,
		CustomerID:     sub.CustomerID,
		TokenHash:      utils.HashResetSecret(raw),
		ExpiresAt:      s.clock.Now().UTC().Add(ttl),
	}
	if err := s.repos.PortalTokens.Create(ctx, token); err != nil {
		return nil, err
	}

	base := strings.TrimRight(s.publicBaseURL, "/")
	return &CreatePortalTokenResult{
		Token: raw,
		URL:   fmt.Sprintf("%s/portal/%s", base, raw),
	}, nil
}

// CreatePaymentMethodCaptureLink returns a customer portal URL where they can add a card for renewals.
func (s *PortalService) CreatePaymentMethodCaptureLink(ctx context.Context, tenantID, subscriptionID uuid.UUID) (string, error) {
	result, err := s.CreateToken(ctx, tenantID, CreatePortalTokenInput{SubscriptionID: subscriptionID})
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

type PortalHomeView struct {
	Subscription       *domain.Subscription `json:"subscription"`
	PlanName           string               `json:"plan_name"`
	CustomerEmail      string               `json:"customer_email"`
	CancelAtPeriodEnd  bool                 `json:"cancel_at_period_end"`
	LatestInvoice      *domain.Invoice      `json:"latest_invoice,omitempty"`
	PaymentMethodLast4 string               `json:"payment_method_last4,omitempty"`
	PaymentMethodBrand string               `json:"payment_method_brand,omitempty"`
}

func (s *PortalService) Home(ctx context.Context, rawToken string) (*PortalHomeView, error) {
	pt, err := s.resolveToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	sub, err := s.repos.Subscriptions.GetByID(ctx, pt.TenantID, pt.SubscriptionID)
	if err != nil {
		return nil, err
	}
	plan, err := s.repos.Plans.GetByID(ctx, pt.TenantID, sub.PlanID)
	if err != nil {
		return nil, err
	}
	customer, err := s.repos.Customers.GetByID(ctx, pt.TenantID, sub.CustomerID)
	if err != nil {
		return nil, err
	}

	view := &PortalHomeView{
		Subscription:      sub,
		PlanName:          plan.Name,
		CustomerEmail:     customer.Email,
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
	}

	inv, err := s.repos.Invoices.GetOpenBySubscription(ctx, pt.TenantID, sub.ID)
	if err == nil {
		view.LatestInvoice = inv
	}

	pm, err := s.repos.PaymentMethods.GetDefaultForCustomer(ctx, pt.TenantID, sub.CustomerID)
	if err == nil {
		view.PaymentMethodLast4 = pm.CardLast4
		view.PaymentMethodBrand = pm.CardBrand
	}

	return view, nil
}

type PortalCancelInput struct {
	CancelAtPeriodEnd *bool `json:"cancel_at_period_end"`
}

func (s *PortalService) Cancel(ctx context.Context, rawToken string, in PortalCancelInput) (*domain.Subscription, error) {
	pt, err := s.resolveToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	cancelAtEnd := true
	if in.CancelAtPeriodEnd != nil {
		cancelAtEnd = *in.CancelAtPeriodEnd
	}

	sub, err := s.subs.Cancel(ctx, pt.TenantID, pt.SubscriptionID, CancelInput{
		CancelAtPeriodEnd: cancelAtEnd,
		Reason:            "customer_portal",
	}, "customer")
	if err != nil {
		return nil, err
	}

	if s.webhooks != nil {
		event := domain.WebhookEventSubscriptionUpdated
		if !cancelAtEnd {
			event = domain.WebhookEventSubscriptionCanceled
		}
		_ = s.webhooks.Emit(ctx, pt.TenantID, event, map[string]any{
			"id":    sub.ID.String(),
			"state": string(sub.State),
		})
	}
	return sub, nil
}

type PortalCheckoutResult struct {
	CheckoutLink   string `json:"checkout_link"`
	OrderReference string `json:"order_reference"`
}

func (s *PortalService) StartPaymentMethodUpdate(ctx context.Context, rawToken string) (*PortalCheckoutResult, error) {
	pt, err := s.resolveToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	tenant, err := s.repos.Tenants.GetByID(ctx, pt.TenantID)
	if err != nil {
		return nil, err
	}
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return nil, err
	}
	customer, err := s.repos.Customers.GetByID(ctx, pt.TenantID, pt.CustomerID)
	if err != nil {
		return nil, err
	}

	orderRef := fmt.Sprintf("portal-%s", pt.ID.String())
	result, err := s.nomba.CreateOrder(ctx, tenant, nomba.CreateOrderRequest{
		TokenizeCard: true,
		Order: nomba.Order{
			OrderReference: orderRef,
			CustomerEmail:  customer.Email,
			Amount:         portalCheckoutAmount,
			Currency:       nomba.CurrencyNGN,
			AccountID:      tenant.NombaOrderAccountID(),
			CallbackURL:    fmt.Sprintf("%s/portal/%s", strings.TrimRight(s.publicBaseURL, "/"), rawToken),
			AllowedPaymentMethods: []nomba.PaymentMethod{
				nomba.PaymentMethodCard,
			},
			OrderMetaData: map[string]string{
				"purpose":         "portal_update_pm",
				"portal_token_id": pt.ID.String(),
				"subscription_id": pt.SubscriptionID.String(),
				"customer_id":     pt.CustomerID.String(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &PortalCheckoutResult{
		CheckoutLink:   result.CheckoutLink,
		OrderReference: orderRef,
	}, nil
}

func (s *PortalService) HandlePaymentSuccess(ctx context.Context, tenantID uuid.UUID, orderRef, tokenKey string, tx nomba.WebhookTransaction) error {
	if !strings.HasPrefix(orderRef, "portal-") {
		return nil
	}
	portalID, err := uuid.Parse(strings.TrimPrefix(orderRef, "portal-"))
	if err != nil {
		return fmt.Errorf("%w: invalid portal order reference", domain.ErrValidation)
	}

	token, err := s.repos.PortalTokens.GetByID(ctx, portalID)
	if err != nil {
		return err
	}
	if token.TenantID != tenantID {
		return domain.ErrNotFound
	}

	if tokenKey == "" {
		tokenKey = tx.TokenKey
	}
	if tokenKey == "" {
		return fmt.Errorf("%w: missing tokenKey for portal payment method update", domain.ErrValidation)
	}

	pm, err := s.paymentMethod.Create(ctx, tenantID, CreatePaymentMethodInput{
		CustomerID: token.CustomerID,
		Type:       domain.PaymentMethodTokenizedCard,
		TokenKey:   tokenKey,
		CardLast4:  "",
		CardBrand:  "",
		IsDefault:  true,
	})
	if err != nil {
		return err
	}

	sub, err := s.repos.Subscriptions.GetByID(ctx, tenantID, token.SubscriptionID)
	if err != nil {
		return err
	}
	sub.PaymentMethodID = &pm.ID
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
	return nil
}

func (s *PortalService) resolveToken(ctx context.Context, rawToken string) (*domain.PortalToken, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, fmt.Errorf("%w: missing portal token", domain.ErrValidation)
	}
	return s.repos.PortalTokens.GetValidByTokenHash(ctx, utils.HashResetSecret(rawToken))
}
