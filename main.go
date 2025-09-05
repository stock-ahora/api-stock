package main

import (
	"context"
	"log"
	"net/http"

	"github.com/stock-ahora/api-stock/internal/config"
	"github.com/stock-ahora/api-stock/internal/http"
)

func main() {

	ctx := context.Background()

	cfg, err := config.LoadSecretManager(ctx)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := config.NewPostgresDB(cfg.ToDBConfig())
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	log.Println("DB Connection Established")
	defer db.Commit()

	config.RunMigrations(cfg.ToDBConfig())

	r := httpserver.NewRouter()

	addr := ":8082"
	log.Printf("API listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}

}

/// Configuración de la base de datos
