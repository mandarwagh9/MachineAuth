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
	mu            sync.RWMutex
	filename      string
	Agents        []Agent        `json:"agents"`
	AuditLogs     []AuditLog     `json:"audit_logs"`
	RefreshTokens []RefreshToken `json:"refresh_tokens"`
	RevokedTokens []RevokedToken `json:"revoked_tokens"`
	Metrics       Metrics        `json:"metrics"`
	Organizations []Organization `json:"organizations"`
	Teams         []Team         `json:"teams"`
	APIKeys       []APIKey       `json:"api_keys"`
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
	ID                string     `json:"id"`
	OrganizationID    string     `json:"organization_id"`
	TeamID            *string    `json:"team_id,omitempty"`
	Name              string     `json:"name"`
	ClientID          string     `json:"client_id"`
	ClientSecretHash  string     `json:"client_secret_hash"`
	Scopes            []string   `json:"scopes"`
	PublicKey         *string    `json:"public_key,omitempty"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	TokenCount        int        `json:"token_count"`
	RefreshCount      int        `json:"refresh_count"`
	LastActivityAt    *time.Time `json:"last_activity_at,omitempty"`
	LastTokenIssuedAt *time.Time `json:"last_token_issued_at,omitempty"`
	RotationHistory   []Rotation `json:"rotation_history"`
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
	CreatedAt time.Time `json:"created_at"`
}

type DB struct {
	*JSONDB
}

func Connect(databaseURL string) (*DB, error) {
	if strings.HasPrefix(databaseURL, "json:") || strings.HasSuffix(databaseURL, ".json") {
		return connectJSON(databaseURL)
	}
	// Default to JSON
	return connectJSON("json:machineauth.json")
}

func connectJSON(databaseURL string) (*DB, error) {
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

func RunMigrations(db *DB) error {
	return nil
}
