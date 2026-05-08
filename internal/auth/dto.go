package auth

import (
	"time"

	"github.com/google/uuid"
)

// ---- Requests ----

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateAPIKeyRequest is the payload for generating a new API key.
type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdatePlanRequest is the payload for changing a user's subscription plan.
type UpdatePlanRequest struct {
	Plan PlanType `json:"plan" binding:"required"`
}

// ---- Responses ----

// UserResponse is the public representation of a user (no password hash).
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
	Plan  PlanType  `json:"plan"`
}

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// APIKeyResponse is the public representation of an API key (no hash).
type APIKeyResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	Revoked    bool       `json:"revoked"`
}

// CreateAPIKeyResponse is returned when a new API key is generated.
// The Key field contains the raw key and is shown exactly once.
type CreateAPIKeyResponse struct {
	Key     string         `json:"key"`
	Details APIKeyResponse `json:"details"`
}

// ToUserResponse converts a User model to its public representation.
func ToUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Plan:  u.Plan,
	}
}

// ToAPIKeyResponse converts an APIKey model to its public representation.
func ToAPIKeyResponse(k *APIKey) APIKeyResponse {
	return APIKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		Prefix:     k.Prefix,
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
		Revoked:    k.Revoked,
	}
}
