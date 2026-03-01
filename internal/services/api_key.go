package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"machineauth/internal/db"
	"machineauth/internal/models"
)

type APIKeyService struct {
	db *db.DB
}

func NewAPIKeyService(database *db.DB) *APIKeyService {
	return &APIKeyService{db: database}
}

func (s *APIKeyService) Create(orgID string, req models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	id := uuid.New()
	key := generateAPIKey()
	keyHash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	prefix := key[:12]

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expTime := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Second)
		expiresAt = &expTime
	}

	var teamID *string
	if req.TeamID != nil {
		tid := req.TeamID.String()
		teamID = &tid
	}

	apiKey := db.APIKey{
		ID:             id.String(),
		OrganizationID: orgID,
		TeamID:         teamID,
		Name:           req.Name,
		KeyHash:        string(keyHash),
		Prefix:         prefix,
		IsActive:       true,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
	}

	if err := s.db.CreateAPIKey(apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	var teamUUID *uuid.UUID
	if teamID != nil {
		tu, _ := uuid.Parse(*teamID)
		teamUUID = &tu
	}

	return &models.CreateAPIKeyResponse{
		APIKey: models.APIKey{
			ID:             id.String(),
			OrganizationID: orgID,
			TeamID:         teamUUID,
			Name:           req.Name,
			Prefix:         prefix,
			IsActive:       true,
			ExpiresAt:      expiresAt,
			CreatedAt:      time.Now(),
		},
		Key: key,
	}, nil
}

func (s *APIKeyService) GetByID(id string) (*models.APIKey, error) {
	key, err := s.db.GetAPIKeyByID(id)
	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}
	return toModelAPIKey(key), nil
}

func (s *APIKeyService) Validate(key string) (*models.APIKey, error) {
	keyHash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	apiKey, err := s.db.GetAPIKeyByKeyHash(string(keyHash))
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	s.db.UpdateAPIKey(apiKey.ID, func(k *db.APIKey) error {
		now := time.Now()
		k.LastUsedAt = &now
		return nil
	})

	return toModelAPIKey(apiKey), nil
}

func (s *APIKeyService) ListByOrganization(orgID string) ([]models.APIKey, error) {
	keys, err := s.db.ListAPIKeysByOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	result := make([]models.APIKey, len(keys))
	for i, k := range keys {
		result[i] = *toModelAPIKey(&k)
	}
	return result, nil
}

func (s *APIKeyService) Revoke(id string) error {
	return s.db.DeleteAPIKey(id)
}

func toModelAPIKey(k *db.APIKey) *models.APIKey {
	var teamUUID *uuid.UUID
	if k.TeamID != nil {
		tu, _ := uuid.Parse(*k.TeamID)
		teamUUID = &tu
	}

	return &models.APIKey{
		ID:             k.ID,
		OrganizationID: k.OrganizationID,
		TeamID:         teamUUID,
		Name:           k.Name,
		Prefix:         k.Prefix,
		LastUsedAt:     k.LastUsedAt,
		ExpiresAt:      k.ExpiresAt,
		IsActive:       k.IsActive,
		CreatedAt:      k.CreatedAt,
	}
}

func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "sk_" + base64.URLEncoding.EncodeToString(b)[:43]
}

func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return base64.URLEncoding.EncodeToString(hash[:])
}
