package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"gorm.io/gorm"
)

type CustomerRepo struct {
	db *DB
}

func NewCustomerRepo(db *DB) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) Create(ctx context.Context, customer *domain.Customer) error {
	m, err := models.CustomerFromDomain(customer)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}

	*customer = *models.CustomerToDomain(m)
	return nil
}

func (r *CustomerRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Customer, error) {
	var m models.Customer
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.CustomerToDomain(&m), nil
}

func (r *CustomerRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Customer, error) {
	var rows []models.Customer
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	customers := make([]*domain.Customer, len(rows))
	for i := range rows {
		customers[i] = models.CustomerToDomain(&rows[i])
	}
	return customers, nil
}

func (r *CustomerRepo) Count(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

func (r *CustomerRepo) Update(ctx context.Context, customer *domain.Customer) error {
	m, err := models.CustomerFromDomain(customer)
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("tenant_id = ? AND id = ?", customer.TenantID, customer.ID).
		Updates(map[string]any{
			"external_id": m.ExternalID,
			"email":       m.Email,
			"name":        m.Name,
			"phone":       m.Phone,
			"metadata":    m.Metadata,
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

var _ domain.CustomerRepository = (*CustomerRepo)(nil)
