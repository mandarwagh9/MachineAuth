package services

import (
	"time"

	"github.com/google/uuid"

	"agentauth/internal/db"
	"agentauth/internal/models"
)

type AuditService struct {
	jsonDB *db.DB
}

func NewAuditService(database *db.DB) *AuditService {
	return &AuditService{jsonDB: database}
}

func (s *AuditService) Log(action string, agentID *uuid.UUID, ipAddress, userAgent string) error {
	log := db.AuditLog{
		ID:        uuid.New().String(),
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}
	if agentID != nil {
		log.AgentID = new(string)
		*log.AgentID = agentID.String()
	}
	return s.jsonDB.AddAuditLog(log)
}

func (s *AuditService) LogAgentCreated(agent *models.Agent, ipAddress, userAgent string) error {
	return s.Log("agent.created", &agent.ID, ipAddress, userAgent)
}

func (s *AuditService) LogAgentDeleted(agentID uuid.UUID, ipAddress, userAgent string) error {
	return s.Log("agent.deleted", &agentID, ipAddress, userAgent)
}

func (s *AuditService) LogCredentialsRotated(agentID uuid.UUID, ipAddress, userAgent string) error {
	return s.Log("agent.credentials_rotated", &agentID, ipAddress, userAgent)
}

func (s *AuditService) LogTokenIssued(agentID uuid.UUID, ipAddress, userAgent string) error {
	return s.Log("token.issued", &agentID, ipAddress, userAgent)
}

func (s *AuditService) LogTokenValidation(agentID *uuid.UUID, success bool, ipAddress, userAgent string) error {
	action := "token.validation_success"
	if !success {
		action = "token.validation_failed"
	}
	return s.Log(action, agentID, ipAddress, userAgent)
}
