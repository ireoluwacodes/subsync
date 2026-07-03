package models

import (
	"encoding/json"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

func NombaEventToDomain(m *NombaEvent) *domain.NombaEvent {
	ev := &domain.NombaEvent{
		ID:        m.ID,
		TenantID:  m.TenantID,
		EventID:   m.EventID,
		EventType: m.EventType,
		Processed: m.Processed,
		CreatedAt: m.CreatedAt,
	}
	if m.ProcessedAt != nil {
		ev.ProcessedAt = m.ProcessedAt
	}
	if m.Error != nil {
		ev.Error = *m.Error
	}
	_ = json.Unmarshal(m.Payload, &ev.Payload)
	return ev
}

func NombaEventFromDomain(e *domain.NombaEvent) (*NombaEvent, error) {
	raw, err := json.Marshal(e.Payload)
	if err != nil {
		return nil, err
	}
	m := &NombaEvent{
		ID:        e.ID,
		TenantID:  e.TenantID,
		EventID:   e.EventID,
		EventType: e.EventType,
		Payload:   datatypes.JSON(raw),
		Processed: e.Processed,
		CreatedAt: e.CreatedAt,
	}
	if e.ProcessedAt != nil {
		m.ProcessedAt = e.ProcessedAt
	}
	if e.Error != "" {
		m.Error = &e.Error
	}
	return m, nil
}

func WebhookEndpointToDomain(m *WebhookEndpoint) *domain.WebhookEndpoint {
	return &domain.WebhookEndpoint{
		ID:        m.ID,
		TenantID:  m.TenantID,
		URL:       m.URL,
		Events:    []string(m.Events),
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func WebhookEndpointFromDomain(ep *domain.WebhookEndpoint) *WebhookEndpoint {
	return &WebhookEndpoint{
		ID:        ep.ID,
		TenantID:  ep.TenantID,
		URL:       ep.URL,
		Events:    pq.StringArray(ep.Events),
		IsActive:  ep.IsActive,
		CreatedAt: ep.CreatedAt,
		UpdatedAt: ep.UpdatedAt,
	}
}

func WebhookDeliveryToDomain(m *WebhookDelivery) *domain.WebhookDelivery {
	d := &domain.WebhookDelivery{
		ID:           m.ID,
		TenantID:     m.TenantID,
		EndpointID:   m.EndpointID,
		EventType:    m.EventType,
		AttemptCount: m.AttemptCount,
		LastStatus:   m.LastStatus,
		DeliveredAt:  m.DeliveredAt,
		NextRetryAt:  m.NextRetryAt,
		CreatedAt:    m.CreatedAt,
	}
	if m.LastError != nil {
		d.LastError = *m.LastError
	}
	_ = json.Unmarshal(m.Payload, &d.Payload)
	return d
}

func WebhookDeliveryFromDomain(d *domain.WebhookDelivery) (*WebhookDelivery, error) {
	raw, err := json.Marshal(d.Payload)
	if err != nil {
		return nil, err
	}
	m := &WebhookDelivery{
		ID:           d.ID,
		TenantID:     d.TenantID,
		EndpointID:   d.EndpointID,
		EventType:    d.EventType,
		Payload:      datatypes.JSON(raw),
		AttemptCount: d.AttemptCount,
		LastStatus:   d.LastStatus,
		DeliveredAt:  d.DeliveredAt,
		NextRetryAt:  d.NextRetryAt,
		CreatedAt:    d.CreatedAt,
	}
	if d.LastError != "" {
		m.LastError = &d.LastError
	}
	return m, nil
}

func PortalTokenToDomain(m *PortalToken) *domain.PortalToken {
	return &domain.PortalToken{
		ID:             m.ID,
		TenantID:       m.TenantID,
		SubscriptionID: m.SubscriptionID,
		CustomerID:     m.CustomerID,
		TokenHash:      m.TokenHash,
		ExpiresAt:      m.ExpiresAt,
		UsedAt:         m.UsedAt,
		CreatedAt:      m.CreatedAt,
	}
}

func PortalTokenFromDomain(t *domain.PortalToken) *PortalToken {
	return &PortalToken{
		ID:             t.ID,
		TenantID:       t.TenantID,
		SubscriptionID: t.SubscriptionID,
		CustomerID:     t.CustomerID,
		TokenHash:      t.TokenHash,
		ExpiresAt:      t.ExpiresAt,
		UsedAt:         t.UsedAt,
		CreatedAt:      t.CreatedAt,
	}
}
