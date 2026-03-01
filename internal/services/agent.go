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

	var teamID *string
	if req.TeamID != nil {
		tid := req.TeamID.String()
		teamID = &tid
	}

	agent := db.Agent{
		ID:               id.String(),
		OrganizationID:   req.OrganizationID,
		TeamID:           teamID,
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

	var teamUUID *uuid.UUID
	if teamID != nil {
		tu, _ := uuid.Parse(*teamID)
		teamUUID = &tu
	}

	return &models.CreateAgentResponse{
		Agent: models.Agent{
			ID:             id,
			OrganizationID: req.OrganizationID,
			TeamID:         teamUUID,
			Name:           req.Name,
			ClientID:       clientID,
			Scopes:         scopes,
			IsActive:       true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			ExpiresAt:      expiresAt,
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

func (s *AgentService) ListByOrganization(orgID string) ([]models.Agent, error) {
	agents, err := s.db.ListAgentsByOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	result := make([]models.Agent, len(agents))
	for i, a := range agents {
		result[i] = *toModelAgent(&a)
	}
	return result, nil
}

func (s *AgentService) ListByTeam(teamID string) ([]models.Agent, error) {
	agents, err := s.db.ListAgentsByTeam(teamID)
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

func (s *AgentService) Count() (int, error) {
	agents, err := s.db.ListAgents()
	if err != nil {
		return 0, err
	}
	return len(agents), nil
}

type Stats struct {
	TotalRequests int
	TokensIssued  int
	ActiveTokens  int
	TotalAgents   int
}

func (s *AgentService) GetStats() (*Stats, error) {
	agents, err := s.db.ListAgents()
	if err != nil {
		return nil, err
	}

	totalTokens := 0
	totalRefreshes := 0
	for _, a := range agents {
		totalTokens += a.TokenCount
		totalRefreshes += a.RefreshCount
	}

	return &Stats{
		TotalRequests: 0,
		TokensIssued:  totalTokens,
		ActiveTokens:  totalTokens - totalRefreshes,
		TotalAgents:   len(agents),
	}, nil
}

func (s *AgentService) RotateCredentials(id uuid.UUID) (string, error) {
	newSecret := generateSecureSecret(32)

	secretHash, err := bcrypt.GenerateFromPassword([]byte(newSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}

	now := time.Now()
	err = s.db.UpdateAgent(id.String(), func(agent *db.Agent) error {
		agent.ClientSecretHash = string(secretHash)
		agent.RotationHistory = append(agent.RotationHistory, db.Rotation{
			RotatedAt: now,
		})
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

	s.RecordActivity(agent.ID)

	return agent, nil
}

func (s *AgentService) RecordActivity(agentID uuid.UUID) error {
	now := time.Now()
	return s.db.UpdateAgent(agentID.String(), func(agent *db.Agent) error {
		agent.LastActivityAt = &now
		return nil
	})
}

func (s *AgentService) RecordTokenIssuance(agentID uuid.UUID) error {
	now := time.Now()
	return s.db.UpdateAgent(agentID.String(), func(agent *db.Agent) error {
		agent.TokenCount++
		agent.LastTokenIssuedAt = &now
		agent.LastActivityAt = &now
		return nil
	})
}

func (s *AgentService) RecordTokenRefresh(agentID uuid.UUID) error {
	now := time.Now()
	return s.db.UpdateAgent(agentID.String(), func(agent *db.Agent) error {
		agent.RefreshCount++
		agent.LastActivityAt = &now
		return nil
	})
}

func (s *AgentService) GetUsage(agentID uuid.UUID) (*models.AgentUsage, error) {
	agent, err := s.db.GetAgentByID(agentID.String())
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	rotationHistory := make([]models.Rotation, len(agent.RotationHistory))
	for i, r := range agent.RotationHistory {
		rotationHistory[i] = models.Rotation{
			RotatedAt:   r.RotatedAt,
			RotatedByIP: r.RotatedByIP,
		}
	}

	return &models.AgentUsage{
		Agent:             *toModelAgent(agent),
		TokenCount:        agent.TokenCount,
		RefreshCount:      agent.RefreshCount,
		LastActivityAt:    agent.LastActivityAt,
		LastTokenIssuedAt: agent.LastTokenIssuedAt,
		RotationHistory:   rotationHistory,
	}, nil
}

func (s *AgentService) Deactivate(id uuid.UUID) error {
	return s.db.UpdateAgent(id.String(), func(agent *db.Agent) error {
		agent.IsActive = false
		return nil
	})
}

func (s *AgentService) Reactivate(id uuid.UUID) error {
	return s.db.UpdateAgent(id.String(), func(agent *db.Agent) error {
		agent.IsActive = true
		return nil
	})
}

func toModelAgent(a *db.Agent) *models.Agent {
	var teamUUID *uuid.UUID
	if a.TeamID != nil {
		tu, _ := uuid.Parse(*a.TeamID)
		teamUUID = &tu
	}

	return &models.Agent{
		ID:                uuid.MustParse(a.ID),
		OrganizationID:    a.OrganizationID,
		TeamID:            teamUUID,
		Name:              a.Name,
		ClientID:          a.ClientID,
		ClientSecretHash:  a.ClientSecretHash,
		Scopes:            a.Scopes,
		PublicKey:         a.PublicKey,
		IsActive:          a.IsActive,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
		ExpiresAt:         a.ExpiresAt,
		TokenCount:        a.TokenCount,
		RefreshCount:      a.RefreshCount,
		LastActivityAt:    a.LastActivityAt,
		LastTokenIssuedAt: a.LastTokenIssuedAt,
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
