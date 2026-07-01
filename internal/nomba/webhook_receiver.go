package nomba

import "errors"

// WebhookReceiver handles inbound Nomba webhook HTTP requests.
type WebhookReceiver struct {
	signingKey string
}

func NewWebhookReceiver(signingKey string) *WebhookReceiver {
	return &WebhookReceiver{signingKey: signingKey}
}

// Verify validates the request signature before processing.
func (r *WebhookReceiver) Verify(body []byte, signatureHeader, timestamp string) error {
	if r == nil {
		return errors.New("webhook receiver not configured")
	}
	return VerifyWebhookSignature(body, signatureHeader, r.signingKey, timestamp)
}
