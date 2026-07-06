package service

import (
	"testing"
	"time"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestValidateDirectDebitSetupInput_RequiresAddress(t *testing.T) {
	err := validateDirectDebitSetupInput(DirectDebitSetupInput{
		CustomerAccountNumber: "0123456789",
		BankCode:              "058",
		CustomerName:          "Jane Doe",
		CustomerAccountName:   "Jane Doe",
		CustomerPhone:         "08012345678",
	})
	require.ErrorIs(t, err, domain.ErrValidation)
	require.Contains(t, err.Error(), "address")

	err = validateDirectDebitSetupInput(DirectDebitSetupInput{
		CustomerAccountNumber: "0123456789",
		BankCode:              "110005",
		CustomerName:          "Jane Doe",
		CustomerAccountName:   "Jane Doe",
		CustomerPhone:         "08012345678",
		CustomerAddress:       "12 Allen Avenue, Ikeja",
	})
	require.ErrorIs(t, err, domain.ErrValidation)
	require.Contains(t, err.Error(), "NIBSS")

	err = validateDirectDebitSetupInput(DirectDebitSetupInput{
		CustomerAccountNumber: "0123456789",
		BankCode:              "058",
		CustomerName:          "Jane Doe",
		CustomerAccountName:   "Jane Doe",
		CustomerPhone:         "08012345678",
		CustomerAddress:       "12 Allen Avenue, Ikeja",
	})
	require.NoError(t, err)
}

func TestMandateScheduleTimes_UsesNigeriaLocalTime(t *testing.T) {
	// 2026-07-06 23:45 UTC is 2026-07-07 00:45 WAT — UTC formatting would send a past date.
	utc := time.Date(2026, 7, 6, 23, 45, 0, 0, time.UTC)
	start, end := mandateScheduleTimes(utc)

	require.Equal(t, "2026-07-07T00:47", start)
	require.Equal(t, "2031-07-07T00:47", end)
}

func TestMandateScheduleTimes_StartIsInFutureInWAT(t *testing.T) {
	utc := time.Date(2026, 7, 6, 20, 0, 0, 0, time.UTC)
	start, _ := mandateScheduleTimes(utc)

	parsed, err := time.ParseInLocation("2006-01-02T15:04", start, nigeriaLoc)
	require.NoError(t, err)
	require.True(t, parsed.After(utc.In(nigeriaLoc)))
}
