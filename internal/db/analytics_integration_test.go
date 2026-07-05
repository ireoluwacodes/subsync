//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

func TestAnalyticsRepo_Integration(t *testing.T) {
	database := testDB(t)
	repos := db.NewRepos(database, testEncryptor(t))
	ctx := context.Background()

	tenant := createTestTenant(t, repos.Tenants, ctx)
	customer := createTestCustomer(t, repos.Customers, ctx, tenant.ID)

	monthlyPlan := &domain.Plan{
		TenantID: tenant.ID,
		Name:     "Monthly",
		Amount:   100000,
		Currency: "NGN",
		Interval: domain.PlanIntervalMonthly,
		IsActive: true,
	}
	require.NoError(t, repos.Plans.Create(ctx, monthlyPlan))

	now := time.Now().UTC()
	sub1 := &domain.Subscription{
		TenantID:           tenant.ID,
		CustomerID:         customer.ID,
		PlanID:             monthlyPlan.ID,
		State:              domain.SubscriptionStateActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		Metadata:           map[string]any{},
	}
	require.NoError(t, repos.Subscriptions.Create(ctx, sub1))

	sub2 := &domain.Subscription{
		TenantID:           tenant.ID,
		CustomerID:         customer.ID,
		PlanID:             monthlyPlan.ID,
		State:              domain.SubscriptionStatePastDue,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		Metadata:           map[string]any{},
	}
	require.NoError(t, repos.Subscriptions.Create(ctx, sub2))

	mrr, err := repos.Analytics.MRR(ctx, tenant.ID, "")
	require.NoError(t, err)
	require.Equal(t, int64(200000), mrr.MRR)
	require.Equal(t, int64(2), mrr.Active)

	from := now.AddDate(0, 0, -7)
	to := now.Add(24 * time.Hour)
	require.NoError(t, repos.Subscriptions.RecordTransition(ctx, &domain.SubscriptionTransition{
		SubscriptionID: sub1.ID,
		TenantID:       tenant.ID,
		FromState:      domain.SubscriptionStateActive,
		ToState:        domain.SubscriptionStateCanceled,
		Reason:         "customer_request",
		Actor:          "system",
	}))

	churn, err := repos.Analytics.Churn(ctx, tenant.ID, from, to)
	require.NoError(t, err)
	require.Equal(t, int64(1), churn.CanceledInPeriod)

	require.NoError(t, repos.Subscriptions.RecordTransition(ctx, &domain.SubscriptionTransition{
		SubscriptionID: sub2.ID,
		TenantID:       tenant.ID,
		FromState:      domain.SubscriptionStateActive,
		ToState:        domain.SubscriptionStatePastDue,
		Reason:         "payment_failed",
		Actor:          "system",
	}))
	require.NoError(t, repos.Subscriptions.RecordTransition(ctx, &domain.SubscriptionTransition{
		SubscriptionID: sub2.ID,
		TenantID:       tenant.ID,
		FromState:      domain.SubscriptionStatePastDue,
		ToState:        domain.SubscriptionStateActive,
		Reason:         "payment_succeeded",
		Actor:          "system",
	}))

	dunning, err := repos.Analytics.Dunning(ctx, tenant.ID, from, to)
	require.NoError(t, err)
	require.Equal(t, int64(1), dunning.EnteredPastDue)
	require.Equal(t, int64(1), dunning.Recovered)
	require.Equal(t, float64(1), dunning.RecoveryRate)

	paidAt := now.Add(-2 * 24 * time.Hour)
	inv := &domain.Invoice{
		TenantID:       tenant.ID,
		SubscriptionID: sub1.ID,
		CustomerID:     customer.ID,
		Status:         domain.InvoiceStatusPaid,
		AmountDue:      100000,
		AmountPaid:     100000,
		Currency:       "NGN",
		PeriodStart:    now.AddDate(0, -1, 0),
		PeriodEnd:      now,
		PaidAt:         &paidAt,
		NombaOrderRef:  uuid.New().String(),
		Metadata:       map[string]any{},
	}
	require.NoError(t, repos.Invoices.Create(ctx, inv))

	revenue, err := repos.Analytics.Revenue(ctx, tenant.ID, from, to, "")
	require.NoError(t, err)
	require.Equal(t, int64(100000), revenue.Total)
	require.NotEmpty(t, revenue.Daily)
}
