package nomba

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidWebhookSignature = errors.New("invalid nomba webhook signature")

// GenerateWebhookSignature reproduces Nomba's HMAC-SHA256 + Base64 signature.
// See https://developer.nomba.com/docs/api-basics/webhook.md
func GenerateWebhookSignature(body []byte, secret, timestamp string) (string, error) {
	if secret == "" {
		return "", errors.New("nomba webhook signing key not configured")
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return "", fmt.Errorf("parse webhook payload: %w", err)
	}

	responseCode := event.Data.Transaction.ResponseCode
	if responseCode == "null" {
		responseCode = ""
	}

	hashingPayload := fmt.Sprintf(
		"%s:%s:%s:%s:%s:%s:%s:%s:%s",
		event.EventType,
		event.RequestID,
		event.Data.Merchant.UserID,
		event.Data.Merchant.WalletID,
		event.Data.Transaction.TransactionID,
		event.Data.Transaction.Type,
		event.Data.Transaction.Time,
		responseCode,
		timestamp,
	)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(hashingPayload))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// VerifyWebhookSignature validates the nomba-signature header against the raw body.
// timestamp must come from the nomba-timestamp header.
func VerifyWebhookSignature(body []byte, signatureHeader, secret, timestamp string) error {
	expected, err := GenerateWebhookSignature(body, secret, timestamp)
	if err != nil {
		return err
	}

	received := strings.TrimSpace(signatureHeader)
	if received == "" {
		return ErrInvalidWebhookSignature
	}

	if !strings.EqualFold(received, expected) {
		return ErrInvalidWebhookSignature
	}

	return nil
}
