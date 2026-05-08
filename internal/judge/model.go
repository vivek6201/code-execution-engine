package judge

import (
	"time"

	"github.com/code-execution-engine/internal/types"
	"github.com/google/uuid"
)

// JobRecord represents a single execution run persisted in the database.
type JobRecord struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	APIKeyID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Language   string    `gorm:"not null"`
	Code       string    `gorm:"not null"`
	Status     types.Status
	Output     string
	Error      string
	FatalError string
	TimeMs     int64
	MemoryKB   int64
	Total      int
	Passed     int
	TestCases  []types.TestCaseResult `gorm:"serializer:json"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
