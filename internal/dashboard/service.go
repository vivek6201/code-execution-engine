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

// hashKey computes the SHA-256 hex digest of a raw API key.
func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
