package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type PortalTokenRepo struct {
	db *DB
}

func NewPortalTokenRepo(db *DB) *PortalTokenRepo {
	return &PortalTokenRepo{db: db}
}

func (r *PortalTokenRepo) Create(ctx context.Context, token *domain.PortalToken) error {
	m := models.PortalTokenFromDomain(token)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	*token = *models.PortalTokenToDomain(m)
	return nil
}

func (r *PortalTokenRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PortalToken, error) {
	var m models.PortalToken
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.PortalTokenToDomain(&m), nil
}

func (r *PortalTokenRepo) GetValidByTokenHash(ctx context.Context, tokenHash string) (*domain.PortalToken, error) {
	var m models.PortalToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now().UTC()).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return models.PortalTokenToDomain(&m), nil
}

func (r *PortalTokenRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&models.PortalToken{}).
		Where("id = ?", id).
		Update("used_at", now)
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var _ domain.PortalTokenRepository = (*PortalTokenRepo)(nil)
