package portalpage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleNombaMandateText = `To complete your e-mandate activation, please make a token payment of ₦50.00 to the account number provided below. Kindly ensure that the payment is made strictly via your Mobile Banking App or Internet Banking platform.Please ensure the payment is made from the same account used to create the mandate. This token payment serves as your consent for the mandate to be activated on your account. Thank you; Account Number: 9880218357 Bank: Paystack Account Name: NIBSS MANDATE ACTIVATION OR Account Number: 9020025928 Bank: Fidelity Bank Account Name: NIBSS DIRECT DEBIT`

func TestParseMandateInstructions_NombaSample(t *testing.T) {
	view := ParseMandateInstructions(sampleNombaMandateText)
	require.True(t, view.Parsed())
	require.Equal(t, "₦50.00", view.AmountDisplay)
	require.Len(t, view.PaymentOptions, 2)

	require.Equal(t, "9880218357", view.PaymentOptions[0].AccountNumber)
	require.Equal(t, "Paystack", view.PaymentOptions[0].Bank)
	require.Equal(t, "NIBSS MANDATE ACTIVATION", view.PaymentOptions[0].AccountName)
	require.Equal(t, "Option 1", view.PaymentOptions[0].Label)

	require.Equal(t, "9020025928", view.PaymentOptions[1].AccountNumber)
	require.Equal(t, "Fidelity Bank", view.PaymentOptions[1].Bank)
	require.Equal(t, "NIBSS DIRECT DEBIT", view.PaymentOptions[1].AccountName)
}

func TestParseMandateInstructions_FallbackForUnknownFormat(t *testing.T) {
	raw := "Please pay N50 to account 1234567890"
	view := ParseMandateInstructions(raw)
	require.False(t, view.Parsed())
	require.Equal(t, raw, view.RawFallback)
}
