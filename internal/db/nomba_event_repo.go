package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type NombaEventRepo struct {
	db *DB
}

func NewNombaEventRepo(db *DB) *NombaEventRepo {
	return &NombaEventRepo{db: db}
}

func (r *NombaEventRepo) Create(ctx context.Context, event *domain.NombaEvent) error {
	m, err := models.NombaEventFromDomain(event)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	*event = *models.NombaEventToDomain(m)
	return nil
}

func (r *NombaEventRepo) GetByEventID(ctx context.Context, tenantID uuid.UUID, eventID string) (*domain.NombaEvent, error) {
	var m models.NombaEvent
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND event_id = ?", tenantID, eventID).
		First(&m).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.NombaEventToDomain(&m), nil
}

func (r *NombaEventRepo) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&models.NombaEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"processed":    true,
			"processed_at": now,
			"error":        nil,
		})
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NombaEventRepo) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	res := r.db.WithContext(ctx).Model(&models.NombaEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{"error": errMsg})
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var _ domain.NombaEventRepository = (*NombaEventRepo)(nil)
