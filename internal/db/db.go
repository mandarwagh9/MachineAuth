package db

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type JSONDB struct {
	mu                sync.RWMutex
	filename          string
	Agents            []Agent           `json:"agents"`
	AdminUsers        []AdminUser       `json:"admin_users"`
	AuditLogs         []AuditLog        `json:"audit_logs"`
	RefreshTokens     []RefreshToken    `json:"refresh_tokens"`
	RevokedTokens     []RevokedToken    `json:"revoked_tokens"`
	Metrics           Metrics           `json:"metrics"`
	Organizations     []Organization    `json:"organizations"`
	Teams             []Team            `json:"teams"`
	APIKeys           []APIKey          `json:"api_keys"`
	WebhookConfigs    []WebhookConfig   `json:"webhook_configs"`
	WebhookDeliveries []WebhookDelivery `json:"webhook_deliveries"`
	OrgMembers        []OrgMember       `json:"org_members"`
	OrgSigningKeys    []OrgSigningKey   `json:"org_signing_keys"`
}

// AdminUser stored in JSON DB.
type AdminUser struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Metrics struct {
	TokensRefreshed int64 `json:"tokens_refreshed"`
	TokensRevoked   int64 `json:"tokens_revoked"`
}

type Organization struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	OwnerEmail     string    `json:"owner_email"`
	JWTIssuer      string    `json:"jwt_issuer"`
	JWTExpirySecs  int       `json:"jwt_expiry_secs"`
	AllowedOrigins string    `json:"allowed_origins"`
	Plan           string    `json:"plan"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Team struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type APIKey struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	TeamID         *string    `json:"team_id,omitempty"`
	Name           string     `json:"name"`
	KeyHash        string     `json:"key_hash"`
	Prefix         string     `json:"prefix"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Agent struct {
	ID                string                 `json:"id"`
	OrganizationID    string                 `json:"organization_id"`
	TeamID            *string                `json:"team_id,omitempty"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ClientID          string                 `json:"client_id"`
	ClientSecretHash  string                 `json:"client_secret_hash"`
	Scopes            []string               `json:"scopes"`
	PublicKey         *string                `json:"public_key,omitempty"`
	Status            string                 `json:"status"`
	IsActive          bool                   `json:"is_active"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
	TokenCount        int                    `json:"token_count"`
	RefreshCount      int                    `json:"refresh_count"`
	LastActivityAt    *time.Time             `json:"last_activity_at,omitempty"`
	LastTokenIssuedAt *time.Time             `json:"last_token_issued_at,omitempty"`
	RotationHistory   []Rotation             `json:"rotation_history"`
}

type Rotation struct {
	RotatedAt   time.Time `json:"rotated_at"`
	RotatedByIP string    `json:"rotated_by_ip,omitempty"`
}

type RefreshToken struct {
	ID        string     `json:"id"`
	AgentID   string     `json:"agent_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type RevokedToken struct {
	JTI     string    `json:"jti"`
	Expires time.Time `json:"expires"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	AgentID   *string   `json:"agent_id,omitempty"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type WebhookConfig struct {
	ID               string     `json:"id"`
	OrganizationID   string     `json:"organization_id,omitempty"`
	TeamID           string     `json:"team_id,omitempty"`
	Name             string     `json:"name"`
	URL              string     `json:"url"`
	Secret           string     `json:"secret"`
	Events           []string   `json:"events"`
	IsActive         bool       `json:"is_active"`
	MaxRetries       int        `json:"max_retries"`
	RetryBackoffBase int        `json:"retry_backoff_base"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastTestedAt     *time.Time `json:"last_tested_at,omitempty"`
	ConsecutiveFails int        `json:"consecutive_fails"`
}

