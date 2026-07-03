package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type WebhookService struct {
	repo      domain.WebhookRepository
	tenants   domain.TenantRepository
	publisher TaskPublisher
	cfg       *config.Config
}

func NewWebhookService(repo domain.WebhookRepository, tenants domain.TenantRepository, publisher TaskPublisher, cfg *config.Config) *WebhookService {
	return &WebhookService{repo: repo, tenants: tenants, publisher: publisher, cfg: cfg}
}

type CreateWebhookEndpointInput struct {
	URL      string
	Events   []string
	IsActive bool
}

type UpdateWebhookEndpointInput struct {
	URL      *string
	Events   []string
	IsActive *bool
}

func (s *WebhookService) CreateEndpoint(ctx context.Context, tenantID uuid.UUID, in CreateWebhookEndpointInput) (*domain.WebhookEndpoint, error) {
	if err := validateWebhookURL(in.URL, s.cfg); err != nil {
		return nil, err
	}
	count, err := s.repo.CountEndpoints(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxWebhookEndpointsPerTenant {
		return nil, fmt.Errorf("%w: maximum webhook endpoints reached", domain.ErrValidation)
	}

	ep := &domain.WebhookEndpoint{
		TenantID: tenantID,
		URL:      strings.TrimSpace(in.URL),
		Events:   in.Events,
		IsActive: true,
	}
	if !in.IsActive {
		ep.IsActive = in.IsActive
	}
	if len(ep.Events) == 0 {
		ep.Events = []string{"*"}
	}
	if err := s.repo.CreateEndpoint(ctx, ep); err != nil {
		return nil, err
	}
	return ep, nil
}

func (s *WebhookService) GetEndpoint(ctx context.Context, tenantID, id uuid.UUID) (*domain.WebhookEndpoint, error) {
	return s.repo.GetEndpoint(ctx, tenantID, id)
}

func (s *WebhookService) ListEndpoints(ctx context.Context, tenantID uuid.UUID) ([]*domain.WebhookEndpoint, error) {
	return s.repo.ListEndpoints(ctx, tenantID)
}

func (s *WebhookService) UpdateEndpoint(ctx context.Context, tenantID, id uuid.UUID, in UpdateWebhookEndpointInput) (*domain.WebhookEndpoint, error) {
	ep, err := s.repo.GetEndpoint(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if in.URL != nil {
		if err := validateWebhookURL(*in.URL, s.cfg); err != nil {
			return nil, err
		}
		ep.URL = strings.TrimSpace(*in.URL)
	}
	if in.Events != nil {
		ep.Events = in.Events
	}
	if in.IsActive != nil {
		ep.IsActive = *in.IsActive
	}
	ep.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateEndpoint(ctx, ep); err != nil {
		return nil, err
	}
	return ep, nil
}

func (s *WebhookService) DeleteEndpoint(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.DeleteEndpoint(ctx, tenantID, id)
}

func (s *WebhookService) ListDeliveries(ctx context.Context, tenantID, endpointID uuid.UUID) ([]*domain.WebhookDelivery, error) {
	return s.repo.ListDeliveries(ctx, tenantID, endpointID)
}

func (s *WebhookService) Emit(ctx context.Context, tenantID uuid.UUID, eventType string, data map[string]any) error {
	endpoints, err := s.repo.ListEndpoints(ctx, tenantID)
	if err != nil {
		return err
	}

	envelope := map[string]any{
		"id":         uuid.New().String(),
		"type":       eventType,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"data":       data,
	}

	for _, ep := range endpoints {
		if !ep.IsActive || !subscribed(ep.Events, eventType) {
			continue
		}
		delivery := &domain.WebhookDelivery{
			TenantID:   tenantID,
			EndpointID: ep.ID,
			EventType:  eventType,
			Payload:    envelope,
		}
		if err := s.repo.CreateDelivery(ctx, delivery); err != nil {
			continue
		}
		s.enqueueDelivery(ctx, delivery.ID, tenantID)
	}
	return nil
}

func (s *WebhookService) enqueueDelivery(ctx context.Context, deliveryID, tenantID uuid.UUID) {
	if s.publisher == nil {
		return
	}
	raw, _ := json.Marshal(webhookDeliverPayload{
		DeliveryID: deliveryID.String(),
		TenantID:   tenantID.String(),
	})
	task := asynq.NewTask(TaskWebhookDeliver, raw, asynq.MaxRetry(5))
	_, _ = s.publisher.EnqueueContext(ctx, task)
}

func subscribed(events []string, eventType string) bool {
	for _, e := range events {
		if e == "*" || e == eventType {
			return true
		}
	}
	return false
}

func validateWebhookURL(raw string, cfg *config.Config) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%w: invalid webhook url", domain.ErrValidation)
	}
	if cfg != nil && !cfg.IsDevelopment() && u.Scheme != "https" {
		return fmt.Errorf("%w: webhook url must use https in production", domain.ErrValidation)
	}
	return nil
}

const TaskWebhookDeliver = "webhook:deliver"

type webhookDeliverPayload struct {
	DeliveryID string `json:"delivery_id"`
	TenantID   string `json:"tenant_id"`
}
