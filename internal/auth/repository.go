package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles all database operations for users and API keys.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ---- User operations ----

// CreateUser inserts a new user record.
func (r *Repository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail retrieves a user by email address.
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by primary key.
func (r *Repository) GetUserByID(id uuid.UUID) (*User, error) {
	var user User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserPlan changes a user's subscription plan.
func (r *Repository) UpdateUserPlan(id uuid.UUID, plan PlanType) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("plan", plan).Error
}

// ---- API Key operations ----

// CreateAPIKey inserts a new API key record.
func (r *Repository) CreateAPIKey(key *APIKey) error {
	return r.db.Create(key).Error
}

// GetAPIKeyByHash retrieves a non-revoked, non-expired API key by its SHA-256 hash.
func (r *Repository) GetAPIKeyByHash(hash string) (*APIKey, error) {
	var key APIKey
	err := r.db.Preload("User").
		Where("key_hash = ? AND revoked = ?", hash, false).
		First(&key).Error
	if err != nil {
		return nil, err
	}

	// Check expiry
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}

	return &key, nil
}

// ListAPIKeysByUser retrieves all API keys for a given user, ordered by creation date.
func (r *Repository) ListAPIKeysByUser(userID uuid.UUID) ([]APIKey, error) {
	var keys []APIKey
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

// RevokeAPIKey marks an API key as revoked. Only the owning user can revoke their keys.
func (r *Repository) RevokeAPIKey(id, userID uuid.UUID) error {
	result := r.db.Model(&APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("revoked", true)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// TouchLastUsed updates the last_used_at timestamp for an API key.
func (r *Repository) TouchLastUsed(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&APIKey{}).Where("id = ?", id).Update("last_used_at", now).Error
}
