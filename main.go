package main

import (
	"context"
	"log"
	"net/http"

	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/http"
	"gorm.io/gorm"
)

func main() {

	//obtenemos Secrets Manager
	ctx := context.Background()

	cfg := getSecrets(ctx)

	//conectamos a la base de datos
	db := getDbConfig(cfg)

	//ejecutamos migraciones
	config.RunMigrations(cfg.ToDBConfig())

	//configuramos S3
	s3 := config.S3ConfigService(cfg.ToS3Config())

	log.Printf("S3 Configured: Bucket %s ", s3.Bucket)

	r := httpserver.NewRouter(*s3, db, nil, nil, cfg.S3Region, "", cfg.ToMQConfig())

	addr := ":8082"
	log.Printf("API listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}

}

func getDbConfig(cfg *config.SecretApp) *gorm.DB {
	db, err := config.NewPostgresDB(cfg.ToDBConfig())
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	log.Println("DB Connection Established")
	return db
}

func getSecrets(ctx context.Context) *config.SecretApp {
	cfg, err := config.LoadSecretManager(ctx)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	return cfg
}
