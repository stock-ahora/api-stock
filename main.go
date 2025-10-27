package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/http"
	"gorm.io/gorm"
)

func main() {

	//obtenemos Secrets Manager
	ctx := context.Background()

	cfg := getSecrets(ctx)

	//conectamos a la base de datos
	db, DBStarts := getDbConfig(cfg)

	//ejecutamos migraciones
	config.RunMigrations(cfg.ToDBConfig())

	//configuramos S3
	s3 := config.S3ConfigService(cfg.ToS3Config())

	log.Printf("S3 Configured: Bucket %s ", s3.Bucket)
	_ = db.Exec("SELECT 1")
	r := httpserver.NewRouter(*s3, db, DBStarts, nil, nil, cfg.S3Region, "", cfg.ToMQConfig())

	addr := ":8082"
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		IdleTimeout:  120 * time.Second, // clave para warm path
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	//startWatchdog()

	log.Printf("API listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

}

func getDbConfig(cfg *config.SecretApp) (*gorm.DB, *gorm.DB) {
	db, DBStarts, err := config.NewPostgresDB(cfg.ToDBConfig())
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	log.Println("DB Connection Established")
	return db, DBStarts
}

func getSecrets(ctx context.Context) *config.SecretApp {
	cfg, err := config.LoadSecretManager(ctx)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	return cfg
}
