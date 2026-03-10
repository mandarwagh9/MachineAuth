package models

import (
	"time"

	"github.com/google/uuid"
)

// AgentStatus represents the lifecycle state of an agent.
type AgentStatus string

const (
	AgentStatusActive    AgentStatus = "active"
	AgentStatusInactive  AgentStatus = "inactive"
	AgentStatusSuspended AgentStatus = "suspended"
	AgentStatusPending   AgentStatus = "pending"
	AgentStatusExpired   AgentStatus = "expired"
)

type Agent struct {
	ID                uuid.UUID              `json:"id"`
	OrganizationID    string                 `json:"organization_id"`
	TeamID            *uuid.UUID             `json:"team_id,omitempty"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ClientID          string                 `json:"client_id"`
	ClientSecretHash  string                 `json:"-"`
	Scopes            []string               `json:"scopes,omitempty"`
	PublicKey         *string                `json:"public_key,omitempty"`
	Status            AgentStatus            `json:"status"`
	IsActive          bool                   `json:"is_active"`
	RedirectURIs      []string               `json:"redirect_uris,omitempty"`
	GrantTypes        []string               `json:"grant_types,omitempty"`
	ClientType        string                 `json:"client_type,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
	TokenCount        int                    `json:"token_count"`
	RefreshCount      int                    `json:"refresh_count"`
	LastActivityAt    *time.Time             `json:"last_activity_at,omitempty"`
	LastTokenIssuedAt *time.Time             `json:"last_token_issued_at,omitempty"`
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
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	OrganizationID string                 `json:"organization_id"`
	TeamID         *uuid.UUID             `json:"team_id,omitempty"`
	Scopes         []string               `json:"scopes,omitempty"`
	ExpiresIn      *int                   `json:"expires_in,omitempty"`
	RedirectURIs   []string               `json:"redirect_uris,omitempty"`
	GrantTypes     []string               `json:"grant_types,omitempty"`
	ClientType     string                 `json:"client_type,omitempty"`
}

type UpdateAgentRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Scopes      []string               `json:"scopes,omitempty"`
	Status      *AgentStatus           `json:"status,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
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
	Agents     []Agent     `json:"agents"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination metadata for list endpoints.
type Pagination struct {
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// PaginationParams are query-string parameters accepted on list endpoints.
type PaginationParams struct {
	Page   int
	Limit  int
	Search string
	Status string
	OrgID  string
	Sort   string
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
	IssuedAt     int64  `json:"issued_at"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// OIDCDiscovery represents the OpenID Connect discovery document.
type OIDCDiscovery struct {
	Issuer                            string   `json:"issuer"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSUri                           string   `json:"jwks_uri"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported,omitempty"`
	ServiceDocumentation              string   `json:"service_documentation,omitempty"`
}

// UserInfoResponse is the profile returned by GET /oauth/userinfo.
type UserInfoResponse struct {
	Sub       string                 `json:"sub"`
	Name      string                 `json:"name,omitempty"`
	AgentID   string                 `json:"agent_id"`
	OrgID     string                 `json:"org_id,omitempty"`
	TeamID    string                 `json:"team_id,omitempty"`
	Scopes    []string               `json:"scopes,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt int64                  `json:"created_at,omitempty"`
	UpdatedAt int64                  `json:"updated_at,omitempty"`
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
	Details   string     `json:"details,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type AuditLogsListResponse struct {
	Logs       []AuditLog  `json:"logs"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type AuditLogQuery struct {
	AgentID   string
	Action    string
	IPAddress string
	From      *time.Time
	To        *time.Time
	Page      int
	Limit     int
}

// AdminUser represents an admin user that can access the management API.
type AdminUser struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OrgMember represents a user's membership in an organization with a role.
type OrgMember struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"` // owner, admin, member, viewer
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OrgMemberRole constants.
const (
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleMember = "member"
	OrgRoleViewer = "viewer"
)

// CreateOrgMemberRequest is the payload for adding a user to an org.
type CreateOrgMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// OrgMembersResponse is the list response for org members.
type OrgMembersResponse struct {
	Members []OrgMemberWithUser `json:"members"`
}

// OrgMemberWithUser extends OrgMember with the user's email.
type OrgMemberWithUser struct {
	OrgMember
	Email string `json:"email"`
}

// SignupRequest is the self-service registration payload.
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgName  string `json:"org_name"`
	OrgSlug  string `json:"org_slug"`
}

// OrgSigningKey represents a per-org RSA signing key stored in the database.
type OrgSigningKey struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID string     `json:"organization_id"`
	KeyID          string     `json:"key_id"`     // JWT kid
	PublicKeyPEM   string     `json:"public_key"` // PEM-encoded public key
	PrivateKeyPEM  string     `json:"-"`          // PEM-encoded private key (never serialized)
	Algorithm      string     `json:"algorithm"`  // RS256
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// AdminLoginRequest is the payload for POST /api/auth/login.
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// AdminTokenResponse is returned on successful admin login.
type AdminTokenResponse struct {
	Success     bool   `json:"success"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Role        string `json:"role"`
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

// Organization / Team / APIKey models

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

// AuthorizationCode represents an OAuth 2.0 authorization code.
type AuthorizationCode struct {
	ID                  uuid.UUID `json:"id"`
	ClientID            string    `json:"client_id"`
	UserID              string    `json:"user_id"`
	OrganizationID      string    `json:"organization_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	CodeChallenge       string    `json:"code_challenge,omitempty"`
	CodeChallengeMethod string    `json:"code_challenge_method,omitempty"`
	Code                string    `json:"-"`
	ExpiresAt           time.Time `json:"expires_at"`
	Used                bool      `json:"used"`
	CreatedAt           time.Time `json:"created_at"`
}

// AuthorizeRequest represents the OAuth 2.0 authorization request parameters.
type AuthorizeRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

// AuthorizeResponse represents the OAuth 2.0 authorization response.
type AuthorizeResponse struct {
	Code  string `json:"code"`
	State string `json:"state,omitempty"`
}

// AuthorizeErrorResponse represents an OAuth 2.0 authorization error.
type AuthorizeErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description,omitempty"`
	State       string `json:"state,omitempty"`
}
