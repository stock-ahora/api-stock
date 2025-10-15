package config_lib

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client(region, bucket string) (*s3.Client, *manager.Uploader, string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, nil, "", fmt.Errorf("no see pudo cargar config AWS: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3Client)

	publicBase := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)

	return s3Client, uploader, publicBase, nil
}
