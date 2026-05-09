package db

import (
	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/models"
	"github.com/code-execution-engine/pkg/telemetry"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB opens a connection to Postgres and auto-migrates all models.
func ConnectDB(dbUrl string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		telemetry.Error("Failed to connect to database", "error", err)
		panic("database connection failed")
	}

	// Auto-migrate all models
	if err = db.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&judge.JobRecord{},
	); err != nil {
		telemetry.Error("Failed to run auto-migration", "error", err)
		panic("database migration failed")
	}

	telemetry.Info("Database connected and migrated")
	return db
}
