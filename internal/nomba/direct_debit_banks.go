package nomba

// nibssDirectDebitBankCodes are CBN codes for commercial banks that support NIBSS direct debit.
// Source: Nomba/NIBSS supported mandate banks (commercial banks list).
// Nomba's GET /v1/transfers/banks returns every NIP participant; filter with this set for mandates.
var nibssDirectDebitBankCodes = map[string]struct{}{
	"011": {}, // First Bank of Nigeria
	"023": {}, // Citibank Nigeria
	"030": {}, // Heritage Bank
	"032": {}, // Union Bank of Nigeria
	"033": {}, // United Bank for Africa (UBA)
	"035": {}, // Wema Bank
	"044": {}, // Access Bank
	"050": {}, // Ecobank Nigeria
	"057": {}, // Zenith Bank
	"058": {}, // Guaranty Trust Bank (GTB)
	"063": {}, // Access Bank (Diamond legacy code)
	"068": {}, // Standard Chartered Bank
	"070": {}, // Fidelity Bank
	"076": {}, // Polaris Bank
	"082": {}, // Keystone Bank
	"100": {}, // Suntrust Bank
	"101": {}, // Providus Bank
	"102": {}, // Titan Trust Bank
	"103": {}, // Globus Bank
	"105": {}, // Premium Trust Bank
	"214": {}, // First City Monument Bank (FCMB)
	"215": {}, // Unity Bank
	"221": {}, // Stanbic IBTC Bank
	"232": {}, // Sterling Bank
	"301": {}, // Jaiz Bank
}

// BankSupportsDirectDebit reports whether a bank code can be used for NIBSS mandates.
func BankSupportsDirectDebit(bankCode string) bool {
	_, ok := nibssDirectDebitBankCodes[bankCode]
	return ok
}

// FilterBanksForDirectDebit keeps only banks that support NIBSS e-mandate.
func FilterBanksForDirectDebit(banks []Bank) []Bank {
	if len(banks) == 0 {
		return banks
	}
	out := make([]Bank, 0, len(banks))
	for _, b := range banks {
		if BankSupportsDirectDebit(b.Code) {
			out = append(out, b)
		}
	}
	return out
}
