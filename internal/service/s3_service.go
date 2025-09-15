package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/dto"
)

// wrapper para el servicio de subida a S3
type S3config struct {
	config.UploadService
}

// implementación concreta
type S3Svc struct {
	config S3config
}

// constructor del servicio
func NewS3Svs(config S3config) *S3Svc {
	return &S3Svc{config: config}
}

// todo modificar servicio de subida de archivos a s3

func (s *S3Svc) doHandleUpload(createRequest *dto.CreateRequestDto, path string) (string, error) {

	s3Var := s.config

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filename := sanitizeFilename(createRequest.FileName)
	key := buildObjectKey(filename, path)

	_, err := s3Var.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3Var.Bucket),
		Key:         aws.String(key),
		Body:        createRequest.File,
		ContentType: aws.String(createRequest.FileType),
	})
	if err != nil {
		return "", fmt.Errorf("error al cargar archivo en S3: %w", err)
	}

	return key, err
}

func buildObjectKey(filename, prefix string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	base := strings.TrimSuffix(filename, ext)
	if base == "" {
		base = "file"
	}
	id := uuid.New().String()
	key := fmt.Sprintf("%s-%s%s", base, id, ext)
	if prefix = strings.Trim(prefix, "/"); prefix != "" {
		key = prefix + "/" + key
	}
	return key
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, string(filepath.Separator), "-")
	if name == "" {
		return "file"
	}
	return name
}
