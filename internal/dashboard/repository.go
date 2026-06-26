package dashboard

import (
	"time"

	"github.com/code-execution-engine/internal/judge"
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

// ListJobsPaginated retrieves historical execution runs for a user with pagination and filter criteria.
func (r *Repository) ListJobsPaginated(userID uuid.UUID, page, limit int, lang, status, apiKeyID string) ([]judge.JobRecord, int64, error) {
	var jobs []judge.JobRecord
	var total int64

	query := r.db.Model(&judge.JobRecord{}).Where("user_id = ?", userID)

	if lang != "" {
		query = query.Where("language = ?", lang)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if apiKeyID != "" {
		query = query.Where("api_key_id = ?", apiKeyID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&jobs).Error

	return jobs, total, err
}

// GetStatusMetrics fetches aggregate counts of jobs grouped by status.
func (r *Repository) GetStatusMetrics(userID uuid.UUID) ([]StatusMetric, error) {
	var metrics []StatusMetric
	err := r.db.Model(&judge.JobRecord{}).
		Select("status, count(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&metrics).Error
	return metrics, err
}

// GetLanguageMetrics fetches aggregate counts of jobs grouped by language.
func (r *Repository) GetLanguageMetrics(userID uuid.UUID) ([]LanguageMetric, error) {
	var metrics []LanguageMetric
	err := r.db.Model(&judge.JobRecord{}).
		Select("language, count(*) as count").
		Where("user_id = ?", userID).
		Group("language").
		Scan(&metrics).Error
	return metrics, err
}

// GetPerformanceMetrics fetches avg runtime time and memory usages per language for successful executions.
func (r *Repository) GetPerformanceMetrics(userID uuid.UUID) ([]PerformanceMetric, error) {
	var metrics []PerformanceMetric
	err := r.db.Model(&judge.JobRecord{}).
		Select("language, avg(time_ms) as avg_time_ms, avg(memory_kb) as avg_memory_kb, count(*) as count").
		Where("user_id = ? AND status = 'SUCCESS'", userID).
		Group("language").
		Scan(&metrics).Error
	return metrics, err
}

// GetDailyUsageMetrics fetches daily total submission volumes over the last 30 days.
func (r *Repository) GetDailyUsageMetrics(userID uuid.UUID) ([]DailyMetric, error) {
	var metrics []DailyMetric
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err := r.db.Model(&judge.JobRecord{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as date, count(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&metrics).Error
	return metrics, err
}
