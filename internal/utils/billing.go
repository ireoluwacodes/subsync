package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

type DunningStep struct {
	DelayDays int    `json:"delay_days"`
	Action    string `json:"action"`
}

func (s DunningStep) String() string {
	return fmt.Sprintf("%s@%dd", s.Action, s.DelayDays)
}

func ParseDunningSteps(config map[string]any) ([]DunningStep, error) {
	if config == nil {
		return defaultDunningSteps(), nil
	}
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Steps []DunningStep `json:"steps"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}
	if len(wrapper.Steps) == 0 {
		return defaultDunningSteps(), nil
	}
	return wrapper.Steps, nil
}

func DefaultDunningConfig() map[string]any {
	return map[string]any{
		"steps": []map[string]any{
			{"delay_days": 1, "action": "retry"},
			{"delay_days": 3, "action": "retry_and_notify"},
			{"delay_days": 7, "action": "mandate_fallback"},
			{"delay_days": 14, "action": "cancel"},
		},
	}
}

func PlanPeriodEnd(start time.Time, plan *domain.Plan) time.Time {
	switch plan.Interval {
	case domain.PlanIntervalAnnual:
		return start.AddDate(1, 0, 0)
	case domain.PlanIntervalCustom:
		days := 30
		if plan.IntervalDays != nil {
			days = *plan.IntervalDays
		}
		return start.AddDate(0, 0, days)
	default:
		return start.AddDate(0, 1, 0)
	}
}

func defaultDunningSteps() []DunningStep {
	return []DunningStep{
		{DelayDays: 1, Action: "retry"},
		{DelayDays: 3, Action: "retry_and_notify"},
		{DelayDays: 7, Action: "mandate_fallback"},
		{DelayDays: 14, Action: "cancel"},
	}
}
