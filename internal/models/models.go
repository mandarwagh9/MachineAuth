package models

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	ClientID          string     `json:"client_id"`
	ClientSecretHash  string     `json:"-"`
	Scopes            []string   `json:"scopes,omitempty"`
	PublicKey         *string    `json:"public_key,omitempty"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	TokenCount        int        `json:"token_count"`
	RefreshCount      int        `json:"refresh_count"`
	LastActivityAt    *time.Time `json:"last_activity_at,omitempty"`
	LastTokenIssuedAt *time.Time `json:"last_token_issued_at,omitempty"`
}

type Rotation struct {
	RotatedAt   time.Time `json:"rotated_at"`
	RotatedByIP string    `json:"rotated_by_ip,omitempty"`
}

type AgentUsage struct {
	Agent             Agent      `json:"agent"`
	TokenCount        int        `json:"token_count"`
	RefreshCount      int        `json:"refresh_count"`
	LastActivityAt    *time.Time `json:"last_activity_at,omitempty"`
	LastTokenIssuedAt *time.Time `json:"last_token_issued_at,omitempty"`
	RotationHistory   []Rotation `json:"rotation_history"`
}

type CreateAgentRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes,omitempty"`
	ExpiresIn *int     `json:"expires_in,omitempty"`
}

type CreateAgentResponse struct {
	Agent        Agent  `json:"agent"`
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
}

type AgentResponse struct {
	Agent Agent `json:"agent"`
}

type AgentsListResponse struct {
	Agents []Agent `json:"agents"`
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
	IssuedAt    int64  `json:"issued_at"`
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type AuditLog struct {
	ID        uuid.UUID  `json:"id"`
	AgentID   *uuid.UUID `json:"agent_id,omitempty"`
	Action    string     `json:"action"`
	IPAddress string     `json:"ip_address,omitempty"`
	UserAgent string     `json:"user_agent,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}
