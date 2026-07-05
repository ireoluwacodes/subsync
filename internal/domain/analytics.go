package domain

import "time"

type AnalyticsMRRResult struct {
	MRR      int64
	Currency string
	Active   int64
}

type AnalyticsChurnResult struct {
	CanceledInPeriod int64
	ActiveCount      int64
	ChurnRate        float64
	From             time.Time
	To               time.Time
}

type AnalyticsDunningResult struct {
	EnteredPastDue int64
	Recovered      int64
	RecoveryRate   float64
	CurrentlyPastDue int64
	From           time.Time
	To             time.Time
}

type RevenueDailyPoint struct {
	Date   string
	Amount int64
}

type AnalyticsRevenueResult struct {
	Total    int64
	Currency string
	From     time.Time
	To       time.Time
	Daily    []RevenueDailyPoint
}
