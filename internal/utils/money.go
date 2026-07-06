package utils

import (
	"fmt"
	"strings"
)

// FormatMoneyDisplay formats a minor-unit amount (e.g. kobo) for human-readable UI.
func FormatMoneyDisplay(minor int64, currency string) string {
	major := float64(minor) / 100
	cur := strings.ToUpper(strings.TrimSpace(currency))
	switch cur {
	case "NGN":
		return fmt.Sprintf("₦%s", formatMoneyMajor(major))
	case "USD":
		return fmt.Sprintf("$%s", formatMoneyMajor(major))
	case "GBP":
		return fmt.Sprintf("£%s", formatMoneyMajor(major))
	case "EUR":
		return fmt.Sprintf("€%s", formatMoneyMajor(major))
	case "":
		return formatMoneyMajor(major)
	default:
		return fmt.Sprintf("%s %s", cur, formatMoneyMajor(major))
	}
}

func formatMoneyMajor(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}
