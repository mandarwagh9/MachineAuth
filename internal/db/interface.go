package db

import "time"

// Database defines the storage interface that all backends must implement.
// Both JSONDB (development) and PostgresDB (production) satisfy this.
type Database interface {
	Close() error

	// ── Agents ─────────────────────────────────────────────────────────

	CreateAgent(agent Agent) error
	GetAgentByClientID(clientID string) (*Agent, error)
	GetAgentByID(id string) (*Agent, error)
	ListAgents() ([]Agent, error)
	ListAgentsByOrganization(orgID string) ([]Agent, error)
	ListAgentsByTeam(teamID string) ([]Agent, error)
	UpdateAgent(id string, updateFn func(*Agent) error) error
	DeleteAgent(id string) error
	CountAgents() (int, error)
	ListAgentsPaginated(search, status, orgID, sort string, page, limit int) ([]Agent, int, error)

	// ── Audit Logs ─────────────────────────────────────────────────────

	AddAuditLog(log AuditLog) error
	ListAuditLogs(agentID, action, ipAddress string, from, to *time.Time, page, limit int) ([]AuditLog, int, error)

	// ── Refresh Tokens ─────────────────────────────────────────────────

	CreateRefreshToken(rt RefreshToken) error
	GetRefreshToken(id string) (*RefreshToken, error)
	RevokeRefreshToken(id string) error

	// ── Revoked Tokens ─────────────────────────────────────────────────

	AddRevokedToken(rt RevokedToken) error
	IsTokenRevoked(jti string) bool

	// ── Metrics ────────────────────────────────────────────────────────

	IncrementTokensRefreshed() error
	IncrementTokensRevoked() error
	GetMetrics() Metrics

	// ── Organizations ──────────────────────────────────────────────────

	CreateOrganization(org Organization) error
	GetOrganization(id string) (*Organization, error)
	GetOrganizationBySlug(slug string) (*Organization, error)
	ListOrganizations() ([]Organization, error)
	UpdateOrganization(id string, updateFn func(*Organization) error) error
	DeleteOrganization(id string) error

	// ── Teams ──────────────────────────────────────────────────────────

	CreateTeam(team Team) error
	GetTeam(id string) (*Team, error)
	ListTeamsByOrganization(orgID string) ([]Team, error)
	UpdateTeam(id string, updateFn func(*Team) error) error
	DeleteTeam(id string) error

	// ── API Keys ───────────────────────────────────────────────────────

	CreateAPIKey(key APIKey) error
	GetAPIKeyByID(id string) (*APIKey, error)
	GetAPIKeyByKeyHash(keyHash string) (*APIKey, error)
	ListAPIKeysByOrganization(orgID string) ([]APIKey, error)
	UpdateAPIKey(id string, updateFn func(*APIKey) error) error
	DeleteAPIKey(id string) error

	// ── Webhooks ───────────────────────────────────────────────────────

	CreateWebhook(webhook WebhookConfig) error
	GetWebhook(id string) (*WebhookConfig, error)
	ListWebhooks() ([]WebhookConfig, error)
	UpdateWebhook(id string, updateFn func(*WebhookConfig) error) error
	DeleteWebhook(id string) error
	ListActiveWebhooksForEvent(event string) ([]WebhookConfig, error)

	// ── Webhook Deliveries ─────────────────────────────────────────────

	AddWebhookDelivery(delivery WebhookDelivery) error
	GetWebhookDelivery(id string) (*WebhookDelivery, error)
	UpdateWebhookDelivery(id string, updateFn func(*WebhookDelivery) error) error
	ListWebhookDeliveries(webhookConfigID string) ([]WebhookDelivery, error)
	ListPendingDeliveries() ([]WebhookDelivery, error)

	// ── Admin Users ────────────────────────────────────────────────────

	CreateAdminUser(user AdminUser) error
	GetAdminUserByEmail(email string) (*AdminUser, error)
	GetAdminUserByID(id string) (*AdminUser, error)
	ListAdminUsers() ([]AdminUser, error)
}
