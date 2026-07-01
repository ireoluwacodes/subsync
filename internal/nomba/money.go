package nomba

import "fmt"

// KoboToNombaAmount converts SubSync kobo (int64) to Nomba's decimal NGN amount.
func KoboToNombaAmount(kobo int64) float64 {
	return float64(kobo) / 100.0
}

// KoboToNombaAmountString formats kobo as a decimal string for mandate debit amounts.
func KoboToNombaAmountString(kobo int64) string {
	return fmt.Sprintf("%.2f", KoboToNombaAmount(kobo))
}
