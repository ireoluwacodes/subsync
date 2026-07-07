package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/service"
)

type CreatePlanRequest struct {
	Name         string              `json:"name" binding:"required"`
	Description  string              `json:"description"`
	Amount       int64               `json:"amount" binding:"required"`
	Currency     string              `json:"currency"`
	Interval     domain.PlanInterval `json:"interval" binding:"required"`
	IntervalDays *int                `json:"interval_days"`
	TrialDays    int                 `json:"trial_days"`
	Features     []string            `json:"features"`
	IsActive     *bool               `json:"is_active"`
}

func (r CreatePlanRequest) ToInput() service.CreatePlanInput {
	active := true
	if r.IsActive != nil {
		active = *r.IsActive
	}
	return service.CreatePlanInput{
		Name:         r.Name,
		Description:  r.Description,
		Amount:       r.Amount,
		Currency:     r.Currency,
		Interval:     r.Interval,
		IntervalDays: r.IntervalDays,
		TrialDays:    r.TrialDays,
		Features:     r.Features,
		IsActive:     active,
	}
}

type UpdatePlanRequest struct {
	Name         string              `json:"name" binding:"required"`
	Description  string              `json:"description"`
	Amount       int64               `json:"amount" binding:"required"`
	Currency     string              `json:"currency" binding:"required"`
	Interval     domain.PlanInterval `json:"interval" binding:"required"`
	IntervalDays *int                `json:"interval_days"`
	TrialDays    int                 `json:"trial_days"`
	Features     []string            `json:"features"`
	IsActive     bool                `json:"is_active"`
}

func (r UpdatePlanRequest) ToInput() service.UpdatePlanInput {
	return service.UpdatePlanInput{
		Name:         r.Name,
		Description:  r.Description,
		Amount:       r.Amount,
		Currency:     r.Currency,
		Interval:     r.Interval,
		IntervalDays: r.IntervalDays,
		TrialDays:    r.TrialDays,
		Features:     r.Features,
		IsActive:     r.IsActive,
	}
}

type CreateCustomerRequest struct {
	ExternalID string         `json:"external_id"`
	Email      string         `json:"email" binding:"required,email"`
	Name       string         `json:"name"`
	Phone      string         `json:"phone"`
	Metadata   map[string]any `json:"metadata"`
}

func (r CreateCustomerRequest) ToInput() service.CreateCustomerInput {
	return service.CreateCustomerInput(r)
}

type UpdateCustomerRequest struct {
	ExternalID string         `json:"external_id"`
	Email      string         `json:"email" binding:"required,email"`
	Name       string         `json:"name"`
	Phone      string         `json:"phone"`
	Metadata   map[string]any `json:"metadata"`
}

func (r UpdateCustomerRequest) ToInput() service.UpdateCustomerInput {
	return service.UpdateCustomerInput(r)
}

type CreatePaymentMethodRequest struct {
	CustomerID string                  `json:"customer_id" binding:"required"`
	Type       domain.PaymentMethodType `json:"type" binding:"required"`
	TokenKey   string                  `json:"token_key"`
	MandateID  string                  `json:"mandate_id"`
	CardLast4  string                  `json:"card_last4"`
	CardBrand  string                  `json:"card_brand"`
	CardExpiry string                  `json:"card_expiry"`
	IsDefault  bool                    `json:"is_default"`
}

func (r CreatePaymentMethodRequest) ToInput() (service.CreatePaymentMethodInput, error) {
	customerID, err := uuid.Parse(r.CustomerID)
	if err != nil {
		return service.CreatePaymentMethodInput{}, fmt.Errorf("%w: invalid customer_id", domain.ErrValidation)
	}
	return service.CreatePaymentMethodInput{
		CustomerID: customerID,
		Type:       r.Type,
		TokenKey:   r.TokenKey,
		MandateID:  r.MandateID,
		CardLast4:  r.CardLast4,
		CardBrand:  r.CardBrand,
		CardExpiry: r.CardExpiry,
		IsDefault:  r.IsDefault,
	}, nil
}

// Response DTOs

type NombaSettingsResponse struct {
	WebhookURL                   string `json:"webhook_url"`
	NombaClientID                string `json:"nomba_client_id,omitempty"`
	NombaAccountID               string `json:"nomba_account_id"`
	NombaSubAccountID            string `json:"nomba_sub_account_id,omitempty"`
	NombaEnv                     string `json:"nomba_env"`
	NombaWebhookSecretConfigured bool   `json:"nomba_webhook_secret_configured"`
}

type NombaOnboardingResponse struct {
	WebhookURL                   string `json:"webhook_url"`
	NombaWebhookSecretConfigured bool   `json:"nomba_webhook_secret_configured"`
}

