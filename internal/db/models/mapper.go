package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

func TenantToDomain(m *Tenant) *domain.Tenant {
	t := &domain.Tenant{
		ID:            m.ID,
		Name:          m.Name,
		Email:         m.Email,
		Website:       m.Website,
		NombaClientID: m.NombaClientID,
		NombaAccountID: m.NombaAccountID,
		NombaEnv:      m.NombaEnv,
		APIKeyPrefix:  m.APIKeyPrefix,
		APIKeyHash:    m.APIKeyHash,
		WebhookSecret: m.WebhookSecret,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.NombaSubAccountID != nil {
		t.NombaSubAccountID = *m.NombaSubAccountID
	}
	t.HasNombaWebhookSecret = m.NombaWebhookSigningKeyEnc != ""
	unmarshalJSON(m.DunningConfig, &t.DunningConfig)
	unmarshalJSON(m.Branding, &t.Branding)
	unmarshalJSON(m.BillingEmail, &t.BillingEmail)
	return t
}

func TenantFromDomain(t *domain.Tenant, clientSecretEnc, webhookSecretEnc string) (*Tenant, error) {
	dunning, err := marshalJSON(t.DunningConfig, "{}")
	if err != nil {
		return nil, err
	}
	branding, err := marshalJSON(t.Branding, "{}")
	if err != nil {
		return nil, err
	}
	billingEmail, err := marshalJSON(t.BillingEmail, "{}")
	if err != nil {
		return nil, err
	}
	m := &Tenant{
		Name:                 t.Name,
		Email:                t.Email,
		Website:              t.Website,
		NombaClientID:            t.NombaClientID,
		NombaClientSecretEnc:     clientSecretEnc,
		NombaWebhookSigningKeyEnc: webhookSecretEnc,
		NombaAccountID:           t.NombaAccountID,
		NombaEnv:             t.NombaEnv,
		APIKeyPrefix:         t.APIKeyPrefix,
		APIKeyHash:           t.APIKeyHash,
		WebhookSecret:        t.WebhookSecret,
		DunningConfig:        dunning,
		Branding:             branding,
		BillingEmail:         billingEmail,
	}
	if t.NombaSubAccountID != "" {
		sub := t.NombaSubAccountID
		m.NombaSubAccountID = &sub
	}
	if t.ID != uuid.Nil {
		m.ID = t.ID
	}
	return m, nil
}

func UserToDomain(m *User) *domain.User {
	return &domain.User{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Name:         m.Name,
		TokenVersion: m.TokenVersion,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func UserFromDomain(u *domain.User) *User {
	m := &User{
		TenantID:     u.TenantID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		TokenVersion: u.TokenVersion,
	}
	if u.ID != uuid.Nil {
		m.ID = u.ID
	}
	return m
}

func SubscriptionToDomain(m *Subscription) *domain.Subscription {
	s := &domain.Subscription{
		ID:                 m.ID,
		TenantID:           m.TenantID,
		CustomerID:         m.CustomerID,
		PlanID:             m.PlanID,
		PaymentMethodID:    m.PaymentMethodID,
		State:              domain.SubscriptionState(m.State),
		TrialEndsAt:        m.TrialEndsAt,
		CurrentPeriodStart: m.CurrentPeriodStart,
		CurrentPeriodEnd:   m.CurrentPeriodEnd,
		NextBillingAt:      m.NextBillingAt,
		CanceledAt:         m.CanceledAt,
		CancelAtPeriodEnd:  m.CancelAtPeriodEnd,
		PauseStartsAt:      m.PauseStartsAt,
		PauseEndsAt:        m.PauseEndsAt,
		DunningStep:        m.DunningStep,
		DunningStartedAt:   m.DunningStartedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
	unmarshalJSON(m.Metadata, &s.Metadata)
	return s
}

func SubscriptionFromDomain(s *domain.Subscription) (*Subscription, error) {
	meta, err := marshalJSON(s.Metadata, "{}")
	if err != nil {
		return nil, err
	}
	m := &Subscription{
		TenantID:           s.TenantID,
		CustomerID:         s.CustomerID,
		PlanID:             s.PlanID,
		PaymentMethodID:    s.PaymentMethodID,
		State:              string(s.State),
		TrialEndsAt:        s.TrialEndsAt,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		NextBillingAt:      s.NextBillingAt,
		CanceledAt:         s.CanceledAt,
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		PauseStartsAt:      s.PauseStartsAt,
		PauseEndsAt:        s.PauseEndsAt,
		DunningStep:        s.DunningStep,
		DunningStartedAt:   s.DunningStartedAt,
		Metadata:           meta,
	}
	if s.ID != uuid.Nil {
		m.ID = s.ID
	}
	return m, nil
}

func SubscriptionTransitionToDomain(m *SubscriptionTransition) *domain.SubscriptionTransition {
	t := &domain.SubscriptionTransition{
		ID:             m.ID,
		SubscriptionID: m.SubscriptionID,
		TenantID:       m.TenantID,
		FromState:      domain.SubscriptionState(m.FromState),
		ToState:        domain.SubscriptionState(m.ToState),
		Reason:         m.Reason,
		Actor:          m.Actor,
		CreatedAt:      m.CreatedAt,
	}
	unmarshalJSON(m.Metadata, &t.Metadata)
	return t
}

func SubscriptionTransitionFromDomain(t *domain.SubscriptionTransition) (*SubscriptionTransition, error) {
	meta, err := marshalJSON(t.Metadata, "{}")
	if err != nil {
		return nil, err
	}
	m := &SubscriptionTransition{
		SubscriptionID: t.SubscriptionID,
		TenantID:       t.TenantID,
		FromState:      string(t.FromState),
		ToState:        string(t.ToState),
		Reason:         t.Reason,
		Actor:          t.Actor,
		Metadata:       meta,
	}
	if t.ID != uuid.Nil {
		m.ID = t.ID
	}
	return m, nil
}

func InvoiceToDomain(m *Invoice) *domain.Invoice {
	inv := &domain.Invoice{
		ID:             m.ID,
		TenantID:       m.TenantID,
		SubscriptionID: m.SubscriptionID,
		CustomerID:     m.CustomerID,
		Status:         domain.InvoiceStatus(m.Status),
		AmountDue:      m.AmountDue,
		AmountPaid:     m.AmountPaid,
		Currency:       m.Currency,
		PeriodStart:    m.PeriodStart,
		PeriodEnd:      m.PeriodEnd,
		DueDate:        m.DueDate,
		PaidAt:         m.PaidAt,
		VoidedAt:       m.VoidedAt,
		AttemptCount:   m.AttemptCount,
		NextAttemptAt:  m.NextAttemptAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.NombaOrderRef != nil {
		inv.NombaOrderRef = *m.NombaOrderRef
	}
	if m.NombaTransactionID != nil {
		inv.NombaTransactionID = *m.NombaTransactionID
	}
	unmarshalJSON(m.Metadata, &inv.Metadata)
	return inv
}

func InvoiceFromDomain(inv *domain.Invoice) (*Invoice, error) {
	meta, err := marshalJSON(inv.Metadata, "{}")
	if err != nil {
		return nil, err
	}
	m := &Invoice{
		TenantID:       inv.TenantID,
		SubscriptionID: inv.SubscriptionID,
		CustomerID:     inv.CustomerID,
		Status:         string(inv.Status),
		AmountDue:      inv.AmountDue,
		AmountPaid:     inv.AmountPaid,
		Currency:       inv.Currency,
		PeriodStart:    inv.PeriodStart,
		PeriodEnd:      inv.PeriodEnd,
		DueDate:        inv.DueDate,
		PaidAt:         inv.PaidAt,
		VoidedAt:       inv.VoidedAt,
		AttemptCount:   inv.AttemptCount,
		NextAttemptAt:  inv.NextAttemptAt,
		Metadata:       meta,
	}
	if inv.NombaOrderRef != "" {
		ref := inv.NombaOrderRef
		m.NombaOrderRef = &ref
	}
	if inv.NombaTransactionID != "" {
		tx := inv.NombaTransactionID
		m.NombaTransactionID = &tx
	}
	if inv.ID != uuid.Nil {
		m.ID = inv.ID
	}
	return m, nil
}

func InvoiceLineItemToDomain(m *InvoiceLineItem) *domain.InvoiceLineItem {
	return &domain.InvoiceLineItem{
		ID:          m.ID,
		InvoiceID:   m.InvoiceID,
		TenantID:    m.TenantID,
		Type:        domain.LineItemType(m.Type),
		Description: m.Description,
		Amount:      m.Amount,
		Currency:    m.Currency,
		PeriodStart: m.PeriodStart,
		PeriodEnd:   m.PeriodEnd,
		CreatedAt:   m.CreatedAt,
	}
}

func InvoiceLineItemFromDomain(item *domain.InvoiceLineItem) *InvoiceLineItem {
	m := &InvoiceLineItem{
		InvoiceID:   item.InvoiceID,
		TenantID:    item.TenantID,
		Type:        string(item.Type),
		Description: item.Description,
		Amount:      item.Amount,
		Currency:    item.Currency,
		PeriodStart: item.PeriodStart,
		PeriodEnd:   item.PeriodEnd,
	}
	if item.ID != uuid.Nil {
		m.ID = item.ID
	}
	return m
}

func PlanToDomain(m *Plan) *domain.Plan {
	p := &domain.Plan{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Name:         m.Name,
		Amount:       m.Amount,
		Currency:     m.Currency,
		Interval:     domain.PlanInterval(m.Interval),
		IntervalDays: m.IntervalDays,
		TrialDays:    m.TrialDays,
		IsActive:     m.IsActive,
		IsArchived:   m.IsArchived,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.Description != nil {
		p.Description = *m.Description
	}
	if len(m.Features) > 0 {
		_ = json.Unmarshal(m.Features, &p.Features)
	}
	return p
}

func PlanFromDomain(p *domain.Plan) (*Plan, error) {
	features, err := marshalJSON(p.Features, "[]")
	if err != nil {
		return nil, err
	}
	m := &Plan{
		TenantID:     p.TenantID,
		Name:         p.Name,
		Amount:       p.Amount,
		Currency:     p.Currency,
		Interval:     string(p.Interval),
		IntervalDays: p.IntervalDays,
		TrialDays:    p.TrialDays,
		Features:     features,
		IsActive:     p.IsActive,
		IsArchived:   p.IsArchived,
	}
	if p.Description != "" {
		desc := p.Description
		m.Description = &desc
	}
	if p.ID != uuid.Nil {
		m.ID = p.ID
	}
	return m, nil
}

func CustomerToDomain(m *Customer) *domain.Customer {
	c := &domain.Customer{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.ExternalID != nil {
		c.ExternalID = *m.ExternalID
	}
	if m.Name != nil {
		c.Name = *m.Name
	}
	if m.Phone != nil {
		c.Phone = *m.Phone
	}
	unmarshalJSON(m.Metadata, &c.Metadata)
	return c
}

func CustomerFromDomain(c *domain.Customer) (*Customer, error) {
	meta, err := marshalJSON(c.Metadata, "{}")
	if err != nil {
		return nil, err
	}
	m := &Customer{
		TenantID: c.TenantID,
		Email:    c.Email,
		Metadata: meta,
	}
	if c.ExternalID != "" {
		ext := c.ExternalID
		m.ExternalID = &ext
	}
	if c.Name != "" {
		name := c.Name
		m.Name = &name
	}
	if c.Phone != "" {
		phone := c.Phone
		m.Phone = &phone
	}
	if c.ID != uuid.Nil {
		m.ID = c.ID
	}
	return m, nil
}

func PaymentMethodToDomain(m *PaymentMethod) *domain.PaymentMethod {
	pm := &domain.PaymentMethod{
		ID:         m.ID,
		TenantID:   m.TenantID,
		CustomerID: m.CustomerID,
		Type:       domain.PaymentMethodType(m.Type),
		IsDefault:  m.IsDefault,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
	if m.TokenKey != nil {
		pm.TokenKey = *m.TokenKey
	}
	if m.MandateID != nil {
		pm.MandateID = *m.MandateID
	}
	if m.CardLast4 != nil {
		pm.CardLast4 = *m.CardLast4
	}
	if m.CardBrand != nil {
		pm.CardBrand = *m.CardBrand
	}
	if m.CardExpiry != nil {
		pm.CardExpiry = *m.CardExpiry
	}
	return pm
}

func PaymentMethodFromDomain(pm *domain.PaymentMethod) *PaymentMethod {
	m := &PaymentMethod{
		TenantID:   pm.TenantID,
		CustomerID: pm.CustomerID,
		Type:       string(pm.Type),
		IsDefault:  pm.IsDefault,
	}
	if pm.TokenKey != "" {
		v := pm.TokenKey
		m.TokenKey = &v
	}
	if pm.MandateID != "" {
		v := pm.MandateID
		m.MandateID = &v
	}
	if pm.CardLast4 != "" {
		v := pm.CardLast4
		m.CardLast4 = &v
	}
	if pm.CardBrand != "" {
		v := pm.CardBrand
		m.CardBrand = &v
	}
	if pm.CardExpiry != "" {
		v := pm.CardExpiry
		m.CardExpiry = &v
	}
	if pm.ID != uuid.Nil {
		m.ID = pm.ID
	}
	return m
}

func marshalJSON(v any, empty string) ([]byte, error) {
	if v == nil {
		return []byte(empty), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return []byte(empty), nil
	}
	return b, nil
}

func unmarshalJSON(data []byte, dest *map[string]any) {
	if len(data) == 0 {
		*dest = map[string]any{}
		return
	}
	_ = json.Unmarshal(data, dest)
	if *dest == nil {
		*dest = map[string]any{}
	}
}
