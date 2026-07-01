package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func NombaWebhookURL(publicBaseURL string, tenantID uuid.UUID) string {
	base := strings.TrimRight(publicBaseURL, "/")
	return fmt.Sprintf("%s/webhooks/nomba/%s", base, tenantID.String())
}
