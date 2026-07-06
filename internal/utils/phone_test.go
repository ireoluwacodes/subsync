package utils

import (
	"testing"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNormalizeNigerianPhone(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"08012345678", "08012345678"},
		{"8012345678", "08012345678"},
		{"+2348012345678", "08012345678"},
		{"2348012345678", "08012345678"},
		{"+234 913 732 9756", "09137329756"},
		{"+2349137329756", "09137329756"},
	}
	for _, tc := range tests {
		got, err := NormalizeNigerianPhone(tc.in)
		require.NoError(t, err, tc.in)
		require.Equal(t, tc.want, got, tc.in)
	}
}

func TestNormalizeNigerianPhone_Invalid(t *testing.T) {
	_, err := NormalizeNigerianPhone("12345")
	require.ErrorIs(t, err, domain.ErrValidation)
}
