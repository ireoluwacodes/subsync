package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/utils"
)

type webhookDeliverJobPayload struct {
	DeliveryID string `json:"delivery_id"`
	TenantID   string `json:"tenant_id"`
}

func (h *Handlers) handleWebhookDeliver(ctx context.Context, raw []byte) error {
	var p webhookDeliverJobPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	deliveryID, err := uuid.Parse(p.DeliveryID)
	if err != nil {
		return err
	}
	tenantID, err := uuid.Parse(p.TenantID)
	if err != nil {
		return err
	}
	return h.deliverWebhook(ctx, tenantID, deliveryID)
}

func (h *Handlers) deliverWebhook(ctx context.Context, tenantID, deliveryID uuid.UUID) error {
	if h.Repos == nil {
		return fmt.Errorf("repos not configured")
	}

	delivery, err := h.Repos.Webhooks.GetDelivery(ctx, tenantID, deliveryID)
	if err != nil {
		return err
	}
	if delivery.DeliveredAt != nil {
		return nil
	}

	endpoint, err := h.Repos.Webhooks.GetEndpoint(ctx, tenantID, delivery.EndpointID)
	if err != nil {
		return err
	}

	tenant, err := h.Repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}

	secret := tenant.WebhookSecret
	if h.Config != nil && h.Config.WebhookSigningSecret != "" {
		secret = h.Config.WebhookSigningSecret
	}

	body, err := json.Marshal(delivery.Payload)
	if err != nil {
		return err
	}

	ts := utils.OutboundWebhookTimestamp()
	sig := utils.SignOutboundWebhook(secret, ts, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("SubSync-Signature", sig)
	req.Header.Set("SubSync-Timestamp", utils.OutboundWebhookTimestampHeader(ts))
	req.Header.Set("SubSync-Event", delivery.EventType)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	delivery.AttemptCount++
	if err != nil {
		delivery.LastError = err.Error()
		delivery.NextRetryAt = utils.PtrTime(time.Now().UTC().Add(backoff(delivery.AttemptCount)))
		_ = h.Repos.Webhooks.UpdateDelivery(ctx, delivery)
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	status := resp.StatusCode
	delivery.LastStatus = &status
	if status >= 200 && status < 300 {
		now := time.Now().UTC()
		delivery.DeliveredAt = &now
		delivery.NextRetryAt = nil
		delivery.LastError = ""
		return h.Repos.Webhooks.UpdateDelivery(ctx, delivery)
	}

	delivery.LastError = fmt.Sprintf("unexpected status %d", status)
	delivery.NextRetryAt = utils.PtrTime(time.Now().UTC().Add(backoff(delivery.AttemptCount)))
	_ = h.Repos.Webhooks.UpdateDelivery(ctx, delivery)
	return fmt.Errorf("webhook delivery failed: status %d", status)
}

func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	secs := 1 << min(attempt, 6)
	return time.Duration(secs) * time.Minute
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

