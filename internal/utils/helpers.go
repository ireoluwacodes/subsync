package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func PtrTime(t time.Time) *time.Time {
	return &t
}

func MaskClientID(clientID string) string {
	if len(clientID) <= 8 {
		if clientID == "" {
			return ""
		}
		return strings.Repeat("*", len(clientID))
	}
	return clientID[:4] + strings.Repeat("*", len(clientID)-8) + clientID[len(clientID)-4:]
}

func NombaWebhookURL(publicBaseURL string, tenantID uuid.UUID) string {
	base := strings.TrimRight(publicBaseURL, "/")
	return fmt.Sprintf("%s/webhooks/nomba/%s", base, tenantID.String())
}
