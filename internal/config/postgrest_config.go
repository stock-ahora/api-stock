package config

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var dsn string
var mu sync.Mutex

func NewPostgresDB(cfg DBConfig) (*gorm.DB, error) {

	portInt := cfg.Port

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Host, portInt, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		//		Logger: logger.Default.LogMode(logger.Warn), // nivel de logging
		//todo cambiar en algun momentos para profecionalizar la wea de los logs
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

func ReconnectDB() error {
	mu.Lock()
	defer mu.Unlock()

	sqlDB, _ := db.DB()
	sqlDB.Close() // cerramos pool actual

	newDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	db = newDb
	return nil
}

func PingDbWithRetry(ctx context.Context) error {
	sqlDB, _ := db.DB()

	// primer intento
	if err := sqlDB.PingContext(ctx); err == nil {
		return nil
	}

	log.Println("⚠ DB ping falló — intentando reconectar...")
	if err := ReconnectDB(); err != nil {
		return err
	}

	// segundo intento después de reconectar
	sqlDB, _ = db.DB()
	return sqlDB.PingContext(ctx)
}
