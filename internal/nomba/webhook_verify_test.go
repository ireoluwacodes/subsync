package nomba

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyWebhookSignature_NombaAlgorithm(t *testing.T) {
	secret := "HkatexKDZg7CLWy96q5sfrVHSvtoz92B"
	timestamp := "2025-09-29T10:51:44Z"
	body := []byte(`{
		"event_type": "payment_success",
		"requestId": "45f2dc2d-d559-4773-bba3-2d5ec17b2e20",
		"data": {
			"merchant": {
				"walletId": "6756ff80aafe04a795f18b38",
				"walletBalance": 6052,
				"userId": "b7b10e81-e57d-41d0-8fdc-f4e23a132bbf"
			},
			"terminal": {},
			"transaction": {
				"aliasAccountNumber": "5343270516",
				"fee": 5,
				"sessionId": "IFAP-TRANSFER-46501-e0339485-1a2f-4b43-9bd5-fec9649e5928",
				"type": "vact_transfer",
				"transactionId": "API-VACT_TRA-B7B10-0435b274-807a-4bc7-8abe-9dbb4548fd7a",
				"aliasAccountName": "ZAXBOX/EZENNA NWACHUKWU",
				"responseCode": "",
				"originatingFrom": "api",
				"transactionAmount": 10,
				"narration": "test",
				"time": "2025-09-29T10:51:44Z",
				"aliasAccountReference": "654f7c80bd4a510c90fb7f92",
				"aliasAccountType": "VIRTUAL"
			},
			"customer": {
				"bankCode": "090645",
				"senderName": "Test User",
				"bankName": "Nombank",
				"accountNumber": "9617811496"
			}
		}
	}`)

	expected := "Kt9095hQxfgmVbx6iz7G2tPhHdbdXgLlyY/mf35sptw="
	got, err := GenerateWebhookSignature(body, secret, timestamp)
	require.NoError(t, err)
	require.Equal(t, expected, got)

	require.NoError(t, VerifyWebhookSignature(body, expected, secret, timestamp))
	require.ErrorIs(t, VerifyWebhookSignature(body, "bad", secret, timestamp), ErrInvalidWebhookSignature)
}
