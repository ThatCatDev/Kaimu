package db

import (
	"fmt"
	"time"

	"github.com/thatcatdev/kaimu/backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

type DB struct {
	DB *gorm.DB
}

func NewDatabase(cfg config.DBConfig) *DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DataBase, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewTracedLogger(),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Add OpenTelemetry tracing and metrics plugin
	if err := db.Use(tracing.NewPlugin(
		tracing.WithDBName(cfg.DataBase),
	)); err != nil {
		panic(fmt.Sprintf("failed to add opentelemetry plugin: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database connection")
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(90 * time.Second)

	return &DB{DB: db}
}
