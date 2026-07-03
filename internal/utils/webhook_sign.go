package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

func SignOutboundWebhook(secret string, timestamp int64, body []byte) string {
	payload := fmt.Sprintf("%d.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func OutboundWebhookTimestamp() int64 {
	return time.Now().UTC().Unix()
}

func OutboundWebhookTimestampHeader(ts int64) string {
	return strconv.FormatInt(ts, 10)
}
