package database

import (
	"fmt"
	"time"

	"github.com/villageFlower/paypilot_dev_session_service/internal/models"
	"github.com/villageFlower/paypilot_dev_session_service/pkg/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize initializes the database connection
func Initialize(cfg *config.DatabaseConfig, log *zap.Logger) error {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Info("Database connection established")

	return nil
}

// AutoMigrate runs auto migration for all models
func AutoMigrate(log *zap.Logger) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Migrate dev session model
	err := DB.AutoMigrate(
		&models.Session{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Info("Database migration completed")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
