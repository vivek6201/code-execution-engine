package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	apiKeyPrefix    = "cee_"
	apiKeyRandBytes = 32 // 32 bytes = 64 hex chars
)

// Service implements the business logic for authentication and API key management.
type Service struct {
	repo      *Repository
	jwtSecret string
}

// NewService creates a new auth service.
func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

// Register creates a new user account.
func (s *Service) Register(req RegisterRequest) (*UserResponse, error) {
	// Check if email already exists
	existing, _ := s.repo.GetUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		Plan:         PlanBasic,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	resp := ToUserResponse(user)
	return &resp, nil
}

// Login authenticates a user and returns a JWT.
func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := GenerateToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User:  ToUserResponse(user),
	}, nil
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
func (s *Service) UpdatePlan(userID uuid.UUID, plan PlanType) error {
	if !ValidPlan(plan) {
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

	apiKey := &APIKey{
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

// ValidateAPIKey checks if a raw API key is valid and returns the associated user and key.
func (s *Service) ValidateAPIKey(rawKey string) (*User, *APIKey, error) {
	keyHash := hashKey(rawKey)

	apiKey, err := s.repo.GetAPIKeyByHash(keyHash)
	if err != nil {
		return nil, nil, errors.New("invalid API key")
	}

	// Touch last used asynchronously — fire and forget
	go func() {
		_ = s.repo.TouchLastUsed(apiKey.ID)
	}()

	return &apiKey.User, apiKey, nil
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

// JWTSecret returns the JWT secret for use by middlewares.
func (s *Service) JWTSecret() string {
	return s.jwtSecret
}

// hashKey computes the SHA-256 hex digest of a raw API key.
func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
