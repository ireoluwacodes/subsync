package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"gorm.io/gorm"
)

type SubscriptionRepo struct {
	db *DB
}

func NewSubscriptionRepo(db *DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	m, err := models.SubscriptionFromDomain(sub)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	*sub = *models.SubscriptionToDomain(m)
	return nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Subscription, error) {
	var m models.Subscription
	if err := r.db.WithContext(ctx).First(&m, "id = ? AND tenant_id = ?", id, tenantID).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.SubscriptionToDomain(&m), nil
}

func (r *SubscriptionRepo) GetByIDGlobal(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	var m models.Subscription
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.SubscriptionToDomain(&m), nil
}

func (r *SubscriptionRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.SubscriptionListFilter) ([]*domain.Subscription, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("tenant_id = ?", tenantID)
	if filter.CustomerID != nil {
		q = q.Where("customer_id = ?", *filter.CustomerID)
	}
	if filter.PlanID != nil {
		q = q.Where("plan_id = ?", *filter.PlanID)
	}
	if filter.State != "" {
		q = q.Where("state = ?", filter.State)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	sort := "created_at DESC"
	if filter.Sort == "updated_at" {
		sort = "updated_at DESC"
	}

	var rows []models.Subscription
	if err := q.Order(sort).Limit(limit).Offset(filter.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*domain.Subscription, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionToDomain(&rows[i])
	}
	return out, total, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	m, err := models.SubscriptionFromDomain(sub)
	if err != nil {
		return err
	}
	res := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ? AND tenant_id = ?", sub.ID, sub.TenantID).Omit("CreatedAt").Save(m)
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepo) RecordTransition(ctx context.Context, t *domain.SubscriptionTransition) error {
	m, err := models.SubscriptionTransitionFromDomain(t)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	t.ID = m.ID
	t.CreatedAt = m.CreatedAt
	return nil
}

func (r *SubscriptionRepo) ListTransitions(ctx context.Context, tenantID, subscriptionID uuid.UUID) ([]*domain.SubscriptionTransition, error) {
	var rows []models.SubscriptionTransition
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND subscription_id = ?", tenantID, subscriptionID).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.SubscriptionTransition, len(rows))
	for i := range rows {
		out[i] = models.SubscriptionTransitionToDomain(&rows[i])
	}
	return out, nil
}

func (r *SubscriptionRepo) CountActiveByPlan(ctx context.Context, tenantID, planID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("tenant_id = ? AND plan_id = ? AND state IN ?", tenantID, planID, []string{
			string(domain.SubscriptionStateActive),
			string(domain.SubscriptionStateTrialing),
			string(domain.SubscriptionStatePastDue),
		}).Count(&count).Error
	return count, err
}

func (r *SubscriptionRepo) Transition(ctx context.Context, sub *domain.Subscription, t *domain.SubscriptionTransition) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := models.SubscriptionFromDomain(sub)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.Subscription{}).Where("id = ? AND tenant_id = ?", sub.ID, sub.TenantID).Updates(m).Error; err != nil {
			return MapGORMError(err)
		}
		tr, err := models.SubscriptionTransitionFromDomain(t)
		if err != nil {
			return err
		}
		if err := tx.Create(tr).Error; err != nil {
			return MapGORMError(err)
		}
		t.ID = tr.ID
		t.CreatedAt = tr.CreatedAt
		return nil
	})
}

var _ domain.SubscriptionRepository = (*SubscriptionRepo)(nil)
