package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stock-ahora/api-stock/internal/utils"
)

// RunMigrations corre todas las migraciones pendientes
func RunMigrations(cfg DBConfig) {
	user := url.QueryEscape(cfg.User)
	pass := url.QueryEscape(cfg.Password)
	portInt, err := utils.ConverToint(cfg.Port)

	migrateURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, pass, cfg.Host, portInt, cfg.DBName, cfg.SSLMode,
	)

	path, _ := os.Getwd()
	migrationsPath := fmt.Sprintf("file://%s/internal/db/migrations", strings.ReplaceAll(path, "\\", "/"))

	m, err := migrate.New(migrationsPath, migrateURL)
	if err != nil {
		if err != nil {
			log.Fatalf("❌ Error creando migrator: %v", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("❌ Error aplicando migraciones: %v", err)
		}

		log.Println("✅ Migraciones aplicadas correctamente")
	}
}
