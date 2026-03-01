# Go SDK Implementation

> **Note**: The Go SDK code below is ready to be placed in `sdk/go/` once the directory is created:
> ```bash
> mkdir -p sdk/go
> ```
> Then create each file listed below.

## `sdk/go/go.mod`

```go
module github.com/mandarwagh9/machineauth/sdk/go

go 1.21
```

## `sdk/go/client.go`

```go
// Package machineauth provides a Go SDK for the MachineAuth API.
package machineauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL.
func WithBaseURL(u string) ClientOption {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithCredentials sets the client credentials.
func WithCredentials(clientID, clientSecret string) ClientOption {
	return func(c *Client) { c.clientID = clientID; c.clientSecret = clientSecret }
}

// WithToken sets the default bearer token.
func WithToken(token string) ClientOption {
	return func(c *Client) { c.token = token }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.http = hc }
}

// WithTimeout sets the request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.http.Timeout = d }
}

// Client is the MachineAuth SDK client.
type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	token        string
	http         *http.Client
}

// New creates a Client with options.
func New(opts ...ClientOption) *Client {
	c := &Client{
		baseURL: "http://localhost:8081",
		http:    &http.Client{Timeout: 10 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// NewFromEnv creates a Client from MACHINEAUTH_* environment variables.
func NewFromEnv(opts ...ClientOption) *Client {
	c := New(opts...)
	if v := os.Getenv("MACHINEAUTH_BASE_URL"); v != "" {
		c.baseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("MACHINEAUTH_CLIENT_ID"); v != "" {
		c.clientID = v
	}
	if v := os.Getenv("MACHINEAUTH_CLIENT_SECRET"); v != "" {
		c.clientSecret = v
	}
	if v := os.Getenv("MACHINEAUTH_ACCESS_TOKEN"); v != "" {
		c.token = v
	}
	return c
}

// ── Error type ──────────────────────────────────────────────────

// APIError represents an error from the MachineAuth API.
type APIError struct {
	StatusCode int
	Message    string
	Body       json.RawMessage
}

func (e *APIError) Error() string {
	return fmt.Sprintf("machineauth: HTTP %d: %s", e.StatusCode, e.Message)
}

// ── Token types ─────────────────────────────────────────────────

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type IntrospectionResponse struct {
	Active   bool   `json:"active"`
	Scope    string `json:"scope,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Exp      int64  `json:"exp,omitempty"`
	Iat      int64  `json:"iat,omitempty"`
}

// ── Agent types ─────────────────────────────────────────────────

type Agent struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ClientID       string   `json:"client_id"`
	Scopes         []string `json:"scopes"`
	OrganizationID string   `json:"organization_id,omitempty"`
	TeamID         string   `json:"team_id,omitempty"`
	IsActive       bool     `json:"is_active"`
	CreatedAt      string   `json:"created_at,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
	TokenCount     int      `json:"token_count,omitempty"`
	RefreshCount   int      `json:"refresh_count,omitempty"`
}

type CreateAgentRequest struct {
	Name           string   `json:"name"`
	Scopes         []string `json:"scopes,omitempty"`
	OrganizationID string   `json:"organization_id,omitempty"`
	TeamID         string   `json:"team_id,omitempty"`
	ExpiresIn      *int     `json:"expires_in,omitempty"`
}

type AgentUsage struct {
	Agent           Agent                    `json:"agent"`
	OrganizationID  string                   `json:"organization_id,omitempty"`
	TokenCount      int                      `json:"token_count"`
	RefreshCount    int                      `json:"refresh_count"`
	RotationHistory []map[string]interface{} `json:"rotation_history,omitempty"`
}

// ── Organization types ──────────────────────────────────────────

type Organization struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	OwnerEmail     string   `json:"owner_email,omitempty"`
	JWTIssuer      string   `json:"jwt_issuer,omitempty"`
	JWTExpirySecs  *int     `json:"jwt_expiry_secs,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
}

type Team struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
	Description    string `json:"description,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}

// ── API Key types ───────────────────────────────────────────────

type APIKey struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
	KeyPrefix      string `json:"key_prefix,omitempty"`
	TeamID         string `json:"team_id,omitempty"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
}

// ── Webhook types ───────────────────────────────────────────────

type WebhookConfig struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	URL              string   `json:"url"`
	Events           []string `json:"events"`
	Secret           string   `json:"secret,omitempty"`
	IsActive         bool     `json:"is_active"`
	MaxRetries       int      `json:"max_retries,omitempty"`
	RetryBackoffBase int      `json:"retry_backoff_base,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
	ConsecutiveFails int      `json:"consecutive_fails,omitempty"`
}

type WebhookDelivery struct {
	ID              string  `json:"id"`
	WebhookConfigID string  `json:"webhook_config_id"`
	Event           string  `json:"event"`
	Status          string  `json:"status"`
	Attempts        int     `json:"attempts,omitempty"`
	LastError       string  `json:"last_error,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	NextRetryAt     *string `json:"next_retry_at,omitempty"`
}

