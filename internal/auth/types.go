// Package auth implements user authentication, API key management,
// and subscription plan enforcement for the code execution engine.
package auth

import (
	"github.com/golang-jwt/jwt/v5"
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

// JWTClaims is the payload embedded in dashboard session tokens.
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}
