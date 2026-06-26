package dashboard

import (
	"time"

	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/models"
	"github.com/google/uuid"
)

// ---- Requests ----

type UpdatePlanRequest struct {
	Plan models.PlanType `json:"plan" binding:"required"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// ---- Responses ----

type UserResponse struct {
	ID    uuid.UUID       `json:"id"`
	Email string          `json:"email"`
	Name  string          `json:"name"`
	Plan  models.PlanType `json:"plan"`
}

type APIKeyResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	Revoked    bool       `json:"revoked"`
}

type CreateAPIKeyResponse struct {
	Key     string         `json:"key"` // The raw key (only returned once)
	Details APIKeyResponse `json:"details"`
}

type PaginatedJobsResponse struct {
	Jobs       []judge.JobRecord `json:"jobs"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
}

type StatusMetric struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type LanguageMetric struct {
	Language string `json:"language"`
	Count    int64  `json:"count"`
}

type PerformanceMetric struct {
	Language string  `json:"language"`
	AvgTime  float64 `json:"avg_time_ms"`
	AvgMem   float64 `json:"avg_memory_kb"`
	Count    int64   `json:"count"`
}

type DailyMetric struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type DashboardMetricsResponse struct {
	StatusDistribution []StatusMetric      `json:"status_distribution"`
	LangDistribution   []LanguageMetric    `json:"lang_distribution"`
	Performance        []PerformanceMetric `json:"performance"`
	UsageOverTime      []DailyMetric       `json:"usage_over_time"`
}

// ---- Mappers ----

func ToUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Plan:  u.Plan,
	}
}

func ToAPIKeyResponse(k *models.APIKey) APIKeyResponse {
	return APIKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		Prefix:     k.Prefix,
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
		Revoked:    k.Revoked,
	}
}
