package models

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key issued to a user for authenticating judge API requests.
// Only the SHA-256 hash of the raw key is stored; the raw key is shown once at creation.
type APIKey struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	User       User      `gorm:"constraint:OnDelete:CASCADE;"`
	Name       string    // user-given label, e.g. "production"
	KeyHash    string    `gorm:"uniqueIndex;not null"` // SHA-256 hex digest
	Prefix     string    `gorm:"not null"`             // first 8 chars for display, e.g. "cee_a1b2"
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	Revoked    bool `gorm:"default:false"`
	CreatedAt  time.Time
}
