package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionCancelScheduledHTML(t *testing.T) {
	subject, html := SubscriptionCancelScheduledHTML("Acme", "Pro", "7 Jul 2026")
	require.Equal(t, "Your subscription cancellation is confirmed", subject)
	require.Contains(t, html, "Pro")
	require.Contains(t, html, "7 Jul 2026")
	require.Contains(t, html, "scheduled to cancel")
}

func TestSubscriptionCanceledHTML(t *testing.T) {
	_, html := SubscriptionCanceledHTML("Acme", "Pro", "7 Jul 2026", "period_ended")
	require.Contains(t, html, "7 Jul 2026")
	require.Contains(t, html, "has been canceled")

	_, html = SubscriptionCanceledHTML("Acme", "Pro", "", "customer_portal")
	require.NotContains(t, html, "access ended")
	require.True(t, strings.Contains(html, "no longer active"))
}
