package service

import (
	"testing"
	"time"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionPortalShowCancel(t *testing.T) {
	sub := &domain.Subscription{State: domain.SubscriptionStateActive}
	require.True(t, subscriptionPortalShowCancel(sub))

	sub.CancelAtPeriodEnd = true
	require.False(t, subscriptionPortalShowCancel(sub))

	sub.CancelAtPeriodEnd = false
	sub.State = domain.SubscriptionStateCanceled
	require.False(t, subscriptionPortalShowCancel(sub))
}

func TestSubscriptionPortalCanManagePaymentMethods(t *testing.T) {
	sub := &domain.Subscription{State: domain.SubscriptionStateActive}
	require.True(t, subscriptionPortalCanManagePaymentMethods(sub))

	sub.State = domain.SubscriptionStateCanceled
	require.False(t, subscriptionPortalCanManagePaymentMethods(sub))
}

func TestFormatPortalDate(t *testing.T) {
	tm := time.Date(2026, 7, 7, 20, 51, 16, 0, time.FixedZone("WAT", 3600))
	require.Equal(t, "7 Jul 2026", formatPortalDate(tm))
}
