package utils_test

import (
	"testing"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestPlanMonthlyAmount(t *testing.T) {
	t.Run("monthly", func(t *testing.T) {
		p := &domain.Plan{Amount: 500000, Interval: domain.PlanIntervalMonthly}
		require.Equal(t, int64(500000), utils.PlanMonthlyAmount(p))
	})

	t.Run("annual", func(t *testing.T) {
		p := &domain.Plan{Amount: 1200000, Interval: domain.PlanIntervalAnnual}
		require.Equal(t, int64(100000), utils.PlanMonthlyAmount(p))
	})

	t.Run("custom 15 days", func(t *testing.T) {
		days := 15
		p := &domain.Plan{Amount: 150000, Interval: domain.PlanIntervalCustom, IntervalDays: &days}
		require.Equal(t, int64(300000), utils.PlanMonthlyAmount(p))
	})

	t.Run("nil plan", func(t *testing.T) {
		require.Equal(t, int64(0), utils.PlanMonthlyAmount(nil))
	})
}
