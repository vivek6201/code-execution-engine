package models

import (
	"time"

	"github.com/google/uuid"
)

// PlanType represents the subscription tier a user is on.
type PlanType string

const (
	PlanBasic    PlanType = "basic"
	PlanPro      PlanType = "pro"
	PlanUltimate PlanType = "ultimate"
)

// ValidPlan checks whether a given plan string is a recognized tier.
func ValidPlan(p PlanType) bool {
	switch p {
	case PlanBasic, PlanPro, PlanUltimate:
		return true
	}
	return false
}

// PlanLimits defines the rate-limiting constraints for a subscription tier.
type PlanLimits struct {
	RequestsPerDay int
	MaxConcurrent  int
}

// Plans maps each subscription tier to its limits.
var Plans = map[PlanType]PlanLimits{
	PlanBasic:    {RequestsPerDay: 100, MaxConcurrent: 2},
	PlanPro:      {RequestsPerDay: 1000, MaxConcurrent: 10},
	PlanUltimate: {RequestsPerDay: 10000, MaxConcurrent: 50},
}

// User represents a registered user in the system.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Name         string
	Plan         PlanType `gorm:"type:varchar(20);default:'basic'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	APIKeys      []APIKey
}
