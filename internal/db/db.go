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
	mu        sync.RWMutex
	filename  string
	Agents    []Agent    `json:"agents"`
	AuditLogs []AuditLog `json:"audit_logs"`
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

func RunMigrations(db *DB) error {
	return nil
}
