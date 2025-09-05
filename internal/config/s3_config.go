package config

import (
	"log"

	"github.com/stock-ahora/api-stock/internal/config_lib"
)

func S3ConfigService(s3 S3Config) *UploadService {

	if s3.Region == "" || s3.Bucket == "" {
		log.Fatal("Faltan variables de entorno: AWS_REGION y/o S3_BUCKET")
	}

	// Configuración de cliente S3
	s3Client, uploader, publicBase, err := config_lib.NewS3Client(s3.Region, s3.Bucket)
	if err != nil {
		log.Fatalf("error creando cliente S3: %v", err)
	}

	// Inicializa el servicio
	srv := &UploadService{
		S3Client:    s3Client,
		Uploader:    uploader,
		Bucket:      s3.Bucket,
		PublicBase:  publicBase,
		MaxUploadMB: 25,
	}

	return srv

}
