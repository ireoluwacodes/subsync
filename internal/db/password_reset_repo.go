package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type PasswordResetRepo struct {
	db *DB
}

func NewPasswordResetRepo(db *DB) *PasswordResetRepo {
	return &PasswordResetRepo{db: db}
}

func (r *PasswordResetRepo) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	m := &models.PasswordResetToken{
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	}
	if token.ID != uuid.Nil {
		m.ID = token.ID
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	token.ID = m.ID
	token.CreatedAt = m.CreatedAt
	return nil
}

func (r *PasswordResetRepo) GetLatestValidByUserID(ctx context.Context, userID uuid.UUID) (*domain.PasswordResetToken, error) {
	var m models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return mapPasswordResetToken(m), nil
}

func (r *PasswordResetRepo) InvalidateUnusedForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now)
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	return nil
}

func (r *PasswordResetRepo) UpdateTokenHash(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Updates(map[string]any{
			"token_hash": tokenHash,
			"expires_at": expiresAt,
		})
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PasswordResetRepo) GetValidByTokenHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	var m models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&m).Error
	if err != nil {
		return nil, MapGORMError(err)
	}
	return mapPasswordResetToken(m), nil
}

func mapPasswordResetToken(m models.PasswordResetToken) *domain.PasswordResetToken {
	return &domain.PasswordResetToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		CreatedAt: m.CreatedAt,
	}
}

func (r *PasswordResetRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&models.PasswordResetToken{}).Where("id = ?", id).Update("used_at", now)
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var _ domain.PasswordResetRepository = (*PasswordResetRepo)(nil)
