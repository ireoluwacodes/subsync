package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	m := models.UserFromDomain(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}
	*user = *models.UserToDomain(m)
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m models.User
	if err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.UserToDomain(&m), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var m models.User
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.UserToDomain(&m), nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	res := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password_hash", passwordHash)
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepo) BumpTokenVersion(ctx context.Context, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).Exec(
		"UPDATE users SET token_version = token_version + 1, updated_at = NOW() WHERE id = ?",
		userID,
	)
	if res.Error != nil {
		return MapGORMError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var _ domain.UserRepository = (*UserRepo)(nil)
