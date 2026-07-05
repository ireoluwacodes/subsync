package service

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

func parsePrefixedSubscriptionID(orderRef, prefix string) (uuid.UUID, bool) {
	if !strings.HasPrefix(orderRef, prefix) {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(strings.TrimPrefix(orderRef, prefix))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func subscriptionMetaBool(sub *domain.Subscription, key string) bool {
	if sub == nil || sub.Metadata == nil {
		return false
	}
	v, ok := sub.Metadata[key].(bool)
	return ok && v
}

func setSubscriptionMeta(sub *domain.Subscription, key string, value any) {
	if sub.Metadata == nil {
		sub.Metadata = map[string]any{}
	}
	if value == nil {
		delete(sub.Metadata, key)
		return
	}
	sub.Metadata[key] = value
}

func subscriptionAwaitingPaymentMethod(sub *domain.Subscription) bool {
	return subscriptionMetaBool(sub, domain.SubscriptionMetaAwaitingPaymentMethod)
}

func pmReminderSent(sub *domain.Subscription, key string) bool {
	return subscriptionMetaBool(sub, key)
}

func markPMReminderSent(sub *domain.Subscription, key string) {
	setSubscriptionMeta(sub, key, true)
}

func clearPMReminderMetadata(sub *domain.Subscription) {
	setSubscriptionMeta(sub, domain.SubscriptionMetaPMReminder7dSent, nil)
	setSubscriptionMeta(sub, domain.SubscriptionMetaPMReminder3dSent, nil)
	setSubscriptionMeta(sub, domain.SubscriptionMetaPMReminder1dSent, nil)
	setSubscriptionMeta(sub, domain.SubscriptionMetaLastPMReminderAt, nil)
}

type pmReminderThreshold struct {
	metaKey  string
	maxHours time.Duration
}

var pmReminderThresholds = []pmReminderThreshold{
	{domain.SubscriptionMetaPMReminder7dSent, 7 * 24 * time.Hour},
	{domain.SubscriptionMetaPMReminder3dSent, 3 * 24 * time.Hour},
	{domain.SubscriptionMetaPMReminder1dSent, 24 * time.Hour},
}

// pmRemindersDue returns metadata keys for scheduled reminders not yet sent.
func pmRemindersDue(sub *domain.Subscription, hoursUntil time.Duration) []string {
	var due []string
	for _, th := range pmReminderThresholds {
		if hoursUntil <= th.maxHours && !pmReminderSent(sub, th.metaKey) {
			due = append(due, th.metaKey)
		}
	}
	return due
}

func ParseCardCaptureSubscriptionID(orderRef string) (uuid.UUID, bool) {
	return parsePrefixedSubscriptionID(orderRef, domain.CardCaptureOrderRefPrefix)
}
