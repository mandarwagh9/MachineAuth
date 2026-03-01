package models

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID                uuid.UUID  `json:"id"`
	OrganizationID    string     `json:"organization_id"`
	TeamID            *uuid.UUID `json:"team_id,omitempty"`
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
	OrganizationID    string     `json:"organization_id"`
	TeamID            *uuid.UUID `json:"team_id,omitempty"`
	TokenCount        int        `json:"token_count"`
	RefreshCount      int        `json:"refresh_count"`
	LastActivityAt    *time.Time `json:"last_activity_at,omitempty"`
	LastTokenIssuedAt *time.Time `json:"last_token_issued_at,omitempty"`
	RotationHistory   []Rotation `json:"rotation_history"`
}

type CreateAgentRequest struct {
	Name           string     `json:"name"`
	OrganizationID string     `json:"organization_id"`
	TeamID         *uuid.UUID `json:"team_id,omitempty"`
	Scopes         []string   `json:"scopes,omitempty"`
	ExpiresIn      *int       `json:"expires_in,omitempty"`
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
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
	IssuedAt     int64  `json:"issued_at"`
	RefreshToken string `json:"refresh_token,omitempty"`
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

type CreateOrganizationRequest struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	OwnerEmail string `json:"owner_email"`
}

type UpdateOrganizationRequest struct {
	Name           string `json:"name"`
	JWTIssuer      string `json:"jwt_issuer"`
	JWTExpirySecs  int    `json:"jwt_expiry_secs"`
	AllowedOrigins string `json:"allowed_origins"`
}

type Team struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}

type TeamsResponse struct {
	Teams []Team `json:"teams"`
}

type APIKey struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	TeamID         *uuid.UUID `json:"team_id,omitempty"`
	Name           string     `json:"name"`
	Prefix         string     `json:"prefix"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	TeamID    *uuid.UUID `json:"team_id,omitempty"`
	ExpiresIn *int       `json:"expires_in,omitempty"`
}

type CreateAPIKeyResponse struct {
	APIKey APIKey `json:"api_key"`
	Key    string `json:"key"`
}

type APIKeysResponse struct {
	APIKeys []APIKey `json:"api_keys"`
}

type IntrospectResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	Revoked   bool   `json:"revoked,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type RevokeRequest struct {
	Token         string `json:"token"`
	TokenTypeHint string `json:"token_type_hint"`
}

type RefreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
