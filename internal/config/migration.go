package config

import (
	"fmt"
	"log"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations corre todas las migraciones pendientes
func RunMigrations(cfg DBConfig) {
	user := url.QueryEscape(cfg.User)
	pass := url.QueryEscape(cfg.Password)

	migrateURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, pass, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	m, err := migrate.New("file://internal/db/migrations", migrateURL)
	if err != nil {
		log.Fatalf("❌ Error creando migrator: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("❌ Error aplicando migraciones: %v", err)
	}

	log.Println("✅ Migraciones aplicadas correctamente")
}
