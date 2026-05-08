package db

import (
	"github.com/code-execution-engine/internal/logger"
	"github.com/code-execution-engine/internal/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB opens a connection to Postgres and auto-migrates all models.
func ConnectDB(dbUrl string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		panic("database connection failed")
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(&auth.User{}, &auth.APIKey{}); err != nil {
		logger.Error("Failed to run auto-migration", "error", err)
		panic("database migration failed")
	}

	logger.Info("Database connected and migrated")
	return db
}
