package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"agentauth/internal/db"
	"agentauth/internal/models"
)

type AgentService struct {
	db *db.DB
}

func NewAgentService(database *db.DB) *AgentService {
	return &AgentService{db: database}
}

func (s *AgentService) Create(req models.CreateAgentRequest) (*models.CreateAgentResponse, error) {
	id := uuid.New()
	clientID := uuid.New().String()
	clientSecret := generateSecureSecret(32)

	secretHash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash secret: %w", err)
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expTime := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Second)
		expiresAt = &expTime
	}

	scopes := req.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	agent := db.Agent{
		ID:               id.String(),
		Name:             req.Name,
		ClientID:         clientID,
		ClientSecretHash: string(secretHash),
		Scopes:           scopes,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ExpiresAt:        expiresAt,
	}

	if err := s.db.CreateAgent(agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &models.CreateAgentResponse{
		Agent: models.Agent{
			ID:        id,
			Name:      req.Name,
			ClientID:  clientID,
			Scopes:    scopes,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: expiresAt,
		},
		ClientSecret: clientSecret,
		ClientID:     clientID,
	}, nil
}

func (s *AgentService) GetByClientID(clientID string) (*models.Agent, error) {
	agent, err := s.db.GetAgentByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}
	return toModelAgent(agent), nil
}

func (s *AgentService) GetByID(id uuid.UUID) (*models.Agent, error) {
	agent, err := s.db.GetAgentByID(id.String())
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}
	return toModelAgent(agent), nil
}

func (s *AgentService) List() ([]models.Agent, error) {
	agents, err := s.db.ListAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	result := make([]models.Agent, len(agents))
	for i, a := range agents {
		result[i] = *toModelAgent(&a)
	}
	return result, nil
}

func (s *AgentService) Delete(id uuid.UUID) error {
	return s.db.DeleteAgent(id.String())
}

func (s *AgentService) RotateCredentials(id uuid.UUID) (string, error) {
	newSecret := generateSecureSecret(32)

	secretHash, err := bcrypt.GenerateFromPassword([]byte(newSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}

	err = s.db.UpdateAgent(id.String(), func(agent *db.Agent) error {
		agent.ClientSecretHash = string(secretHash)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to rotate credentials: %w", err)
	}

	return newSecret, nil
}

func (s *AgentService) ValidateCredentials(clientID, clientSecret string) (*models.Agent, error) {
	agent, err := s.GetByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !agent.IsActive {
		return nil, fmt.Errorf("agent is inactive")
	}

	if agent.ExpiresAt != nil && time.Now().After(*agent.ExpiresAt) {
		return nil, fmt.Errorf("agent has expired")
	}

	dbAgent, err := s.db.GetAgentByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbAgent.ClientSecretHash), []byte(clientSecret)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return agent, nil
}

func toModelAgent(a *db.Agent) *models.Agent {
	return &models.Agent{
		ID:               uuid.MustParse(a.ID),
		Name:             a.Name,
		ClientID:         a.ClientID,
		ClientSecretHash: a.ClientSecretHash,
		Scopes:           a.Scopes,
		PublicKey:        a.PublicKey,
		IsActive:         a.IsActive,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
		ExpiresAt:        a.ExpiresAt,
	}
}

func generateSecureSecret(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func HashClientID(clientID string) string {
	hash := sha256.Sum256([]byte(clientID))
	return base64.URLEncoding.EncodeToString(hash[:])
}