// ── Token operations ────────────────────────────────────────────

func (c *Client) ClientCredentialsToken(ctx context.Context) (*TokenResponse, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return nil, fmt.Errorf("machineauth: client_id and client_secret are required")
	}
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
	}
	var tok TokenResponse
	if err := c.doForm(ctx, "POST", "/oauth/token", form, &tok); err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	c.token = tok.AccessToken
	return &tok, nil
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	if c.clientID != "" {
		form.Set("client_id", c.clientID)
	}
	if c.clientSecret != "" {
		form.Set("client_secret", c.clientSecret)
	}
	var tok TokenResponse
	if err := c.doForm(ctx, "POST", "/oauth/token", form, &tok); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	c.token = tok.AccessToken
	return &tok, nil
}

func (c *Client) Introspect(ctx context.Context, token string) (*IntrospectionResponse, error) {
	var resp IntrospectionResponse
	err := c.doForm(ctx, "POST", "/oauth/introspect", url.Values{"token": {token}}, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect: %w", err)
	}
	return &resp, nil
}

func (c *Client) Revoke(ctx context.Context, token string) error {
	if err := c.doForm(ctx, "POST", "/oauth/revoke", url.Values{"token": {token}}, nil); err != nil {
		return fmt.Errorf("failed to revoke: %w", err)
	}
	return nil
}

// ── Agent management ────────────────────────────────────────────

func (c *Client) ListAgents(ctx context.Context) ([]Agent, error) {
	var resp struct{ Agents []Agent `json:"agents"` }
	if err := c.doJSON(ctx, "GET", "/api/agents", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	return resp.Agents, nil
}

func (c *Client) CreateAgent(ctx context.Context, req *CreateAgentRequest) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.doJSON(ctx, "POST", "/api/agents", req, &raw); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	return raw, nil
}

func (c *Client) GetAgent(ctx context.Context, id string) (*Agent, error) {
	var resp struct{ Agent Agent `json:"agent"` }
	if err := c.doJSON(ctx, "GET", "/api/agents/"+id, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return &resp.Agent, nil
}

func (c *Client) DeleteAgent(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, "DELETE", "/api/agents/"+id, nil, nil); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	return nil
}

func (c *Client) RotateAgent(ctx context.Context, id string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.doJSON(ctx, "POST", "/api/agents/"+id+"/rotate", nil, &raw); err != nil {
		return nil, fmt.Errorf("failed to rotate agent: %w", err)
	}
	return raw, nil
}

// ── Self-service ────────────────────────────────────────────────

func (c *Client) GetMe(ctx context.Context) (*Agent, error) {
	var resp struct{ Agent Agent `json:"agent"` }
	if err := c.doJSON(ctx, "GET", "/api/agents/me", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get self: %w", err)
	}
	return &resp.Agent, nil
}

func (c *Client) GetMyUsage(ctx context.Context) (*AgentUsage, error) {
	var resp AgentUsage
	if err := c.doJSON(ctx, "GET", "/api/agents/me/usage", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}
	return &resp, nil
}

func (c *Client) RotateMe(ctx context.Context) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.doJSON(ctx, "POST", "/api/agents/me/rotate", nil, &raw); err != nil {
		return nil, fmt.Errorf("failed to rotate self: %w", err)
	}
	return raw, nil
}

func (c *Client) DeactivateMe(ctx context.Context) error {
	return c.doJSON(ctx, "POST", "/api/agents/me/deactivate", nil, nil)
}

func (c *Client) ReactivateMe(ctx context.Context) error {
	return c.doJSON(ctx, "POST", "/api/agents/me/reactivate", nil, nil)
}

func (c *Client) DeleteMe(ctx context.Context) error {
	return c.doJSON(ctx, "DELETE", "/api/agents/me", nil, nil)
}

// ── Organizations ───────────────────────────────────────────────

func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	var resp struct{ Organizations []Organization `json:"organizations"` }
	if err := c.doJSON(ctx, "GET", "/api/organizations", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list orgs: %w", err)
	}
	return resp.Organizations, nil
}

func (c *Client) CreateOrganization(ctx context.Context, name, slug, ownerEmail string) (*Organization, error) {
	var org Organization
	body := map[string]string{"name": name, "slug": slug, "owner_email": ownerEmail}
	if err := c.doJSON(ctx, "POST", "/api/organizations", body, &org); err != nil {
		return nil, fmt.Errorf("failed to create org: %w", err)
	}
	return &org, nil
}

func (c *Client) GetOrganization(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+id, nil, &org); err != nil {
		return nil, fmt.Errorf("failed to get org: %w", err)
	}
	return &org, nil
}

