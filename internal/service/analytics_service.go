package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type AnalyticsService struct {
	repo  *db.AnalyticsRepo
	clock clock.Clock
}

func NewAnalyticsService(repo *db.AnalyticsRepo, clk clock.Clock) *AnalyticsService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &AnalyticsService{repo: repo, clock: clk}
}

func (s *AnalyticsService) MRR(ctx context.Context, tenantID uuid.UUID, currency string) (*domain.AnalyticsMRRResult, error) {
	return s.repo.MRR(ctx, tenantID, currency)
}

func (s *AnalyticsService) Churn(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (*domain.AnalyticsChurnResult, error) {
	return s.repo.Churn(ctx, tenantID, from, to)
}

func (s *AnalyticsService) Dunning(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (*domain.AnalyticsDunningResult, error) {
	return s.repo.Dunning(ctx, tenantID, from, to)
}

func (s *AnalyticsService) Revenue(ctx context.Context, tenantID uuid.UUID, from, to time.Time, currency string) (*domain.AnalyticsRevenueResult, error) {
	return s.repo.Revenue(ctx, tenantID, from, to, currency)
}

func (s *AnalyticsService) DefaultRange(from, to *time.Time) (time.Time, time.Time) {
	now := s.clock.Now().UTC()
	end := now
	if to != nil {
		end = to.UTC()
	}
	start := end.AddDate(0, 0, -30)
	if from != nil {
		start = from.UTC()
	}
	return start, end
}