func NombaOnboardingFromTenant(t *domain.Tenant, webhookURL string) NombaOnboardingResponse {
	return NombaOnboardingResponse{
		WebhookURL:                   webhookURL,
		NombaWebhookSecretConfigured: t.HasNombaWebhookSecret,
	}
}

type TenantResponse struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Email             string         `json:"email"`
	Website           string         `json:"website,omitempty"`
	NombaClientID     string         `json:"nomba_client_id,omitempty"`
	NombaAccountID    string         `json:"nomba_account_id"`
	NombaSubAccountID string         `json:"nomba_sub_account_id,omitempty"`
	NombaEnv          string         `json:"nomba_env"`
	WebhookSecret     string         `json:"webhook_secret,omitempty"`
	DunningConfig     map[string]any `json:"dunning_config,omitempty"`
	Branding          map[string]any `json:"branding,omitempty"`
	BillingEmail      map[string]any `json:"billing_email,omitempty"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
}

func TenantToResponse(t *domain.Tenant, includeWebhookSecret bool) TenantResponse {
	resp := TenantResponse{
		ID:                t.ID.String(),
		Name:              t.Name,
		Email:             t.Email,
		Website:           t.Website,
		NombaClientID:     maskClientID(t.NombaClientID),
		NombaAccountID:    t.NombaAccountID,
		NombaSubAccountID: t.NombaSubAccountID,
		NombaEnv:          t.NombaEnv,
		DunningConfig:     t.DunningConfig,
		Branding:          t.Branding,
		BillingEmail:      t.BillingEmail,
		CreatedAt:         t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if includeWebhookSecret {
		resp.WebhookSecret = t.WebhookSecret
	}
	return resp
}

type UserResponse struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func UserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		TenantID:  u.TenantID.String(),
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type SubscriptionResponse struct {
	ID                      string  `json:"id"`
	TenantID                string  `json:"tenant_id"`
	CustomerID              string  `json:"customer_id"`
	PlanID                  string  `json:"plan_id"`
	PaymentMethodID         *string `json:"payment_method_id,omitempty"`
	FallbackPaymentMethodID *string `json:"fallback_payment_method_id,omitempty"`
	State                   string  `json:"state"`
	CurrentPeriodStart      string  `json:"current_period_start"`
	CurrentPeriodEnd        string  `json:"current_period_end"`
	CancelAtPeriodEnd       bool    `json:"cancel_at_period_end"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`

	Customer              *CustomerResponse   `json:"customer,omitempty"`
	Plan                  *PlanResponse       `json:"plan,omitempty"`
	PaymentMethod         *PaymentMethodBrief `json:"payment_method,omitempty"`
	FallbackPaymentMethod *PaymentMethodBrief `json:"fallback_payment_method,omitempty"`
}

func SubscriptionToResponse(s *domain.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:                 s.ID.String(),
		TenantID:           s.TenantID.String(),
		CustomerID:         s.CustomerID.String(),
		PlanID:             s.PlanID.String(),
		State:              string(s.State),
		CurrentPeriodStart: s.CurrentPeriodStart.Format(time.RFC3339),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Format(time.RFC3339),
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CreatedAt:          s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          s.UpdatedAt.Format(time.RFC3339),
	}
	if s.PaymentMethodID != nil {
		id := s.PaymentMethodID.String()
		resp.PaymentMethodID = &id
	}
	if s.FallbackPaymentMethodID != nil {
		id := s.FallbackPaymentMethodID.String()
		resp.FallbackPaymentMethodID = &id
	}
	return resp
}

// SubscriptionToResponseWithRelations builds a subscription response with nested
// customer, plan, and payment method objects when they are available.
func SubscriptionToResponseWithRelations(s *domain.Subscription, customer *domain.Customer, plan *domain.Plan, pm, fallbackPM *domain.PaymentMethod) SubscriptionResponse {
	resp := SubscriptionToResponse(s)
	if customer != nil {
		c := CustomerToResponse(customer)
		resp.Customer = &c
	}
	if plan != nil {
		p := PlanToResponse(plan)
		resp.Plan = &p
	}
	if pm != nil {
		m := PaymentMethodToBrief(pm)
		resp.PaymentMethod = &m
	}
	if fallbackPM != nil {
		m := PaymentMethodToBrief(fallbackPM)
		resp.FallbackPaymentMethod = &m
	}
	return resp
}

func PlansToResponse(plans []*domain.Plan) []PlanResponse {
	out := make([]PlanResponse, len(plans))
	for i, p := range plans {
		out[i] = PlanToResponse(p)
	}
	return out
}

type CustomerResponse struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	ExternalID string         `json:"external_id,omitempty"`
	Email      string         `json:"email"`
	Name       string         `json:"name,omitempty"`
	Phone      string         `json:"phone,omitempty"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`

	Subscriptions  []SubscriptionResponse `json:"subscriptions,omitempty"`
	PaymentMethods []PaymentMethodBrief   `json:"payment_methods,omitempty"`
}

func CustomerToResponse(c *domain.Customer) CustomerResponse {
	return CustomerResponse{
		ID:         c.ID.String(),
		TenantID:   c.TenantID.String(),
		ExternalID: c.ExternalID,
		Email:      c.Email,
		Name:       c.Name,
		Phone:      c.Phone,
		Metadata:   c.Metadata,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CustomerToResponseWithRelations builds a customer response with nested
// subscriptions and payment methods when they are available.
func CustomerToResponseWithRelations(c *domain.Customer, subs []*domain.Subscription, pms []*domain.PaymentMethod) CustomerResponse {
	resp := CustomerToResponse(c)
	if len(subs) > 0 {
		resp.Subscriptions = make([]SubscriptionResponse, len(subs))
		for i, sub := range subs {
			resp.Subscriptions[i] = SubscriptionToResponse(sub)
		}
	}
	if len(pms) > 0 {
		resp.PaymentMethods = make([]PaymentMethodBrief, len(pms))
		for i, pm := range pms {
			resp.PaymentMethods[i] = PaymentMethodToBrief(pm)
		}
	}
	return resp
}

func CustomersToResponse(customers []*domain.Customer) []CustomerResponse {
	out := make([]CustomerResponse, len(customers))
	for i, c := range customers {
		out[i] = CustomerToResponse(c)
	}
	return out
}

type PaymentMethodResponse struct {
	ID         string                  `json:"id"`
	TenantID   string                  `json:"tenant_id"`
	CustomerID string                  `json:"customer_id"`
	Type       domain.PaymentMethodType `json:"type"`
	TokenKey   string                  `json:"token_key,omitempty"`
	MandateID  string                  `json:"mandate_id,omitempty"`
	CardLast4  string                  `json:"card_last4,omitempty"`
	CardBrand  string                  `json:"card_brand,omitempty"`
	CardExpiry string                  `json:"card_expiry,omitempty"`
	IsDefault  bool                    `json:"is_default"`
	CreatedAt  string                  `json:"created_at"`
	UpdatedAt  string                  `json:"updated_at"`
}

func PaymentMethodToResponse(pm *domain.PaymentMethod) PaymentMethodResponse {
	return PaymentMethodResponse{
		ID:         pm.ID.String(),
		TenantID:   pm.TenantID.String(),
		CustomerID: pm.CustomerID.String(),
		Type:       pm.Type,
		TokenKey:   pm.TokenKey,
		MandateID:  pm.MandateID,
		CardLast4:  pm.CardLast4,
		CardBrand:  pm.CardBrand,
		CardExpiry: pm.CardExpiry,
		IsDefault:  pm.IsDefault,
		CreatedAt:  pm.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  pm.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// PaymentMethodBrief is a non-sensitive payment method view for embedding in
// other responses. It omits token_key and mandate_id.
type PaymentMethodBrief struct {
	ID            string                   `json:"id"`
	TenantID      string                   `json:"tenant_id"`
	CustomerID    string                   `json:"customer_id"`
	Type          domain.PaymentMethodType `json:"type"`
	MandateStatus domain.MandateStatus     `json:"mandate_status,omitempty"`
	CardLast4     string                   `json:"card_last4,omitempty"`
	CardBrand     string                   `json:"card_brand,omitempty"`
	CardExpiry    string                   `json:"card_expiry,omitempty"`
	IsDefault     bool                     `json:"is_default"`
	CreatedAt     string                   `json:"created_at"`
	UpdatedAt     string                   `json:"updated_at"`
}

func PaymentMethodToBrief(pm *domain.PaymentMethod) PaymentMethodBrief {
	return PaymentMethodBrief{
		ID:            pm.ID.String(),
		TenantID:      pm.TenantID.String(),
		CustomerID:    pm.CustomerID.String(),
		Type:          pm.Type,
		MandateStatus: pm.MandateStatus,
		CardLast4:     pm.CardLast4,
		CardBrand:     pm.CardBrand,
		CardExpiry:    pm.CardExpiry,
		IsDefault:     pm.IsDefault,
		CreatedAt:     pm.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     pm.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func maskClientID(clientID string) string {
	if len(clientID) <= 8 {
		if clientID == "" {
			return ""
		}
		return "****"
	}
	return clientID[:4] + "****" + clientID[len(clientID)-4:]
}
