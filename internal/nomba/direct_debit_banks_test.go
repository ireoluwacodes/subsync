package nomba

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterBanksForDirectDebit(t *testing.T) {
	banks := []Bank{
		{Code: "058", Name: "Guaranty Trust Bank"},
		{Code: "110005", Name: "3LINE CARD MANAGEMENT LIMITED"},
		{Code: "044", Name: "Access Bank"},
	}
	filtered := FilterBanksForDirectDebit(banks)
	require.Len(t, filtered, 2)
	require.Equal(t, "058", filtered[0].Code)
	require.Equal(t, "044", filtered[1].Code)
}

func TestBankSupportsDirectDebit(t *testing.T) {
	require.True(t, BankSupportsDirectDebit("057"))  // Zenith
	require.True(t, BankSupportsDirectDebit("030"))  // Heritage
	require.True(t, BankSupportsDirectDebit("301"))  // Jaiz
	require.True(t, BankSupportsDirectDebit("076"))  // Polaris
	require.True(t, BankSupportsDirectDebit("102"))  // Titan Trust
	require.True(t, BankSupportsDirectDebit("105"))  // Premium Trust
	require.False(t, BankSupportsDirectDebit("110005"))
}

func TestNibssDirectDebitBankCodes_CoversSupportedCommercialBanks(t *testing.T) {
	want := []string{
		"044", "023", "050", "070", "011", "214", "103", "058", "030", "301",
		"082", "076", "105", "101", "221", "068", "232", "100", "102", "032",
		"033", "215", "035", "057",
	}
	for _, code := range want {
		require.True(t, BankSupportsDirectDebit(code), "missing bank code %s", code)
	}
	require.Len(t, nibssDirectDebitBankCodes, len(want)+1) // +1 for Access Diamond 063
}
