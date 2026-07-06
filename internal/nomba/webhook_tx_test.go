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

func TestIsTransferPayment_onlineCheckoutBankTransfer(t *testing.T) {
	tx := WebhookTransaction{
		Type:          "online_checkout",
		MerchantTxRef: "100004260706175520164629180694",
	}
	order := &WebhookOrder{PaymentMethod: "bank_transfer"}
	require.True(t, IsTransferPayment(tx, order))
	require.False(t, IsTransferPayment(tx, nil))
}

func TestCheckoutOrderReference_prefersOrderField(t *testing.T) {
	tx := WebhookTransaction{MerchantTxRef: "nomba-internal-ref"}
	order := &WebhookOrder{OrderReference: "e50530f0-2a12-44b8-b3c1-40c9e654bff3"}
	require.Equal(t, "e50530f0-2a12-44b8-b3c1-40c9e654bff3", CheckoutOrderReference(tx, order))
}

func TestEffectiveTokenKey_ignoresPlaceholder(t *testing.T) {
	tx := WebhookTransaction{}
	tokenized := &WebhookTokenizedCardData{TokenKey: "N/A"}
	require.Equal(t, "", EffectiveTokenKey(tx, tokenized))
	require.True(t, IsPlaceholderToken("N/A"))
}

func TestCardDetailsFromWebhook_onlineCheckoutCard(t *testing.T) {
	tx := WebhookTransaction{CardIssuer: "Visa"}
	order := &WebhookOrder{CardLast4Digits: "6424", CardType: "Visa"}
	tokenized := &WebhookTokenizedCardData{
		TokenKey: "5844618949",
		CardType: "Visa",
		CardPan:  "492069**** ****6424",
	}
	customer := &WebhookCustomer{BillerID: "492069**** ****6424"}

	last4, brand := CardDetailsFromWebhook(tx, order, tokenized, customer)
	require.Equal(t, "6424", last4)
	require.Equal(t, "Visa", brand)
}

func TestCardDetailsFromWebhook_fromPANOnly(t *testing.T) {
	tokenized := &WebhookTokenizedCardData{CardPan: "492069**** ****6424", CardType: "N/A"}
	last4, brand := CardDetailsFromWebhook(WebhookTransaction{}, nil, tokenized, nil)
	require.Equal(t, "6424", last4)
	require.Equal(t, "", brand)
}

func TestParsePaymentMethods(t *testing.T) {
	require.Equal(t, []PaymentMethod{PaymentMethodCard}, ParsePaymentMethods([]string{"Card"}))
	require.Equal(t, []PaymentMethod{PaymentMethodCard, PaymentMethodTransfer},
		ParsePaymentMethods([]string{"Card", "Transfer"}))
	require.Nil(t, ParsePaymentMethods(nil))
}
