package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"gorm.io/gorm"
)

type PlanRepo struct {
	db *DB
}

func NewPlanRepo(db *DB) *PlanRepo {
	return &PlanRepo{db: db}
}

func (r *PlanRepo) Create(ctx context.Context, plan *domain.Plan) error {
	m, err := models.PlanFromDomain(plan)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}

	*plan = *models.PlanToDomain(m)
	return nil
}

func (r *PlanRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Plan, error) {
	var m models.Plan
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.PlanToDomain(&m), nil
}

func (r *PlanRepo) List(ctx context.Context, tenantID uuid.UUID, activeOnly bool, limit, offset int) ([]*domain.Plan, error) {
	q := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_archived = ?", tenantID, false)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}

	var rows []models.Plan
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}

	plans := make([]*domain.Plan, len(rows))
	for i := range rows {
		plans[i] = models.PlanToDomain(&rows[i])
	}
	return plans, nil
}

func (r *PlanRepo) Count(ctx context.Context, tenantID uuid.UUID, activeOnly bool) (int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Plan{}).
		Where("tenant_id = ? AND is_archived = ?", tenantID, false)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PlanRepo) Update(ctx context.Context, plan *domain.Plan) error {
	m, err := models.PlanFromDomain(plan)
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).
		Model(&models.Plan{}).
		Where("tenant_id = ? AND id = ? AND is_archived = ?", plan.TenantID, plan.ID, false).
		Updates(map[string]any{
			"name":          m.Name,
			"description":   m.Description,
			"amount":        m.Amount,
			"currency":      m.Currency,
			"interval":      m.Interval,
			"interval_days": m.IntervalDays,
			"trial_days":    m.TrialDays,
			"features":      m.Features,
			"is_active":     m.IsActive,
			"updated_at":    gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return MapGORMError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PlanRepo) Archive(ctx context.Context, tenantID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Plan{}).
		Where("tenant_id = ? AND id = ? AND is_archived = ?", tenantID, id, false).
		Updates(map[string]any{
			"is_archived": true,
			"is_active":   false,
			"updated_at":  gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return MapGORMError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PlanRepo) IsReferenced(ctx context.Context, planID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM subscriptions WHERE plan_id = ?)", planID).
		Scan(&exists).Error
	return exists, err
}

var _ domain.PlanRepository = (*PlanRepo)(nil)
