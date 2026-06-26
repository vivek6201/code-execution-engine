package db

import (
	"github.com/code-execution-engine/pkg/telemetry"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB opens a connection to Postgres.
func ConnectDB(dbUrl string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		telemetry.Error("Failed to connect to database", "error", err)
		panic("database connection failed")
	}

	telemetry.Info("Database connected")
	return db
}
