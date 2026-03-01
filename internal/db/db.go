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
	mu               sync.RWMutex
	filename         string
	Agents           []Agent           `json:"agents"`
	AuditLogs        []AuditLog        `json:"audit_logs"`
	WebhookConfigs   []WebhookConfig   `json:"webhook_configs"`
	WebhookDeliveries []WebhookDelivery `json:"webhook_deliveries"`
}

type Agent struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	ClientID         string     `json:"client_id"`
	ClientSecretHash string     `json:"client_secret_hash"`
	Scopes           []string   `json:"scopes"`
	PublicKey        *string    `json:"public_key,omitempty"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	AgentID   *string   `json:"agent_id,omitempty"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
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

func Connect(databaseURL string) (*DB, error) {
	if strings.HasPrefix(databaseURL, "json:") || strings.HasSuffix(databaseURL, ".json") {
		return connectJSON(databaseURL)
	}
	// Default to JSON
	return connectJSON("json:agentauth.json")
}

func connectJSON(databaseURL string) (*DB, error) {
	filename := "agentauth.json"
	if strings.HasPrefix(databaseURL, "json:") {
		filename = strings.TrimPrefix(databaseURL, "json:")
		if filename == "json" {
			filename = "agentauth.json"
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

func RunMigrations(db *DB) error {
	return nil
}
