package models

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	ClientID         string     `json:"client_id"`
	ClientSecretHash string     `json:"-"`
	Scopes           []string   `json:"scopes,omitempty"`
	PublicKey        *string    `json:"public_key,omitempty"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
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

// Webhook models

type WebhookConfig struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   string     `json:"organization_id,omitempty"`
	TeamID           string     `json:"team_id,omitempty"`
	Name             string     `json:"name"`
	URL              string     `json:"url"`
	Secret           string     `json:"-"`
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
	ID              uuid.UUID  `json:"id"`
	WebhookConfigID uuid.UUID  `json:"webhook_config_id"`
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

// Webhook request/response types

type CreateWebhookRequest struct {
	Name             string   `json:"name"`
	URL              string   `json:"url"`
	Events           []string `json:"events"`
	MaxRetries       *int     `json:"max_retries,omitempty"`
	RetryBackoffBase *int     `json:"retry_backoff_base,omitempty"`
}

type UpdateWebhookRequest struct {
	Name             *string  `json:"name,omitempty"`
	URL              *string  `json:"url,omitempty"`
	Events           []string `json:"events,omitempty"`
	IsActive         *bool    `json:"is_active,omitempty"`
	MaxRetries       *int     `json:"max_retries,omitempty"`
	RetryBackoffBase *int     `json:"retry_backoff_base,omitempty"`
}

type WebhookResponse struct {
	Webhook WebhookConfig `json:"webhook"`
}

type CreateWebhookResponse struct {
	Webhook WebhookConfig `json:"webhook"`
	Secret  string        `json:"secret"`
}

type WebhooksListResponse struct {
	Webhooks []WebhookConfig `json:"webhooks"`
}

type WebhookDeliveryResponse struct {
	Delivery WebhookDelivery `json:"delivery"`
}

type WebhookDeliveriesListResponse struct {
	Deliveries []WebhookDelivery `json:"deliveries"`
}

type TestWebhookRequest struct {
	Event   string `json:"event"`
	Payload string `json:"payload,omitempty"`
}

type TestWebhookResponse struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
}
