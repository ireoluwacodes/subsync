package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/pressly/goose/v3"

	appmigrations "github.com/ireoluwacodes/subsync/migrations"
)

var setMigrationsFSOnce sync.Once

func Migrate(ctx context.Context, database *DB) error {
	setMigrationsFSOnce.Do(func() {
		goose.SetBaseFS(appmigrations.FS)
	})

	sqlDB, err := database.Gorm.DB()
	if err != nil {
		return fmt.Errorf("migrate: sql db: %w", err)
	}
	if err := goose.UpContext(ctx, sqlDB, "."); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
