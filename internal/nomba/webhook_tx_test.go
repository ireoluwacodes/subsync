package nomba

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsTransferTransaction(t *testing.T) {
	require.True(t, IsTransferTransaction(WebhookTransaction{
		Type:                  "vact_transfer",
		AliasAccountReference: "ref-1",
	}))
	require.True(t, IsTransferTransaction(WebhookTransaction{
		Type:      "transfer",
		SessionID: "sess-1",
	}))
	require.False(t, IsTransferTransaction(WebhookTransaction{
		Type:     "purchase",
		TokenKey: "tok_abc",
	}))
}

func TestParsePaymentMethods(t *testing.T) {
	require.Equal(t, []PaymentMethod{PaymentMethodCard}, ParsePaymentMethods([]string{"Card"}))
	require.Equal(t, []PaymentMethod{PaymentMethodCard, PaymentMethodTransfer},
		ParsePaymentMethods([]string{"Card", "Transfer"}))
	require.Nil(t, ParsePaymentMethods(nil))
}
