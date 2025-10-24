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
	db := getDbConfig(cfg)

	//ejecutamos migraciones
	config.RunMigrations(cfg.ToDBConfig())

	//configuramos S3
	s3 := config.S3ConfigService(cfg.ToS3Config())

	log.Printf("S3 Configured: Bucket %s ", s3.Bucket)
	_ = db.Exec("SELECT 1")
	r := httpserver.NewRouter(*s3, db, nil, nil, cfg.S3Region, "", cfg.ToMQConfig())

	addr := ":8082"
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		IdleTimeout:  120 * time.Second, // clave para warm path
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	go warmUp("https://xe1qfkl3f5.execute-api.us-east-2.amazonaws.com/prod/api/v1/stock/request")

	log.Printf("API listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

func warmUp(albURL string) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		resp, err := client.Get(albURL)
		if err != nil {
			log.Println("warmup error:", err)
			continue
		}
		_ = resp.Body.Close()
	}
}
