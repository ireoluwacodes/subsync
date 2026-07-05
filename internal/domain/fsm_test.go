package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateTransition(t *testing.T) {
	assert.NoError(t, ValidateTransition(SubscriptionStateTrialing, SubscriptionStateActive))
	assert.NoError(t, ValidateTransition(SubscriptionStateIncomplete, SubscriptionStateActive))
	assert.NoError(t, ValidateTransition(SubscriptionStateIncomplete, SubscriptionStateTrialing))
	assert.NoError(t, ValidateTransition(SubscriptionStateIncomplete, SubscriptionStateCanceled))
	assert.ErrorIs(t, ValidateTransition(SubscriptionStateCanceled, SubscriptionStateActive), ErrInvalidTransition)
}

func TestCalculateProration(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	change := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	result := CalculateProration(10000, 20000, start, end, change)

	assert.Greater(t, result.DaysRemaining, 0)
	assert.Equal(t, result.NetAmount, result.DebitAmount-result.CreditAmount)
}
