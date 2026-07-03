package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignOutboundWebhook(t *testing.T) {
	body := []byte(`{"type":"invoice.paid"}`)
	ts := int64(1700000000)
	sig := SignOutboundWebhook("secret", ts, body)
	require.NotEmpty(t, sig)
	require.Equal(t, sig, SignOutboundWebhook("secret", ts, body))
	require.NotEqual(t, sig, SignOutboundWebhook("other", ts, body))
}
