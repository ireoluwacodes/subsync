package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryStrategy struct {
	cld    *cloudinary.Cloudinary
	folder string
}

func NewCloudinaryStrategy(cloudName, apiKey, apiSecret, folder string) (*CloudinaryStrategy, error) {
	if folder == "" {
		folder = "subsync/invoices"
	}
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &CloudinaryStrategy{cld: cld, folder: folder}, nil
}

func (s *CloudinaryStrategy) Upload(ctx context.Context, key string, data []byte, _ string) (string, error) {
	publicID := strings.TrimPrefix(key, "/")
	result, err := s.cld.Upload.Upload(ctx, bytes.NewReader(data), uploader.UploadParams{
		Folder:       s.folder,
		PublicID:     publicID,
		ResourceType: string(api.File),
	})
	if err != nil {
		return "", err
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("cloudinary: %s", result.Error.Message)
	}
	return result.SecureURL, nil
}
