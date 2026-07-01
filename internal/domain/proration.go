package domain

import "time"

type ProrationResult struct {
	CreditAmount int64
	DebitAmount  int64
	NetAmount    int64
	DaysRemaining int
	TotalDays     int
}

// CalculateProration computes mid-cycle upgrade/downgrade amounts in kobo.
func CalculateProration(oldAmount, newAmount int64, periodStart, periodEnd, changeAt time.Time) ProrationResult {
	total := periodEnd.Sub(periodStart)
	if total <= 0 {
		return ProrationResult{}
	}

	remaining := periodEnd.Sub(changeAt)
	if remaining < 0 {
		remaining = 0
	}

	totalDays := int(total.Hours() / 24)
	if totalDays == 0 {
		totalDays = 1
	}
	daysRemaining := int(remaining.Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	credit := oldAmount * int64(daysRemaining) / int64(totalDays)
	debit := newAmount * int64(daysRemaining) / int64(totalDays)

	return ProrationResult{
		CreditAmount:  credit,
		DebitAmount:   debit,
		NetAmount:     debit - credit,
		DaysRemaining: daysRemaining,
		TotalDays:     totalDays,
	}
}
