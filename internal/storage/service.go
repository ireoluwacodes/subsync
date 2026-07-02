package storage

import (
	"context"

	"github.com/ireoluwacodes/subsync/internal/config"
)

type StorageService struct {
	strategy StorageStrategy
}

func NewStorageService(cfg *config.Config) *StorageService {
	var strategy StorageStrategy = NoopStrategy{}
	if cfg != nil && cfg.CloudinaryCloudName != "" && cfg.CloudinaryAPIKey != "" && cfg.CloudinaryAPISecret != "" {
		if cld, err := NewCloudinaryStrategy(
			cfg.CloudinaryCloudName,
			cfg.CloudinaryAPIKey,
			cfg.CloudinaryAPISecret,
			cfg.CloudinaryFolder,
		); err == nil {
			strategy = cld
		}
	}
	return &StorageService{strategy: strategy}
}

func (s *StorageService) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	return s.strategy.Upload(ctx, key, data, contentType)
}
