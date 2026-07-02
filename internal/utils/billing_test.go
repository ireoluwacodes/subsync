package utils

import (
	"testing"
	"time"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDunningSteps_Default(t *testing.T) {
	steps, err := ParseDunningSteps(nil)
	require.NoError(t, err)
	require.Len(t, steps, 4)
	assert.Equal(t, "retry", steps[0].Action)
	assert.Equal(t, "cancel", steps[3].Action)
}

func TestParseDunningSteps_Custom(t *testing.T) {
	cfg := map[string]any{
		"steps": []map[string]any{
			{"delay_days": 2, "action": "retry"},
			{"delay_days": 5, "action": "cancel"},
		},
	}
	steps, err := ParseDunningSteps(cfg)
	require.NoError(t, err)
	require.Len(t, steps, 2)
	assert.Equal(t, 2, steps[0].DelayDays)
}

func TestPlanPeriodEnd_Monthly(t *testing.T) {
	start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	end := PlanPeriodEnd(start, &domain.Plan{Interval: domain.PlanIntervalMonthly})
	assert.Equal(t, time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC), end)
}
