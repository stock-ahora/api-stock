package main

import (
	"context"
	"log"
	"net/http"
	"runtime"
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

	go func() {
		for range time.Tick(10 * time.Second) {
			buf := make([]byte, 1<<20)
			n := runtime.Stack(buf, true)
			log.Printf("\n\n=== STACK DUMP ===\n%s\n\n", buf[:n])
		}
	}()

	startWatchdog()

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

func startWatchdog() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:8082/api/v1/health", nil)
			c := http.Client{Timeout: 2 * time.Second}
			resp, err := c.Do(req)
			if err != nil {
				log.Printf("⚠ WATCHDOG: server unreachable locally: %v", err)
				continue
			}
			_ = resp.Body.Close()
			if resp.StatusCode != 200 {
				log.Printf("⚠ WATCHDOG: health returned %d", resp.StatusCode)
			}
		}
	}()
}
