package utils

import (
	"fmt"
	"strings"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// NormalizeNigerianPhone formats a phone number for Nomba/NIBSS (11 digits, leading 0).
// Accepts 08012345678, 8012345678, +2348012345678, and 2348012345678.
func NormalizeNigerianPhone(raw string) (string, error) {
	digits := onlyDigits(strings.TrimSpace(raw))
	if digits == "" {
		return "", fmt.Errorf("%w: phone number is required", domain.ErrValidation)
	}

	switch {
	case len(digits) == 11 && digits[0] == '0':
		return digits, nil
	case len(digits) == 10:
		return "0" + digits, nil
	case len(digits) == 13 && strings.HasPrefix(digits, "234"):
		local := digits[3:]
		if len(local) != 10 {
			break
		}
		return "0" + local, nil
	}

	return "", fmt.Errorf("%w: enter a valid Nigerian phone number (e.g. 08012345678 or +2348012345678)", domain.ErrValidation)
}

func onlyDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
