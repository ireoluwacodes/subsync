package storage

import (
	"context"

	"go.uber.org/zap"
)

type NoopStrategy struct{}

func (NoopStrategy) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	zap.L().Debug("storage noop upload", zap.String("key", key), zap.Int("bytes", len(data)))
	return "", nil
}
