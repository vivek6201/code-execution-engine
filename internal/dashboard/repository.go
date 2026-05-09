package dashboard

import (
	"github.com/code-execution-engine/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles database operations for dashboard features.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new dashboard repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetUserByID retrieves a user by primary key.
func (r *Repository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserPlan changes a user's subscription plan.
func (r *Repository) UpdateUserPlan(id uuid.UUID, plan models.PlanType) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("plan", plan).Error
}

// CreateAPIKey inserts a new API key record.
func (r *Repository) CreateAPIKey(key *models.APIKey) error {
	return r.db.Create(key).Error
}

// ListAPIKeysByUser retrieves all API keys for a given user, ordered by creation date.
func (r *Repository) ListAPIKeysByUser(userID uuid.UUID) ([]models.APIKey, error) {
	var keys []models.APIKey
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

// RevokeAPIKey marks an API key as revoked. Only the owning user can revoke their keys.
func (r *Repository) RevokeAPIKey(id, userID uuid.UUID) error {
	result := r.db.Model(&models.APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("revoked", true)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
