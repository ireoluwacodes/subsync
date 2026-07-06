package nomba

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMandateRequest_VariableOmitsAmount(t *testing.T) {
	raw, err := json.Marshal(CreateMandateRequest{
		CustomerAccountNumber: "0123456789",
		BankCode:              "058",
		CustomerName:          "Jane Doe",
		CustomerAccountName:   "Jane Doe",
		CustomerAddress:       "Lagos",
		Frequency:             MandateFrequencyVariable,
		MerchantReference:     "12003074001",
		StartDate:             "2026-07-07T00:00",
		EndDate:               "2031-07-07T00:00",
		CustomerEmail:         "jane@example.com",
	})
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"amount"`)
}

func TestCreateMandateRequest_FixedFrequencyIncludesAmount(t *testing.T) {
	amount := 1500.0
	raw, err := json.Marshal(CreateMandateRequest{
		Frequency: MandateFrequencyMonthly,
		Amount:    &amount,
	})
	require.NoError(t, err)
	require.Contains(t, string(raw), `"amount":1500`)
}
