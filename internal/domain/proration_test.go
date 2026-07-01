package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateProration_HalfPeriod(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	mid := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	result := CalculateProration(30000, 60000, start, end, mid)

	assert.Positive(t, result.CreditAmount)
	assert.Positive(t, result.DebitAmount)
	assert.Equal(t, result.NetAmount, result.DebitAmount-result.CreditAmount)
}
