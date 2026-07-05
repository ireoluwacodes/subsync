package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/stretchr/testify/require"
)

func TestCheckoutAllowedPaymentMethods_DefaultCardOnly(t *testing.T) {
	methods := checkoutAllowedPaymentMethods(SubscriptionCheckoutInput{})
	require.Equal(t, []nomba.PaymentMethod{nomba.PaymentMethodCard}, methods)
}

func TestCheckoutAllowedPaymentMethods_AllowBankTransfer(t *testing.T) {
	methods := checkoutAllowedPaymentMethods(SubscriptionCheckoutInput{AllowBankTransfer: true})
	require.Equal(t, []nomba.PaymentMethod{nomba.PaymentMethodCard, nomba.PaymentMethodTransfer}, methods)
}

func TestCheckoutAllowedPaymentMethods_ExplicitList(t *testing.T) {
	methods := checkoutAllowedPaymentMethods(SubscriptionCheckoutInput{
		AllowedPaymentMethods: []string{"Transfer"},
	})
	require.Equal(t, []nomba.PaymentMethod{nomba.PaymentMethodTransfer}, methods)
}

func TestPMRemindersDue(t *testing.T) {
	sub := &domain.Subscription{ID: uuid.New(), Metadata: map[string]any{}}

	due := pmRemindersDue(sub, 8*24*time.Hour)
	require.Empty(t, due)

	due = pmRemindersDue(sub, 7*24*time.Hour)
	require.Equal(t, []string{domain.SubscriptionMetaPMReminder7dSent}, due)

	markPMReminderSent(sub, domain.SubscriptionMetaPMReminder7dSent)
	due = pmRemindersDue(sub, 2*24*time.Hour)
	require.Equal(t, []string{domain.SubscriptionMetaPMReminder3dSent}, due)

	markPMReminderSent(sub, domain.SubscriptionMetaPMReminder3dSent)
	due = pmRemindersDue(sub, 12*time.Hour)
	require.Equal(t, []string{domain.SubscriptionMetaPMReminder1dSent}, due)
}
