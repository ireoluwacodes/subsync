package utils

import "github.com/ireoluwacodes/subsync/internal/domain"

// PlanMonthlyAmount normalizes a plan price to a monthly amount in minor units (kobo).
func PlanMonthlyAmount(plan *domain.Plan) int64 {
	if plan == nil || plan.Amount <= 0 {
		return 0
	}
	switch plan.Interval {
	case domain.PlanIntervalAnnual:
		return plan.Amount / 12
	case domain.PlanIntervalCustom:
		days := 30
		if plan.IntervalDays != nil && *plan.IntervalDays > 0 {
			days = *plan.IntervalDays
		}
		return (plan.Amount * 30) / int64(days)
	default:
		return plan.Amount
	}
}
