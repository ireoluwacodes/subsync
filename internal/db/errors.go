package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"gorm.io/gorm"
)

const pgUniqueViolation = "23505"

func MapGORMError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrConflict
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return domain.ErrConflict
	}
	return err
}

// MapPGError is an alias kept for compatibility during migration to GORM.
func MapPGError(err error) error {
	return MapGORMError(err)
}
