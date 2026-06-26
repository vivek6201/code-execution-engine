package dashboard

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/code-execution-engine/internal/models"
	"github.com/google/uuid"
)

const (
	apiKeyPrefix    = "cee_"
	apiKeyRandBytes = 32 // 32 bytes = 64 hex chars
)

// Service implements the business logic for dashboard operations.
type Service struct {
	repo *Repository
}

// NewService creates a new dashboard service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetUser returns the public profile for a user.
func (s *Service) GetUser(userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	resp := ToUserResponse(user)
	return &resp, nil
}

// UpdatePlan changes a user's subscription plan.
func (s *Service) UpdatePlan(userID uuid.UUID, plan models.PlanType) error {
	if !models.ValidPlan(plan) {
		return errors.New("invalid plan")
	}
	return s.repo.UpdateUserPlan(userID, plan)
}

// GenerateAPIKey creates a new API key for a user.
// The raw key is returned exactly once; only its SHA-256 hash is persisted.
func (s *Service) GenerateAPIKey(userID uuid.UUID, name string) (*CreateAPIKeyResponse, error) {
	// Generate cryptographically random bytes
	rawBytes := make([]byte, apiKeyRandBytes)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	rawKey := apiKeyPrefix + hex.EncodeToString(rawBytes)
	keyHash := hashKey(rawKey)
	prefix := rawKey[:8] // "cee_" + first 4 hex chars

	apiKey := &models.APIKey{
		UserID:  userID,
		Name:    name,
		KeyHash: keyHash,
		Prefix:  prefix,
	}

	if err := s.repo.CreateAPIKey(apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return &CreateAPIKeyResponse{
		Key:     rawKey,
		Details: ToAPIKeyResponse(apiKey),
	}, nil
}

// ListAPIKeys returns all API keys for a user.
func (s *Service) ListAPIKeys(userID uuid.UUID) ([]APIKeyResponse, error) {
	keys, err := s.repo.ListAPIKeysByUser(userID)
	if err != nil {
		return nil, err
	}

	resp := make([]APIKeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = ToAPIKeyResponse(&k)
	}
	return resp, nil
}

// RevokeAPIKey marks an API key as revoked.
func (s *Service) RevokeAPIKey(userID, keyID uuid.UUID) error {
	return s.repo.RevokeAPIKey(keyID, userID)
}

// GetJobs returns a paginated list of job execution history for a user.
func (s *Service) GetJobs(userID uuid.UUID, page, limit int, lang, status, apiKeyID string) (*PaginatedJobsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	jobs, total, err := s.repo.ListJobsPaginated(userID, page, limit, lang, status, apiKeyID)
	if err != nil {
		return nil, err
	}

	return &PaginatedJobsResponse{
		Jobs:       jobs,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

// GetMetrics aggregates multi-dimensional usage data for dashboard visualization.
func (s *Service) GetMetrics(userID uuid.UUID) (*DashboardMetricsResponse, error) {
	statusDist, err := s.repo.GetStatusMetrics(userID)
	if err != nil {
		return nil, err
	}

	langDist, err := s.repo.GetLanguageMetrics(userID)
	if err != nil {
		return nil, err
	}

	perf, err := s.repo.GetPerformanceMetrics(userID)
	if err != nil {
		return nil, err
	}

	usage, err := s.repo.GetDailyUsageMetrics(userID)
	if err != nil {
		return nil, err
	}

	// Default to empty array if nil to keep JSON format clean
	if statusDist == nil {
		statusDist = []StatusMetric{}
	}
	if langDist == nil {
		langDist = []LanguageMetric{}
	}
	if perf == nil {
		perf = []PerformanceMetric{}
	}
	if usage == nil {
		usage = []DailyMetric{}
	}

	return &DashboardMetricsResponse{
		StatusDistribution: statusDist,
		LangDistribution:   langDist,
		Performance:        perf,
		UsageOverTime:      usage,
	}, nil
}

// hashKey computes the SHA-256 hex digest of a raw API key.
func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