// OrgMember represents a user's membership in an organization.
type OrgMember struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"` // owner, admin, member, viewer
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OrgSigningKey stores per-org RSA signing keys.
type OrgSigningKey struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	KeyID          string     `json:"key_id"`
	PublicKeyPEM   string     `json:"public_key_pem"`
	PrivateKeyPEM  string     `json:"private_key_pem"`
	Algorithm      string     `json:"algorithm"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type WebhookDelivery struct {
	ID              string     `json:"id"`
	WebhookConfigID string     `json:"webhook_config_id"`
	Event           string     `json:"event"`
	Payload         string     `json:"payload"`
	Headers         string     `json:"headers,omitempty"`
	Status          string     `json:"status"`
	Attempts        int        `json:"attempts"`
	LastAttemptAt   *time.Time `json:"last_attempt_at,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
	NextRetryAt     *time.Time `json:"next_retry_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type DB struct {
	*JSONDB
}

// Connect returns a Database backed by PostgreSQL or JSON file.
// postgres:// or postgresql:// → PostgresDB, json:filename → JSONDB.
func Connect(databaseURL string) (Database, error) {
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return connectPostgres(databaseURL)
	}
	if strings.HasPrefix(databaseURL, "json:") || strings.HasSuffix(databaseURL, ".json") {
		return connectJSON(databaseURL)
	}
	// Default to JSON
	return connectJSON("json:machineauth.json")
}

func connectJSON(databaseURL string) (Database, error) {
	filename := "machineauth.json"
	if strings.HasPrefix(databaseURL, "json:") {
		filename = strings.TrimPrefix(databaseURL, "json:")
		if filename == "json" {
			filename = "machineauth.json"
		}
	}

	jdb := &JSONDB{filename: filename}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			jdb.Agents = []Agent{}
			jdb.AuditLogs = []AuditLog{}
			jdb.WebhookConfigs = []WebhookConfig{}
			jdb.WebhookDeliveries = []WebhookDelivery{}
			return &DB{JSONDB: jdb}, nil
		}
		return nil, fmt.Errorf("failed to read database file: %w", err)
	}

	if err := json.Unmarshal(data, &jdb); err != nil {
		return nil, fmt.Errorf("failed to parse database file: %w", err)
	}

	return &DB{JSONDB: jdb}, nil
}

func (db *JSONDB) Save() error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	if err := os.WriteFile(db.filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write database file: %w", err)
	}

	return nil
}

func (db *DB) Close() error {
	return db.JSONDB.Save()
}

func (db *DB) CreateAgent(agent Agent) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Agents = append(db.Agents, agent)
	return db.Save()
}

func (db *DB) GetAgentByClientID(clientID string) (*Agent, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.Agents {
		if db.Agents[i].ClientID == clientID {
			return &db.Agents[i], nil
		}
	}
	return nil, fmt.Errorf("agent not found")
}

func (db *DB) GetAgentByID(id string) (*Agent, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.Agents {
		if db.Agents[i].ID == id {
			return &db.Agents[i], nil
		}
	}
	return nil, fmt.Errorf("agent not found")
}

func (db *DB) ListAgents() ([]Agent, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	agents := make([]Agent, len(db.Agents))
	copy(agents, db.Agents)
	return agents, nil
}

func (db *DB) ListAgentsByOrganization(orgID string) ([]Agent, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var agents []Agent
	for i := range db.Agents {
		if db.Agents[i].OrganizationID == orgID {
			agents = append(agents, db.Agents[i])
		}
	}
	return agents, nil
}

func (db *DB) ListAgentsByTeam(teamID string) ([]Agent, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var agents []Agent
	for i := range db.Agents {
		if db.Agents[i].TeamID != nil && *db.Agents[i].TeamID == teamID {
			agents = append(agents, db.Agents[i])
		}
	}
	return agents, nil
}

func (db *DB) UpdateAgent(id string, updateFn func(*Agent) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.Agents {
		if db.Agents[i].ID == id {
			if err := updateFn(&db.Agents[i]); err != nil {
				return err
			}
			db.Agents[i].UpdatedAt = time.Now()
			return db.Save()
		}
	}
	return fmt.Errorf("agent not found")
}

func (db *DB) DeleteAgent(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.Agents {
		if db.Agents[i].ID == id {
			db.Agents = append(db.Agents[:i], db.Agents[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("agent not found")
}

func (db *DB) AddAuditLog(log AuditLog) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.AuditLogs = append(db.AuditLogs, log)
	return db.Save()
}

// Upstream: RefreshToken, RevokedToken, Metrics, Organization, Team, APIKey methods

func (db *DB) CreateRefreshToken(rt RefreshToken) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.RefreshTokens = append(db.RefreshTokens, rt)
	return db.Save()
}

func (db *DB) GetRefreshToken(id string) (*RefreshToken, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.RefreshTokens {
		if db.RefreshTokens[i].ID == id {
			return &db.RefreshTokens[i], nil
		}
	}
	return nil, fmt.Errorf("refresh token not found")
}

func (db *DB) RevokeRefreshToken(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.RefreshTokens {
		if db.RefreshTokens[i].ID == id {
			now := time.Now()
			db.RefreshTokens[i].RevokedAt = &now
			return db.Save()
		}
	}
	return fmt.Errorf("refresh token not found")
}

func (db *DB) AddRevokedToken(rt RevokedToken) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.RevokedTokens = append(db.RevokedTokens, rt)
	return db.Save()
}

func (db *DB) IsTokenRevoked(jti string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, rt := range db.RevokedTokens {
		if rt.JTI == jti {
			return true
		}
	}
	return false
}

func (db *DB) IncrementTokensRefreshed() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Metrics.TokensRefreshed++
	return db.Save()
}

func (db *DB) IncrementTokensRevoked() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Metrics.TokensRevoked++
	return db.Save()
}

func (db *DB) GetMetrics() Metrics {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.Metrics
}

func (db *DB) CreateOrganization(org Organization) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Organizations = append(db.Organizations, org)
	return db.Save()
}

func (db *DB) GetOrganization(id string) (*Organization, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.Organizations {
		if db.Organizations[i].ID == id {
			return &db.Organizations[i], nil
		}
	}
	return nil, fmt.Errorf("organization not found")
}

func (db *DB) GetOrganizationBySlug(slug string) (*Organization, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.Organizations {
		if db.Organizations[i].Slug == slug {
			return &db.Organizations[i], nil
		}
	}
	return nil, fmt.Errorf("organization not found")
}

func (db *DB) ListOrganizations() ([]Organization, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	orgs := make([]Organization, len(db.Organizations))
	copy(orgs, db.Organizations)
	return orgs, nil
}

func (db *DB) UpdateOrganization(id string, updateFn func(*Organization) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.Organizations {
		if db.Organizations[i].ID == id {
			if err := updateFn(&db.Organizations[i]); err != nil {
				return err
			}
			db.Organizations[i].UpdatedAt = time.Now()
			return db.Save()
		}
	}
	return fmt.Errorf("organization not found")
}

func (db *DB) DeleteOrganization(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.Organizations {
		if db.Organizations[i].ID == id {
			db.Organizations = append(db.Organizations[:i], db.Organizations[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("organization not found")
}

func (db *DB) CreateTeam(team Team) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Teams = append(db.Teams, team)
	return db.Save()
}

func (db *DB) GetTeam(id string) (*Team, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.Teams {
		if db.Teams[i].ID == id {
			return &db.Teams[i], nil
		}
	}
	return nil, fmt.Errorf("team not found")
}

func (db *DB) ListTeamsByOrganization(orgID string) ([]Team, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var teams []Team
	for i := range db.Teams {
		if db.Teams[i].OrganizationID == orgID {
			teams = append(teams, db.Teams[i])
		}
	}
	return teams, nil
}

func (db *DB) UpdateTeam(id string, updateFn func(*Team) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.Teams {
		if db.Teams[i].ID == id {
			if err := updateFn(&db.Teams[i]); err != nil {
				return err
			}
			db.Teams[i].UpdatedAt = time.Now()
			return db.Save()
		}
	}
	return fmt.Errorf("team not found")
}

func (db *DB) DeleteTeam(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.Teams {
		if db.Teams[i].ID == id {
			db.Teams = append(db.Teams[:i], db.Teams[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("team not found")
}

func (db *DB) CreateAPIKey(key APIKey) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.APIKeys = append(db.APIKeys, key)
	return db.Save()
}

func (db *DB) GetAPIKeyByID(id string) (*APIKey, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.APIKeys {
		if db.APIKeys[i].ID == id {
			return &db.APIKeys[i], nil
		}
	}
	return nil, fmt.Errorf("API key not found")
}

func (db *DB) GetAPIKeyByKeyHash(keyHash string) (*APIKey, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.APIKeys {
		if db.APIKeys[i].KeyHash == keyHash && db.APIKeys[i].IsActive {
			return &db.APIKeys[i], nil
		}
	}
	return nil, fmt.Errorf("API key not found")
}

func (db *DB) ListAPIKeysByOrganization(orgID string) ([]APIKey, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var keys []APIKey
	for i := range db.APIKeys {
		if db.APIKeys[i].OrganizationID == orgID {
			keys = append(keys, db.APIKeys[i])
		}
	}
	return keys, nil
}

func (db *DB) UpdateAPIKey(id string, updateFn func(*APIKey) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.APIKeys {
		if db.APIKeys[i].ID == id {
			if err := updateFn(&db.APIKeys[i]); err != nil {
				return err
			}
			return db.Save()
		}
	}
	return fmt.Errorf("API key not found")
}

func (db *DB) DeleteAPIKey(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.APIKeys {
		if db.APIKeys[i].ID == id {
			db.APIKeys = append(db.APIKeys[:i], db.APIKeys[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("API key not found")
}

// Webhook CRUD methods

func (db *DB) CreateWebhook(webhook WebhookConfig) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.WebhookConfigs == nil {
		db.WebhookConfigs = []WebhookConfig{}
	}
	db.WebhookConfigs = append(db.WebhookConfigs, webhook)
	return db.Save()
}

func (db *DB) GetWebhook(id string) (*WebhookConfig, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.WebhookConfigs {
		if db.WebhookConfigs[i].ID == id {
			return &db.WebhookConfigs[i], nil
		}
	}
	return nil, fmt.Errorf("webhook not found")
}

func (db *DB) ListWebhooks() ([]WebhookConfig, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	webhooks := make([]WebhookConfig, len(db.WebhookConfigs))
	copy(webhooks, db.WebhookConfigs)
	return webhooks, nil
}

func (db *DB) UpdateWebhook(id string, updateFn func(*WebhookConfig) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.WebhookConfigs {
		if db.WebhookConfigs[i].ID == id {
			if err := updateFn(&db.WebhookConfigs[i]); err != nil {
				return err
			}
			db.WebhookConfigs[i].UpdatedAt = time.Now()
			return db.Save()
		}
	}
	return fmt.Errorf("webhook not found")
}

func (db *DB) DeleteWebhook(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.WebhookConfigs {
		if db.WebhookConfigs[i].ID == id {
			db.WebhookConfigs = append(db.WebhookConfigs[:i], db.WebhookConfigs[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("webhook not found")
}

func (db *DB) ListActiveWebhooksForEvent(event string) ([]WebhookConfig, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []WebhookConfig
	for _, wh := range db.WebhookConfigs {
		if !wh.IsActive {
			continue
		}
		for _, e := range wh.Events {
			if e == event || e == "*" {
				result = append(result, wh)
				break
			}
		}
	}
	return result, nil
}

// Webhook Delivery methods

func (db *DB) AddWebhookDelivery(delivery WebhookDelivery) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.WebhookDeliveries == nil {
		db.WebhookDeliveries = []WebhookDelivery{}
	}
	db.WebhookDeliveries = append(db.WebhookDeliveries, delivery)
	return db.Save()
}

func (db *DB) GetWebhookDelivery(id string) (*WebhookDelivery, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.WebhookDeliveries {
		if db.WebhookDeliveries[i].ID == id {
			return &db.WebhookDeliveries[i], nil
		}
	}
	return nil, fmt.Errorf("webhook delivery not found")
}

func (db *DB) UpdateWebhookDelivery(id string, updateFn func(*WebhookDelivery) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.WebhookDeliveries {
		if db.WebhookDeliveries[i].ID == id {
			if err := updateFn(&db.WebhookDeliveries[i]); err != nil {
				return err
			}
			return db.Save()
		}
	}
	return fmt.Errorf("webhook delivery not found")
}

func (db *DB) ListWebhookDeliveries(webhookConfigID string) ([]WebhookDelivery, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []WebhookDelivery
	for _, d := range db.WebhookDeliveries {
		if d.WebhookConfigID == webhookConfigID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (db *DB) ListPendingDeliveries() ([]WebhookDelivery, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	now := time.Now()
	var result []WebhookDelivery
	for _, d := range db.WebhookDeliveries {
		if d.Status == "pending" || (d.Status == "retrying" && d.NextRetryAt != nil && !d.NextRetryAt.After(now)) {
			result = append(result, d)
		}
	}
	return result, nil
}

// ── Admin User methods ───────────────────────────────────────────────

func (db *DB) CreateAdminUser(user AdminUser) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.AdminUsers = append(db.AdminUsers, user)
	return db.Save()
}

func (db *DB) GetAdminUserByEmail(email string) (*AdminUser, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for i := range db.AdminUsers {
		if db.AdminUsers[i].Email == email {
			return &db.AdminUsers[i], nil
		}
	}
	return nil, fmt.Errorf("admin user not found")
}

func (db *DB) GetAdminUserByID(id string) (*AdminUser, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for i := range db.AdminUsers {
		if db.AdminUsers[i].ID == id {
			return &db.AdminUsers[i], nil
		}
	}
	return nil, fmt.Errorf("admin user not found")
}

func (db *DB) ListAdminUsers() ([]AdminUser, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	out := make([]AdminUser, len(db.AdminUsers))
	copy(out, db.AdminUsers)
	return out, nil
}

// ── Audit log query methods ──────────────────────────────────────────

func (db *DB) ListAuditLogs(agentID, action, ipAddress string, from, to *time.Time, page, limit int) ([]AuditLog, int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var filtered []AuditLog
	for _, l := range db.AuditLogs {
		if agentID != "" && (l.AgentID == nil || *l.AgentID != agentID) {
			continue
		}
		if action != "" && l.Action != action {
			continue
		}
		if ipAddress != "" && l.IPAddress != ipAddress {
			continue
		}
		if from != nil && l.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && l.CreatedAt.After(*to) {
			continue
		}
		filtered = append(filtered, l)
	}

	total := len(filtered)

	// Sort descending by created_at (most recent first).
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].CreatedAt.After(filtered[i].CreatedAt) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Paginate.
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	start := (page - 1) * limit
	if start >= len(filtered) {
		return []AuditLog{}, total, nil
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

// ── Agent count (for pagination) ─────────────────────────────────────

func (db *DB) CountAgents() (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.Agents), nil
}

func (db *DB) ListAgentsPaginated(search, status, orgID, sort string, page, limit int) ([]Agent, int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var filtered []Agent
	for _, a := range db.Agents {
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(a.Name), searchLower) &&
				!strings.Contains(strings.ToLower(a.ClientID), searchLower) &&
				!strings.Contains(strings.ToLower(a.Description), searchLower) {
				continue
			}
		}
		if status != "" && a.Status != status {
			continue
		}
		if orgID != "" && a.OrganizationID != orgID {
			continue
		}
		filtered = append(filtered, a)
	}

	total := len(filtered)

	// Sort by created_at descending by default.
	if sort == "name" {
		for i := 0; i < len(filtered); i++ {
			for j := i + 1; j < len(filtered); j++ {
				if strings.ToLower(filtered[i].Name) > strings.ToLower(filtered[j].Name) {
					filtered[i], filtered[j] = filtered[j], filtered[i]
				}
			}
		}
	} else {
		// Default: newest first.
		for i := 0; i < len(filtered); i++ {
			for j := i + 1; j < len(filtered); j++ {
				if filtered[j].CreatedAt.After(filtered[i].CreatedAt) {
					filtered[i], filtered[j] = filtered[j], filtered[i]
				}
			}
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	start := (page - 1) * limit
	if start >= len(filtered) {
		return []Agent{}, total, nil
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

// RunMigrations is now in migrate.go (applies SQL migrations for Postgres, no-op for JSON).

// ── Org-scoped webhook methods ───────────────────────────────────────

func (db *DB) ListWebhooksByOrganization(orgID string) ([]WebhookConfig, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []WebhookConfig
	for _, wh := range db.WebhookConfigs {
		if wh.OrganizationID == orgID {
			result = append(result, wh)
		}
	}
	return result, nil
}

func (db *DB) ListActiveWebhooksForEventByOrg(event, orgID string) ([]WebhookConfig, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []WebhookConfig
	for _, wh := range db.WebhookConfigs {
		if !wh.IsActive {
			continue
		}
		if orgID != "" && wh.OrganizationID != orgID {
			continue
		}
		for _, e := range wh.Events {
			if e == event || e == "*" {
				result = append(result, wh)
				break
			}
		}
	}
	return result, nil
}

// ── OrgMember methods ────────────────────────────────────────────────

func (db *DB) CreateOrgMember(member OrgMember) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.OrgMembers = append(db.OrgMembers, member)
	return db.Save()
}

func (db *DB) GetOrgMember(userID, orgID string) (*OrgMember, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.OrgMembers {
		if db.OrgMembers[i].UserID == userID && db.OrgMembers[i].OrganizationID == orgID {
			return &db.OrgMembers[i], nil
		}
	}
	return nil, fmt.Errorf("org member not found")
}

func (db *DB) ListOrgMembersByOrg(orgID string) ([]OrgMember, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []OrgMember
	for _, m := range db.OrgMembers {
		if m.OrganizationID == orgID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (db *DB) ListOrgMembersByUser(userID string) ([]OrgMember, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []OrgMember
	for _, m := range db.OrgMembers {
		if m.UserID == userID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (db *DB) DeleteOrgMember(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.OrgMembers {
		if db.OrgMembers[i].ID == id {
			db.OrgMembers = append(db.OrgMembers[:i], db.OrgMembers[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("org member not found")
}

func (db *DB) UpdateOrgMember(id string, updateFn func(*OrgMember) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.OrgMembers {
		if db.OrgMembers[i].ID == id {
			if err := updateFn(&db.OrgMembers[i]); err != nil {
				return err
			}
			db.OrgMembers[i].UpdatedAt = time.Now()
			return db.Save()
		}
	}
	return fmt.Errorf("org member not found")
}

// ── OrgSigningKey methods ────────────────────────────────────────────

func (db *DB) CreateOrgSigningKey(key OrgSigningKey) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.OrgSigningKeys = append(db.OrgSigningKeys, key)
	return db.Save()
}

func (db *DB) GetActiveOrgSigningKey(orgID string) (*OrgSigningKey, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i := range db.OrgSigningKeys {
		k := &db.OrgSigningKeys[i]
		if k.OrganizationID == orgID && k.IsActive {
			if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
				continue
			}
			return k, nil
		}
	}
	return nil, fmt.Errorf("no active signing key for org %s", orgID)
}

func (db *DB) ListOrgSigningKeys(orgID string) ([]OrgSigningKey, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []OrgSigningKey
	for _, k := range db.OrgSigningKeys {
		if k.OrganizationID == orgID {
			result = append(result, k)
		}
	}
	return result, nil
}

func (db *DB) DeleteOrgSigningKey(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.OrgSigningKeys {
		if db.OrgSigningKeys[i].ID == id {
			db.OrgSigningKeys = append(db.OrgSigningKeys[:i], db.OrgSigningKeys[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("signing key not found")
}
