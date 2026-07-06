package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"gorm.io/gorm"
)

type PaymentMethodRepo struct {
	db *DB
}

func NewPaymentMethodRepo(db *DB) *PaymentMethodRepo {
	return &PaymentMethodRepo{db: db}
}

func (r *PaymentMethodRepo) Create(ctx context.Context, pm *domain.PaymentMethod) error {
	m := models.PaymentMethodFromDomain(pm)

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}

	*pm = *models.PaymentMethodToDomain(m)
	return nil
}

func (r *PaymentMethodRepo) Update(ctx context.Context, pm *domain.PaymentMethod) error {
	m := models.PaymentMethodFromDomain(pm)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return MapGORMError(err)
	}
	*pm = *models.PaymentMethodToDomain(m)
	return nil
}

func (r *PaymentMethodRepo) GetDefaultForCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*domain.PaymentMethod, error) {
	var m models.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND customer_id = ? AND is_default = ?", tenantID, customerID, true).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.PaymentMethodToDomain(&m), nil
}

func (r *PaymentMethodRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.PaymentMethod, error) {
	var m models.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.PaymentMethodToDomain(&m), nil
}

func (r *PaymentMethodRepo) ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) ([]*domain.PaymentMethod, error) {
	var rows []models.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	out := make([]*domain.PaymentMethod, len(rows))
	for i := range rows {
		out[i] = models.PaymentMethodToDomain(&rows[i])
	}
	return out, nil
}

func (r *PaymentMethodRepo) GetDirectDebitForCustomer(ctx context.Context, tenantID, customerID uuid.UUID, preferID *uuid.UUID) (*domain.PaymentMethod, error) {
	if preferID != nil && *preferID != uuid.Nil {
		pm, err := r.GetByID(ctx, tenantID, *preferID)
		if err == nil && pm.Type == domain.PaymentMethodDirectDebit && pm.MandateReady() {
			return pm, nil
		}
	}
	var m models.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND customer_id = ? AND type = ? AND mandate_status = ?",
			tenantID, customerID, string(domain.PaymentMethodDirectDebit), string(domain.MandateStatusReady)).
		Order("created_at DESC").
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.PaymentMethodToDomain(&m), nil
}

func (r *PaymentMethodRepo) ListPendingMandates(ctx context.Context, limit int) ([]*domain.PaymentMethod, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("type = ? AND mandate_status = ?", string(domain.PaymentMethodDirectDebit), string(domain.MandateStatusPending)).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	out := make([]*domain.PaymentMethod, len(rows))
	for i := range rows {
		out[i] = models.PaymentMethodToDomain(&rows[i])
	}
	return out, nil
}

func (r *PaymentMethodRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&models.PaymentMethod{})
	if result.Error != nil {
		return MapGORMError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PaymentMethodRepo) SetDefault(ctx context.Context, tenantID, customerID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.PaymentMethod{}).
			Where("tenant_id = ? AND customer_id = ? AND id = ?", tenantID, customerID, id).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return domain.ErrNotFound
		}

		if err := tx.Model(&models.PaymentMethod{}).
			Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).
			Updates(map[string]any{
				"is_default": false,
				"updated_at": gorm.Expr("NOW()"),
			}).Error; err != nil {
			return err
		}

		result := tx.Model(&models.PaymentMethod{}).
			Where("tenant_id = ? AND id = ?", tenantID, id).
			Updates(map[string]any{
				"is_default": true,
				"updated_at": gorm.Expr("NOW()"),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
}

var _ domain.PaymentMethodRepository = (*PaymentMethodRepo)(nil)
