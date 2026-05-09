package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/code-execution-engine/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Service implements the business logic for authentication.
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

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		Plan:         models.PlanBasic,
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

// ValidateAPIKey checks if a raw API key is valid and returns the associated user and key.
func (s *Service) ValidateAPIKey(rawKey string) (*models.User, *models.APIKey, error) {
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

// JWTSecret returns the JWT secret for use by middlewares.
func (s *Service) JWTSecret() string {
	return s.jwtSecret
}

// hashKey computes the SHA-256 hex digest of a raw API key.
func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
