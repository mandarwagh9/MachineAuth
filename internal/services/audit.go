package services

import (
	"time"

	"github.com/google/uuid"

	"machineauth/internal/db"
	"machineauth/internal/models"
)

type AuditService struct {
	jsonDB         db.Database
	webhookService *WebhookService
}

func NewAuditService(database db.Database) *AuditService {
	return &AuditService{jsonDB: database}
}

// SetWebhookService sets the webhook service for event triggering.
// Called after both services are created to avoid circular dependency.
func (s *AuditService) SetWebhookService(ws *WebhookService) {
	s.webhookService = ws
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

// triggerWebhook fires webhook events if webhookService is configured
func (s *AuditService) triggerWebhook(event string, payload interface{}) {
	if s.webhookService != nil {
		s.webhookService.TriggerEvent(event, payload)
	}
}

func (s *AuditService) LogAgentCreated(agent *models.Agent, ipAddress, userAgent string) error {
	err := s.Log(EventAgentCreated, &agent.ID, ipAddress, userAgent)
	s.triggerWebhook(EventAgentCreated, map[string]interface{}{
		"event":      EventAgentCreated,
		"agent_id":   agent.ID.String(),
		"agent_name": agent.Name,
		"client_id":  agent.ClientID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

func (s *AuditService) LogAgentDeleted(agentID uuid.UUID, ipAddress, userAgent string) error {
	err := s.Log(EventAgentDeleted, &agentID, ipAddress, userAgent)
	s.triggerWebhook(EventAgentDeleted, map[string]interface{}{
		"event":     EventAgentDeleted,
		"agent_id":  agentID.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

func (s *AuditService) LogAgentUpdated(agentID uuid.UUID, ipAddress, userAgent string) error {
	err := s.Log(EventAgentUpdated, &agentID, ipAddress, userAgent)
	s.triggerWebhook(EventAgentUpdated, map[string]interface{}{
		"event":     EventAgentUpdated,
		"agent_id":  agentID.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

func (s *AuditService) LogCredentialsRotated(agentID uuid.UUID, ipAddress, userAgent string) error {
	err := s.Log(EventAgentCredentialsRotated, &agentID, ipAddress, userAgent)
	s.triggerWebhook(EventAgentCredentialsRotated, map[string]interface{}{
		"event":     EventAgentCredentialsRotated,
		"agent_id":  agentID.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

func (s *AuditService) LogTokenIssued(agentID uuid.UUID, ipAddress, userAgent string) error {
	err := s.Log(EventTokenIssued, &agentID, ipAddress, userAgent)
	s.triggerWebhook(EventTokenIssued, map[string]interface{}{
		"event":     EventTokenIssued,
		"agent_id":  agentID.String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

func (s *AuditService) LogTokenValidation(agentID *uuid.UUID, success bool, ipAddress, userAgent string) error {
	action := EventTokenValidationSuccess
	if !success {
		action = EventTokenValidationFailed
	}
	err := s.Log(action, agentID, ipAddress, userAgent)

	payload := map[string]interface{}{
		"event":     action,
		"success":   success,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if agentID != nil {
		payload["agent_id"] = agentID.String()
	}
	s.triggerWebhook(action, payload)
	return err
}

// LogWebhook logs a webhook-related event and optionally triggers webhooks
func (s *AuditService) LogWebhook(action string, webhookID uuid.UUID, ipAddress, userAgent string) error {
	return s.Log(action, nil, ipAddress, userAgent)
}

// ListAuditLogs queries the audit log with filtering and pagination.
func (s *AuditService) ListAuditLogs(query models.AuditLogQuery) ([]models.AuditLog, int, error) {
	dbLogs, total, err := s.jsonDB.ListAuditLogs(query.AgentID, query.Action, query.IPAddress, query.From, query.To, query.Page, query.Limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]models.AuditLog, len(dbLogs))
	for i, l := range dbLogs {
		result[i] = models.AuditLog{
			ID:        uuid.MustParse(l.ID),
			Action:    l.Action,
			IPAddress: l.IPAddress,
			UserAgent: l.UserAgent,
			Details:   l.Details,
			CreatedAt: l.CreatedAt,
		}
		if l.AgentID != nil {
			aid := uuid.MustParse(*l.AgentID)
			result[i].AgentID = &aid
		}
	}
	return result, total, nil
}
