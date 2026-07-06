package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatMoneyDisplay_NGN(t *testing.T) {
	require.Equal(t, "₦1000", FormatMoneyDisplay(100000, "NGN"))
}

func TestFormatMoneyDisplay_NGNWithCents(t *testing.T) {
	require.Equal(t, "₦10.50", FormatMoneyDisplay(1050, "NGN"))
}

func TestFormatMoneyDisplay_USD(t *testing.T) {
	require.Equal(t, "$25", FormatMoneyDisplay(2500, "USD"))
}
