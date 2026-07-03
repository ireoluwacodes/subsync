package service

import (
	"encoding/json"
	"testing"

	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/stretchr/testify/require"
)

func TestNombaEventService_PaymentSuccessPayload(t *testing.T) {
	var event nomba.WebhookEvent
	require.NoError(t, json.Unmarshal([]byte(`{"event_type":"payment_success","requestId":"x","data":{"merchant":{"userId":"u"},"transaction":{"merchantTxRef":"abc","transactionId":"tx","type":"purchase","time":"2026-01-01T00:00:00Z"}}}`), &event))
	require.Equal(t, nomba.WebhookEventPaymentSuccess, event.EventType)
	require.Equal(t, "abc", event.Data.Transaction.MerchantTxRef)
}