func (c *Client) UpdateOrganization(ctx context.Context, id string, fields map[string]interface{}) (*Organization, error) {
	var org Organization
	if err := c.doJSON(ctx, "PUT", "/api/organizations/"+id, fields, &org); err != nil {
		return nil, fmt.Errorf("failed to update org: %w", err)
	}
	return &org, nil
}

func (c *Client) DeleteOrganization(ctx context.Context, id string) error {
	return c.doJSON(ctx, "DELETE", "/api/organizations/"+id, nil, nil)
}

// ── Teams ───────────────────────────────────────────────────────

func (c *Client) ListTeams(ctx context.Context, orgID string) ([]Team, error) {
	var resp struct{ Teams []Team `json:"teams"` }
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+orgID+"/teams", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	return resp.Teams, nil
}

func (c *Client) CreateTeam(ctx context.Context, orgID, name string, description string) (*Team, error) {
	var team Team
	body := map[string]string{"name": name}
	if description != "" {
		body["description"] = description
	}
	if err := c.doJSON(ctx, "POST", "/api/organizations/"+orgID+"/teams", body, &team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}
	return &team, nil
}

// ── API Keys ────────────────────────────────────────────────────

func (c *Client) ListAPIKeys(ctx context.Context, orgID string) ([]APIKey, error) {
	var resp struct{ APIKeys []APIKey `json:"api_keys"` }
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+orgID+"/api-keys", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	return resp.APIKeys, nil
}

func (c *Client) CreateAPIKey(ctx context.Context, orgID, name string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.doJSON(ctx, "POST", "/api/organizations/"+orgID+"/api-keys", map[string]string{"name": name}, &raw); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}
	return raw, nil
}

func (c *Client) DeleteAPIKey(ctx context.Context, orgID, keyID string) error {
	return c.doJSON(ctx, "DELETE", "/api/organizations/"+orgID+"/api-keys/"+keyID, nil, nil)
}

// ── Webhooks ────────────────────────────────────────────────────

func (c *Client) ListWebhooks(ctx context.Context) ([]WebhookConfig, error) {
	var resp struct{ Webhooks []WebhookConfig `json:"webhooks"` }
	if err := c.doJSON(ctx, "GET", "/api/webhooks", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	return resp.Webhooks, nil
}

func (c *Client) CreateWebhook(ctx context.Context, name, webhookURL string, events []string) (json.RawMessage, error) {
	var raw json.RawMessage
	body := map[string]interface{}{"name": name, "url": webhookURL, "events": events}
	if err := c.doJSON(ctx, "POST", "/api/webhooks", body, &raw); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return raw, nil
}

func (c *Client) GetWebhook(ctx context.Context, id string) (*WebhookConfig, error) {
	var resp struct{ Webhook WebhookConfig `json:"webhook"` }
	if err := c.doJSON(ctx, "GET", "/api/webhooks/"+id, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return &resp.Webhook, nil
}

func (c *Client) UpdateWebhook(ctx context.Context, id string, fields map[string]interface{}) (*WebhookConfig, error) {
	var resp struct{ Webhook WebhookConfig `json:"webhook"` }
	if err := c.doJSON(ctx, "PUT", "/api/webhooks/"+id, fields, &resp); err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}
	return &resp.Webhook, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	return c.doJSON(ctx, "DELETE", "/api/webhooks/"+id, nil, nil)
}

func (c *Client) TestWebhook(ctx context.Context, id, event string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.doJSON(ctx, "POST", "/api/webhooks/"+id+"/test", map[string]string{"event": event}, &raw); err != nil {
		return nil, fmt.Errorf("failed to test webhook: %w", err)
	}
	return raw, nil
}

func (c *Client) ListDeliveries(ctx context.Context, webhookID string) ([]WebhookDelivery, error) {
	var resp struct{ Deliveries []WebhookDelivery `json:"deliveries"` }
	if err := c.doJSON(ctx, "GET", "/api/webhooks/"+webhookID+"/deliveries", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}
	return resp.Deliveries, nil
}

func (c *Client) ListWebhookEvents(ctx context.Context) ([]string, error) {
	var resp struct{ Events []string `json:"events"` }
	if err := c.doJSON(ctx, "GET", "/api/webhook-events", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	return resp.Events, nil
}

// ── HTTP helpers ────────────────────────────────────────────────

func (c *Client) doForm(ctx context.Context, method, path string, form url.Values, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.do(req, out)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: data}
		var parsed map[string]interface{}
		if json.Unmarshal(data, &parsed) == nil {
			if msg, ok := parsed["message"].(string); ok {
				apiErr.Message = msg
			} else if msg, ok := parsed["error"].(string); ok {
				apiErr.Message = msg
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = resp.Status
		}
		return apiErr
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
```

## Setup Instructions

```bash
mkdir -p sdk/go
# Copy the above client.go content into sdk/go/client.go
# Copy the go.mod content into sdk/go/go.mod
cd sdk/go && go build ./...
```
