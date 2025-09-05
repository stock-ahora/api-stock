package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(cfg DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // nivel de logging
	})
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}

	// Configurar pool de conexiones
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)                 // conexiones máximas abiertas
	sqlDB.SetMaxIdleConns(25)                 // conexiones en reposo
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // tiempo máximo de vida de una conexión

	log.Println("✅ Conectado a PostgreSQL correctamente")
	return db, nil
}
