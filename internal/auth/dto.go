package auth

import (
	"github.com/code-execution-engine/internal/models"
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

// ---- Responses ----

// UserResponse is the public representation of a user (no password hash).
type UserResponse struct {
	ID    uuid.UUID       `json:"id"`
	Email string          `json:"email"`
	Name  string          `json:"name"`
	Plan  models.PlanType `json:"plan"`
}

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ToUserResponse converts a User model to its public representation.
func ToUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Plan:  u.Plan,
	}
}
