package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/config"
)

// wrapper para el servicio de subida a S3
type S3config struct {
	config.UploadService
}

type S3Svc struct {
	config S3config
}

type S3Service interface {
	HandleUpload(w http.ResponseWriter, r *http.Request)
}

func NewS3Svs(config S3config) S3Service {
	return &S3Svc{config: config}
}

func (s *S3Svc) HandleUpload(w http.ResponseWriter, r *http.Request) {

	s3Var := s.config

	r.Body = http.MaxBytesReader(w, r.Body, s3Var.MaxUploadMB*1024*1024)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	if err := r.ParseMultipartForm(s3Var.MaxUploadMB * 1024 * 1024); err != nil {
		httpError(w, http.StatusRequestEntityTooLarge, fmt.Errorf("archivo demasiado grande o inválido: %w", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httpError(w, http.StatusBadRequest, fmt.Errorf("campo 'file' requerido: %w", err))
		return
	}
	defer file.Close()

	filename := sanitizeFilename(header.Filename)
	contentType := detectContentType(file, header)
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	key := buildObjectKey(filename, r.FormValue("prefix"))

	_, err = s3Var.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3Var.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		httpError(w, http.StatusInternalServerError, fmt.Errorf("error subiendo a S3: %w", err))
		return
	}

	publicURL := fmt.Sprintf("%s/%s", s3Var.PublicBase, key)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"key":%q,"public_url":%q}`, key, publicURL)))
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

func detectContentType(file multipart.File, header *multipart.FileHeader) string {
	if ct := header.Header.Get("Content-Type"); ct != "" {
		return ct
	}
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	return http.DetectContentType(buf[:n])
}

func httpError(w http.ResponseWriter, code int, err error) {
	msg := err.Error()
	if errors.Is(err, http.ErrBodyNotAllowed) {
		msg = "método no permitido"
	}
	http.Error(w, msg, code)
}
