package nomba

import (
	"fmt"
	"strings"
	"time"
)

const webhookTimestampSkew = 5 * time.Minute

// ParseWebhookTimestamp parses the RFC3339 nomba-timestamp header.
func ParseWebhookTimestamp(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(raw))
}

// ValidateWebhookTimestamp rejects payloads outside the allowed clock skew.
func ValidateWebhookTimestamp(raw string, now time.Time) error {
	ts, err := ParseWebhookTimestamp(raw)
	if err != nil {
		return fmt.Errorf("invalid nomba-timestamp: %w", err)
	}
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > webhookTimestampSkew {
		return fmt.Errorf("nomba-timestamp outside allowed skew")
	}
	return nil
}
