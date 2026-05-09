package auth

import (
	"time"

	"github.com/code-execution-engine/internal/models"
	"gorm.io/gorm"
	"github.com/google/uuid"
)

// Repository handles database operations for authentication.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user record.
func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail retrieves a user by email address.
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAPIKeyByHash retrieves a non-revoked, non-expired API key by its SHA-256 hash.
func (r *Repository) GetAPIKeyByHash(hash string) (*models.APIKey, error) {
	var key models.APIKey
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

// TouchLastUsed updates the last_used_at timestamp for an API key.
func (r *Repository) TouchLastUsed(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("last_used_at", now).Error
}
