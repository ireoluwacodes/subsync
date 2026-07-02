package storage

import "context"

type StorageStrategy interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
}
