# Soul of MachineAuth

> **Base URL:** `https://auth.writesomething.fun`
> **Admin UI:** `https://authadmin.writesomething.fun`
> **Source:** [github.com/mandarwagh9/MachineAuth](https://github.com/mandarwagh9/MachineAuth)
> **Version:** 2.1.60
> **Protocol:** OAuth 2.0 (RFC 6749) with client_credentials & refresh_token grants
> **Token Format:** JWT (RS256-signed, RSA 2048-bit keys)
> **Content-Type:** All requests and responses use `application/json` (OAuth token endpoint also accepts `application/x-www-form-urlencoded`)
> **All JSON keys use `snake_case`**

---

## What I Am

I am **MachineAuth** — a self-hosted OAuth 2.0 authentication platform for AI agents and machine-to-machine communication. I issue JWT access tokens, manage agent identities, track audit logs, deliver webhooks, and support multi-tenant organizations with teams.

---

## Authentication Overview

### How Agents Authenticate

1. **An agent is created** via the Admin API → receives a `client_id` and `client_secret`
2. **The agent requests a token** by POSTing credentials to `/oauth/token`
3. **I return a JWT access token** (RS256-signed) + a refresh token
4. **The agent uses the token** as a Bearer token in `Authorization` headers
5. **Any service can verify the token** using my public keys at `/.well-known/jwks.json`
6. **Tokens can be introspected** via `/oauth/introspect` and **revoked** via `/oauth/revoke`
7. **Tokens can be refreshed** via `/oauth/refresh` before they expire

### JWT Access Token Claims

Every access token I issue contains these claims:

| Claim | Type | Description |
|-------|------|-------------|
| `iss` | string | Issuer — `https://auth.example.com` (configurable via `JWT_ISSUER` env var) |
| `sub` | string | Subject — the agent's `client_id` |
| `agent_id` | string | The agent's UUID |
| `org_id` | string | Organization ID the agent belongs to |
| `team_id` | string | Team ID (if assigned) |
| `aud` | string | Audience — `machineauth-api` |
| `iat` | number | Issued at (Unix timestamp) |
| `exp` | number | Expiration (Unix timestamp) |
| `scope` | array | Array of scope strings granted to this token |
| `jti` | string | Unique token ID (used for revocation tracking) |

### Token Lifetimes

| Token | Default Lifetime | Notes |
|-------|-----------------|-------|
| Access Token | 3600s (1 hour) in production, 86400s (24 hours) in development | Configurable via `JWT_ACCESS_TOKEN_EXPIRY` env var |
| Refresh Token | 604800s (7 days) | Fixed |

---

## Complete API Reference

### Service Discovery & Health

---

#### `GET /`
**Description:** Service info and available endpoints.
**Auth:** None

**Response `200`:**
```json
{
  "service": "MachineAuth",
  "status": "running",
  "version": "2.1.60",
  "endpoints": "/oauth/token, /oauth/introspect, /oauth/revoke, /oauth/refresh, /api/agents, /.well-known/jwks.json"
}
```

---

#### `GET /health`
**Description:** Basic health check.
**Auth:** None

**Response `200`:**
```json
{
  "status": "ok"
}
```

---

#### `GET /health/ready`
**Description:** Readiness check with agent count.
**Auth:** None

**Response `200`:**
```json
{
  "status": "ok",
  "timestamp": "2026-03-02T12:00:00Z",
  "agents_count": 5
}
```

---

#### `GET /metrics`
**Description:** Service metrics (token counts, agent counts).
**Auth:** None

**Response `200`:**
```json
{
  "requests": 0,
  "tokens_issued": 42,
  "tokens_refreshed": 10,
  "tokens_revoked": 3,
  "active_tokens": 39,
  "total_agents": 5
}
```

---

#### `GET /.well-known/jwks.json`
**Description:** JSON Web Key Set — public keys for verifying JWT signatures.
**Auth:** None

**Response `200`:**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "key-1",
      "use": "sig",
      "alg": "RS256",
      "n": "<base64url-encoded RSA modulus>",
      "e": "<base64url-encoded RSA exponent>"
    }
  ]
}
```

**How to use:** Fetch this endpoint, find the key matching the `kid` header in the JWT, and verify the RS256 signature using the public key constructed from `n` and `e`.

---

### OAuth 2.0 Endpoints

---

#### `POST /oauth/token`
**Description:** Obtain an access token using client credentials or a refresh token.
**Auth:** None (credentials are in the body)
**Content-Type:** `application/json` or `application/x-www-form-urlencoded`

**Request body (client_credentials grant):**
```json
{
  "grant_type": "client_credentials",
  "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "client_secret": "your-secret-here",
  "scope": "read write"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `grant_type` | string | Yes | Must be `client_credentials` or `refresh_token` |
| `client_id` | string | Yes (for client_credentials) | The agent's client ID (UUID) |
| `client_secret` | string | Yes (for client_credentials) | The agent's client secret |
| `scope` | string | No | Space or comma-separated scopes to request (filtered against agent's allowed scopes) |
| `refresh_token` | string | Yes (for refresh_token) | The refresh token string |

**Response `200`:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read write",
  "issued_at": 1709380800,
  "refresh_token": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

**Error `400`:**
```json
{
  "error": "invalid_client",
  "error_description": "invalid client credentials"
}
```

**Possible error codes:**
| Error | Description |
|-------|-------------|
| `invalid_request` | Missing or malformed parameters |
| `invalid_client` | Bad client_id or client_secret |
| `invalid_grant` | Invalid or expired refresh token |
| `unsupported_grant_type` | grant_type must be `client_credentials` or `refresh_token` |
| `server_error` | Internal error generating token |

---

#### `POST /oauth/introspect`
**Description:** Check if an access token is valid and get its metadata.
**Auth:** None
**Content-Type:** `application/json` or `application/x-www-form-urlencoded`

**Request body:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**Response `200` (active token):**
```json
{
  "active": true,
  "scope": "read write",
  "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "exp": 1709384400,
  "iat": 1709380800,
  "token_type": "Bearer"
}
```

**Response `200` (inactive token):**
```json
{
  "active": false,
  "revoked": true,
  "reason": "revoked"
}
```

| `reason` values | Description |
|----------------|-------------|
| `revoked` | Token was explicitly revoked |
| `expired` | Token has expired |
| *(empty)* | Token is malformed or unrecognized |

---

#### `POST /oauth/revoke`
**Description:** Revoke an access token or refresh token.
**Auth:** None
**Content-Type:** `application/json` or `application/x-www-form-urlencoded`

**Request body:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type_hint": "access_token"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `token` | string | Yes | The token to revoke |
| `token_type_hint` | string | No | `access_token` or `refresh_token`. If omitted, both types are tried. |

**Response `200`:**
```json
{
  "status": "revoked"
}
```

---

#### `POST /oauth/refresh`
**Description:** Refresh an access token using a refresh token.
**Auth:** None
**Content-Type:** `application/json` or `application/x-www-form-urlencoded`

**Request body:**
```json
{
  "grant_type": "refresh_token",
  "refresh_token": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "client_id": "optional-client-id",
  "client_secret": "optional-client-secret"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `refresh_token` | string | Yes | The refresh token |
| `client_id` | string | No | If provided with client_secret, validates client credentials too |
| `client_secret` | string | No | Used with client_id for additional validation |

**Response `200`:** Same as `/oauth/token` response (new access token).

---

### Agent Management (Admin API)

These endpoints manage agent identities. Currently protected by admin session (Basic Auth with `ADMIN_EMAIL` / `ADMIN_PASSWORD`).

---

#### `GET /api/agents`
**Description:** List all agents.
**Auth:** Admin

**Response `200`:**
```json
{
  "agents": [
    {
      "id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      "organization_id": "org-123",
      "team_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      "name": "OpenClaw Agent",
      "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      "scopes": ["read", "write", "admin"],
      "is_active": true,
      "created_at": "2026-03-01T10:00:00Z",
      "updated_at": "2026-03-01T10:00:00Z",
      "expires_at": "2026-06-01T10:00:00Z",
      "token_count": 15,
      "refresh_count": 3,
      "last_activity_at": "2026-03-02T09:00:00Z",
      "last_token_issued_at": "2026-03-02T09:00:00Z"
    }
  ]
}
```

---

#### `POST /api/agents`
**Description:** Create a new agent. Returns client credentials (secret shown only once).
**Auth:** Admin

**Request body:**
```json
{
  "name": "OpenClaw Agent",
  "organization_id": "org-123",
  "team_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "scopes": ["read", "write"],
  "expires_in": 7776000
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable agent name |
| `organization_id` | string | No | Organization to assign agent to |
| `team_id` | string (UUID) | No | Team to assign agent to |
| `scopes` | array of strings | No | Scopes the agent is allowed (defaults to `[]`) |
| `expires_in` | integer | No | Seconds until the agent credential expires. Omit for no expiry. |

**Response `201`:**
```json
{
  "agent": {
    "id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "name": "OpenClaw Agent",
    "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "scopes": ["read", "write"],
    "is_active": true,
    "created_at": "2026-03-02T12:00:00Z",
    "updated_at": "2026-03-02T12:00:00Z",
    "expires_at": "2026-06-01T12:00:00Z",
    "token_count": 0,
    "refresh_count": 0
  },
  "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "client_secret": "abcdef1234567890abcdef1234567890abcdef123456"
}
```

> **IMPORTANT:** The `client_secret` is only returned at creation time. Store it securely. It cannot be retrieved later — only rotated.

---

#### `GET /api/agents/{id}`
**Description:** Get a single agent by UUID.
**Auth:** Admin

**Response `200`:**
```json
{
  "agent": {
    "id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "organization_id": "org-123",
    "name": "OpenClaw Agent",
    "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "scopes": ["read", "write"],
    "is_active": true,
    "created_at": "2026-03-02T12:00:00Z",
    "updated_at": "2026-03-02T12:00:00Z",
    "token_count": 15,
    "refresh_count": 3,
    "last_activity_at": "2026-03-02T09:00:00Z",
    "last_token_issued_at": "2026-03-02T09:00:00Z"
  }
}
```

**Error `404`:** Agent not found.

---

#### `DELETE /api/agents/{id}`
**Description:** Delete an agent permanently.
**Auth:** Admin

**Response `204`:** No content (success).

---

#### `POST /api/agents/{id}` (with `{"action": "rotate"}`)
**Description:** Rotate an agent's client secret. The old secret is immediately invalidated.
**Auth:** Admin

**Request body:**
```json
{
  "action": "rotate"
}
```

**Response `200`:**
```json
{
  "client_secret": "new-secret-value-shown-only-once"
}
```

> **IMPORTANT:** The new secret is only shown once. All existing tokens remain valid until expiry, but the old secret can no longer be used to get new tokens.

---

### Agent Self-Service (Bearer Token Auth)

These endpoints let an authenticated agent manage itself. Requires `Authorization: Bearer <access_token>` header.

---

#### `GET /api/agents/me`
**Description:** Get the authenticated agent's own profile.
**Auth:** Bearer token

**Response `200`:**
```json
{
  "agent": {
    "id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "name": "OpenClaw Agent",
    "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "scopes": ["read", "write"],
    "is_active": true,
    "created_at": "2026-03-02T12:00:00Z",
    "updated_at": "2026-03-02T12:00:00Z",
    "token_count": 15,
    "refresh_count": 3
  }
}
```

---

#### `GET /api/agents/me/usage`
**Description:** Get the authenticated agent's usage statistics and rotation history.
**Auth:** Bearer token

**Response `200`:**
```json
{
  "agent": { "...agent object..." },
  "organization_id": "org-123",
  "team_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "token_count": 15,
  "refresh_count": 3,
  "last_activity_at": "2026-03-02T09:00:00Z",
  "last_token_issued_at": "2026-03-02T09:00:00Z",
  "rotation_history": [
    {
      "rotated_at": "2026-02-15T08:00:00Z",
      "rotated_by_ip": "192.168.1.1"
    }
  ]
}
```

---

#### `POST /api/agents/me/rotate`
**Description:** Agent rotates its own credentials. Old secret is invalidated immediately.
**Auth:** Bearer token

**Response `200`:**
```json
{
  "client_secret": "new-secret-shown-only-once"
}
```

---

#### `POST /api/agents/me/deactivate`
**Description:** Agent deactivates itself. It will no longer be able to get new tokens.
**Auth:** Bearer token

**Response `200`:**
```json
{
  "message": "agent deactivated successfully"
}
```

---

#### `POST /api/agents/me/reactivate`
**Description:** Agent reactivates itself.
**Auth:** Bearer token

**Response `200`:**
```json
{
  "message": "agent reactivated successfully"
}
```

---

#### `DELETE /api/agents/me/delete` or `POST /api/agents/me/delete`
**Description:** Agent permanently deletes itself.
**Auth:** Bearer token

**Response `204`:** No content (success).

---

#### `GET /api/verify`
**Description:** Verify the Bearer token and return the authenticated agent's identity.
**Auth:** Bearer token

**Response `200`:**
```json
{
  "valid": true,
  "agent_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "name": "OpenClaw Agent",
  "scopes": ["read", "write"],
  "is_active": true,
  "token_count": 15
}
```

**Error `401`:** Invalid or expired token.

---

### Organization Management

---

#### `GET /api/organizations`
**Description:** List all organizations.
**Auth:** Admin

**Response `200`:**
```json
{
  "organizations": [
    {
      "id": "org-123",
      "name": "My Organization",
      "slug": "my-org",
      "owner_email": "owner@example.com",
      "jwt_issuer": "",
      "jwt_expiry_secs": 0,
      "allowed_origins": "",
      "plan": "free",
      "created_at": "2026-03-01T10:00:00Z",
      "updated_at": "2026-03-01T10:00:00Z"
    }
  ]
}
```

---

#### `POST /api/organizations`
**Description:** Create a new organization.
**Auth:** Admin

**Request body:**
```json
{
  "name": "My Organization",
  "slug": "my-org",
  "owner_email": "owner@example.com"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Organization display name |
| `slug` | string | Yes | URL-friendly identifier |
| `owner_email` | string | No | Owner's email address |

**Response `201`:** The created organization object.

---

#### `GET /api/organizations/{id}`
**Description:** Get a single organization.
**Auth:** Admin

**Response `200`:** Organization object.

---

#### `PUT /api/organizations/{id}` or `PATCH /api/organizations/{id}`
**Description:** Update an organization.
**Auth:** Admin

**Request body:**
```json
{
  "name": "Updated Name",
  "jwt_issuer": "https://custom-issuer.example.com",
  "jwt_expiry_secs": 7200,
  "allowed_origins": "https://myapp.com,https://other.com"
}
```

---

#### `DELETE /api/organizations/{id}`
**Description:** Delete an organization.
**Auth:** Admin

**Response `204`:** No content.

---

### Team Management

---

#### `GET /api/organizations/{org_id}/teams`
**Description:** List all teams in an organization.
**Auth:** Admin

**Response `200`:**
```json
{
  "teams": [
    {
      "id": "team-uuid",
      "organization_id": "org-123",
      "name": "Backend Team",
      "description": "Handles backend services",
      "created_at": "2026-03-01T10:00:00Z",
      "updated_at": "2026-03-01T10:00:00Z"
    }
  ]
}
```

---

#### `POST /api/organizations/{org_id}/teams`
**Description:** Create a team within an organization.
**Auth:** Admin

**Request body:**
```json
{
  "name": "Backend Team",
  "description": "Handles backend services"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Team name |
| `description` | string | No | Team description |

**Response `201`:** The created team object.

---

#### `GET /api/organizations/{org_id}/agents`
**Description:** List all agents belonging to an organization.
**Auth:** Admin

**Response `200`:**
```json
{
  "agents": [ "...array of agent objects..." ]
}
```

---

#### `POST /api/organizations/{org_id}/agents`
**Description:** Create an agent within a specific organization.
**Auth:** Admin

**Request body:** Same as `POST /api/agents` (the `organization_id` is set from the URL).

**Response `201`:** Same as `POST /api/agents`.

---

### API Key Management

---

#### `GET /api/organizations/{org_id}/api-keys`
**Description:** List all API keys for an organization.
**Auth:** Admin

**Response `200`:**
```json
{
  "api_keys": [
    {
      "id": "key-uuid",
      "organization_id": "org-123",
      "team_id": "team-uuid",
      "name": "Production Key",
      "prefix": "mach_abc123de",
      "last_used_at": "2026-03-02T09:00:00Z",
      "expires_at": "2026-06-01T10:00:00Z",
      "is_active": true,
      "created_at": "2026-03-01T10:00:00Z"
    }
  ]
}
```

---

#### `POST /api/organizations/{org_id}/api-keys`
**Description:** Create an API key for an organization.
**Auth:** Admin

**Request body:**
```json
{
  "name": "Production Key",
  "team_id": "optional-team-uuid",
  "expires_in": 7776000
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Key display name |
| `team_id` | string (UUID) | No | Team to scope the key to |
| `expires_in` | integer | No | Seconds until expiry |

**Response `201`:**
```json
{
  "api_key": {
    "id": "key-uuid",
    "organization_id": "org-123",
    "name": "Production Key",
    "prefix": "mach_abc123de",
    "is_active": true,
    "created_at": "2026-03-02T12:00:00Z"
  },
  "key": "mach_abc123defull-api-key-shown-only-once"
}
```

> **IMPORTANT:** The full `key` value is shown only at creation time.

---

#### `DELETE /api/organizations/{org_id}/api-keys/{key_id}`
**Description:** Revoke (delete) an API key.
**Auth:** Admin

**Response `204`:** No content.

---

### Webhook Management

---

#### `GET /api/webhooks`
**Description:** List all webhook configurations.
**Auth:** Admin

**Response `200`:**
```json
{
  "webhooks": [
    {
      "id": "webhook-uuid",
      "organization_id": "",
      "team_id": "",
      "name": "My Webhook",
      "url": "https://example.com/webhook",
      "events": ["agent.created", "agent.deleted"],
      "is_active": true,
      "max_retries": 10,
      "retry_backoff_base": 2,
      "created_at": "2026-03-01T10:00:00Z",
      "updated_at": "2026-03-01T10:00:00Z",
      "last_tested_at": null,
      "consecutive_fails": 0
    }
  ]
}
```

---

#### `POST /api/webhooks`
**Description:** Create a webhook. A signing secret is generated automatically.
**Auth:** Admin

**Request body:**
```json
{
  "name": "My Webhook",
  "url": "https://example.com/webhook",
  "events": ["agent.created", "agent.deleted", "token.issued"],
  "max_retries": 10,
  "retry_backoff_base": 2
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Webhook display name |
| `url` | string | Yes | The HTTPS URL to deliver events to |
| `events` | array of strings | Yes | Which events to subscribe to (see events list below) |
| `max_retries` | integer | No | Max delivery retries (default: 10) |
| `retry_backoff_base` | integer | No | Exponential backoff base in seconds (default: 2) |

**Response `201`:**
```json
{
  "webhook": { "...webhook object..." },
  "secret": "webhook-signing-secret-shown-only-once"
}
```

> **IMPORTANT:** The `secret` is shown only once. Use it to verify webhook signatures via HMAC-SHA256.

---

#### `GET /api/webhooks/{id}`
**Description:** Get a single webhook configuration.
**Auth:** Admin

**Response `200`:**
```json
{
  "webhook": { "...webhook object..." }
}
```

---

#### `PUT /api/webhooks/{id}`
**Description:** Update a webhook.
**Auth:** Admin

**Request body (all fields optional):**
```json
{
  "name": "Updated Name",
  "url": "https://new-url.example.com/webhook",
  "events": ["agent.created"],
  "is_active": false,
  "max_retries": 5,
  "retry_backoff_base": 3
}
```

---

#### `DELETE /api/webhooks/{id}`
**Description:** Delete a webhook.
**Auth:** Admin

**Response `204`:** No content.

---

#### `POST /api/webhooks/{id}/test`
**Description:** Send a test event to a webhook URL.
**Auth:** Admin

**Request body:**
```json
{
  "event": "webhook.test",
  "payload": "{\"test\": true}"
}
```

**Response `200`:**
```json
{
  "success": true,
  "status_code": 200,
  "error": ""
}
```

---

#### `GET /api/webhooks/{id}/deliveries`
**Description:** List delivery history for a webhook.
**Auth:** Admin

**Response `200`:**
```json
{
  "deliveries": [
    {
      "id": "delivery-uuid",
      "webhook_config_id": "webhook-uuid",
      "event": "agent.created",
      "payload": "{...}",
      "headers": "",
      "status": "delivered",
      "attempts": 1,
      "last_attempt_at": "2026-03-02T12:00:01Z",
      "last_error": "",
      "next_retry_at": null,
      "created_at": "2026-03-02T12:00:00Z"
    }
  ]
}
```

**Delivery status values:**
| Status | Description |
|--------|-------------|
| `pending` | Not yet attempted |
| `delivered` | Successfully delivered (HTTP 2xx) |
| `failed` | Delivery failed, no more retries |
| `retrying` | Failed, will retry with exponential backoff |
| `dead` | Max retries exceeded |

---

#### `GET /api/webhooks/{id}/deliveries/{delivery_id}`
**Description:** Get a single delivery record.
**Auth:** Admin

**Response `200`:**
```json
{
  "delivery": { "...delivery object..." }
}
```

---

#### `GET /api/webhook-events`
**Description:** List all available webhook event types.
**Auth:** Admin

**Response `200`:**
```json
{
  "events": [
    "agent.created",
    "agent.deleted",
    "agent.updated",
    "agent.credentials_rotated",
    "token.issued",
    "token.validation_success",
    "token.validation_failed",
    "webhook.created",
    "webhook.updated",
    "webhook.deleted",
    "webhook.test"
  ]
}
```

---

### Webhook Event Types (Complete Reference)

| Event | Fired When | Payload Fields |
|-------|-----------|----------------|
| `agent.created` | A new agent is created | `event`, `agent_id`, `agent_name`, `client_id`, `timestamp` |
| `agent.deleted` | An agent is deleted | `event`, `agent_id`, `timestamp` |
| `agent.updated` | An agent is modified | `event`, `agent_id`, `timestamp` |
| `agent.credentials_rotated` | An agent's secret is rotated | `event`, `agent_id`, `timestamp` |
| `token.issued` | An access token is issued | `event`, `agent_id`, `timestamp` |
| `token.validation_success` | A token passes introspection | `event`, `agent_id`, `success`, `timestamp` |
| `token.validation_failed` | A token fails introspection | `event`, `agent_id`, `success`, `timestamp` |
| `webhook.created` | A webhook config is created | `event`, `webhook_id`, `timestamp` |
| `webhook.updated` | A webhook config is updated | `event`, `webhook_id`, `timestamp` |
| `webhook.deleted` | A webhook config is deleted | `event`, `webhook_id`, `timestamp` |
| `webhook.test` | A test delivery is triggered | `event`, `webhook_id`, `timestamp` |

### Webhook Signature Verification

Every webhook delivery includes an HMAC-SHA256 signature in the `X-Webhook-Signature` header. To verify:

```
expected = HMAC-SHA256(webhook_secret, request_body)
actual = request.headers["X-Webhook-Signature"]
if expected == actual → delivery is authentic
```

---

## Data Models (Complete Schema)

### Agent
```json
{
  "id": "UUID",
  "organization_id": "string",
  "team_id": "UUID | null",
  "name": "string",
  "client_id": "UUID (auto-generated)",
  "scopes": ["string"],
  "public_key": "string | null (for DPoP/mTLS future use)",
  "is_active": "boolean",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp",
  "expires_at": "ISO 8601 timestamp | null",
  "token_count": "integer",
  "refresh_count": "integer",
  "last_activity_at": "ISO 8601 timestamp | null",
  "last_token_issued_at": "ISO 8601 timestamp | null"
}
```

### Organization
```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "owner_email": "string",
  "jwt_issuer": "string",
  "jwt_expiry_secs": "integer",
  "allowed_origins": "string (comma-separated)",
  "plan": "string (free|pro|enterprise)",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp"
}
```

### Team
```json
{
  "id": "string",
  "organization_id": "string",
  "name": "string",
  "description": "string",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp"
}
```

### API Key
```json
{
  "id": "string",
  "organization_id": "string",
  "team_id": "UUID | null",
  "name": "string",
  "prefix": "string (first 12 chars of key)",
  "last_used_at": "ISO 8601 timestamp | null",
  "expires_at": "ISO 8601 timestamp | null",
  "is_active": "boolean",
  "created_at": "ISO 8601 timestamp"
}
```

### Webhook Config
```json
{
  "id": "UUID",
  "organization_id": "string",
  "team_id": "string",
  "name": "string",
  "url": "string (HTTPS URL)",
  "events": ["string"],
  "is_active": "boolean",
  "max_retries": "integer (default: 10)",
  "retry_backoff_base": "integer (default: 2)",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp",
  "last_tested_at": "ISO 8601 timestamp | null",
  "consecutive_fails": "integer"
}
```

### Webhook Delivery
```json
{
  "id": "UUID",
  "webhook_config_id": "UUID",
  "event": "string",
  "payload": "string (JSON)",
  "headers": "string",
  "status": "pending | delivered | failed | retrying | dead",
  "attempts": "integer",
  "last_attempt_at": "ISO 8601 timestamp | null",
  "last_error": "string",
  "next_retry_at": "ISO 8601 timestamp | null",
  "created_at": "ISO 8601 timestamp"
}
```

### Audit Log
```json
{
  "id": "UUID",
  "agent_id": "UUID | null",
  "action": "string (event type)",
  "ip_address": "string",
  "user_agent": "string",
  "created_at": "ISO 8601 timestamp"
}
```

### Error Response
```json
{
  "error": "string (error code)",
  "error_description": "string (human-readable message)"
}
```

---

## Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|-----------|-------------|
| 400 | `invalid_request` | Missing or malformed request parameters |
| 400 | `unsupported_grant_type` | grant_type not client_credentials or refresh_token |
| 401 | `invalid_client` | Bad client_id or client_secret |
| 401 | `invalid_grant` | Invalid or expired refresh token |
| 401 | (plain text) | Missing/invalid Authorization header for Bearer-auth endpoints |
| 404 | (plain text) | Resource not found |
| 405 | (plain text) | HTTP method not allowed for this endpoint |
| 500 | `server_error` | Internal server error |

---

## Configuration (Environment Variables)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port |
| `ENV` | `development` | Environment (`development` or `production`) |
| `DATABASE_URL` | `json:machineauth.json` | PostgreSQL URL or `json:filename.json` for JSON file storage |
| `JWT_SIGNING_ALGORITHM` | `RS256` | JWT signing algorithm |
| `JWT_KEY_ID` | `key-1` | Key ID in JWKS |
| `JWT_ACCESS_TOKEN_EXPIRY` | `3600` | Access token lifetime in seconds (overridden to 86400 in development) |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins (comma-separated) |
| `REQUIRE_HTTPS` | `false` | Require HTTPS connections |
| `ADMIN_EMAIL` | `admin@example.com` | Admin login email |
| `ADMIN_PASSWORD` | `changeme` | Admin login password |
| `WEBHOOK_WORKER_COUNT` | `3` | Number of concurrent webhook delivery workers |
| `WEBHOOK_MAX_RETRIES` | `10` | Max webhook delivery retry attempts |
| `WEBHOOK_TIMEOUT_SECS` | `10` | Webhook HTTP request timeout |

---

## Quick Start for AI Agents

### Step 1: Get Your Credentials
Your admin creates an agent for you and gives you a `client_id` and `client_secret`.

### Step 2: Get an Access Token
```bash
curl -X POST https://auth.writesomething.fun/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "YOUR_CLIENT_ID",
    "client_secret": "YOUR_CLIENT_SECRET"
  }'
```

### Step 3: Use the Token
```bash
curl https://auth.writesomething.fun/api/agents/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Step 4: Refresh Before Expiry
```bash
curl -X POST https://auth.writesomething.fun/oauth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### Step 5: Verify Your Identity
```bash
curl https://auth.writesomething.fun/api/verify \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                    MachineAuth                        │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌───────┐  ┌────────┐  │
│  │  OAuth    │  │  Agent   │  │ Audit │  │Webhook │  │
│  │  Engine   │  │  Manager │  │  Log  │  │ System │  │
│  │          │  │          │  │       │  │        │  │
│  │ /oauth/* │  │/api/agent│  │ Every │  │ HMAC   │  │
│  │ RS256 JWT│  │ CRUD +   │  │ action│  │ signed │  │
│  │ JWKS     │  │ self-svc │  │logged │  │delivery│  │
│  └────┬─────┘  └────┬─────┘  └───┬───┘  └───┬────┘  │
│       │              │            │          │        │
│  ┌────┴──────────────┴────────────┴──────────┴─────┐  │
│  │              Core Services Layer                │  │
│  │  Organizations · Teams · API Keys · Tokens      │  │
│  └─────────────────────┬───────────────────────────┘  │
│                        │                              │
│  ┌─────────────────────┴───────────────────────────┐  │
│  │              Storage Layer                      │  │
│  │    PostgreSQL 15 (production)                   │  │
│  │    JSON File   (development)                    │  │
│  └─────────────────────────────────────────────────┘  │
│                                                      │
│  ┌─────────────────────────────────────────────────┐  │
│  │              Middleware                          │  │
│  │  CORS · Request Logging · JWT Auth              │  │
│  └─────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

---

## Endpoint Summary Table

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/` | None | Service info |
| GET | `/health` | None | Health check |
| GET | `/health/ready` | None | Readiness + agent count |
| GET | `/metrics` | None | Token & agent metrics |
| GET | `/.well-known/jwks.json` | None | RSA public keys (JWKS) |
| POST | `/oauth/token` | None | Get access + refresh token |
| POST | `/oauth/introspect` | None | Validate a token |
| POST | `/oauth/revoke` | None | Revoke a token |
| POST | `/oauth/refresh` | None | Refresh an access token |
| GET | `/api/agents` | Admin | List all agents |
| POST | `/api/agents` | Admin | Create agent |
| GET | `/api/agents/{id}` | Admin | Get agent |
| DELETE | `/api/agents/{id}` | Admin | Delete agent |
| POST | `/api/agents/{id}` | Admin | Rotate agent secret |
| GET | `/api/agents/me` | Bearer | Get own profile |
| GET | `/api/agents/me/usage` | Bearer | Get own usage stats |
| POST | `/api/agents/me/rotate` | Bearer | Rotate own secret |
| POST | `/api/agents/me/deactivate` | Bearer | Deactivate self |
| POST | `/api/agents/me/reactivate` | Bearer | Reactivate self |
| POST/DELETE | `/api/agents/me/delete` | Bearer | Delete self |
| GET | `/api/verify` | Bearer | Verify token + get identity |
| GET | `/api/organizations` | Admin | List organizations |
| POST | `/api/organizations` | Admin | Create organization |
| GET | `/api/organizations/{id}` | Admin | Get organization |
| PUT/PATCH | `/api/organizations/{id}` | Admin | Update organization |
| DELETE | `/api/organizations/{id}` | Admin | Delete organization |
| GET | `/api/organizations/{id}/teams` | Admin | List teams |
| POST | `/api/organizations/{id}/teams` | Admin | Create team |
| GET | `/api/organizations/{id}/agents` | Admin | List org agents |
| POST | `/api/organizations/{id}/agents` | Admin | Create agent in org |
| GET | `/api/organizations/{id}/api-keys` | Admin | List API keys |
| POST | `/api/organizations/{id}/api-keys` | Admin | Create API key |
| DELETE | `/api/organizations/{id}/api-keys/{key_id}` | Admin | Revoke API key |
| GET | `/api/webhooks` | Admin | List webhooks |
| POST | `/api/webhooks` | Admin | Create webhook |
| GET | `/api/webhooks/{id}` | Admin | Get webhook |
| PUT | `/api/webhooks/{id}` | Admin | Update webhook |
| DELETE | `/api/webhooks/{id}` | Admin | Delete webhook |
| POST | `/api/webhooks/{id}/test` | Admin | Test webhook |
| GET | `/api/webhooks/{id}/deliveries` | Admin | Delivery history |
| GET | `/api/webhooks/{id}/deliveries/{did}` | Admin | Get delivery |
| GET | `/api/webhook-events` | Admin | List event types |

---

*Built with Go 1.23. Secured with RS256. Owned by you.*
*[github.com/mandarwagh9/MachineAuth](https://github.com/mandarwagh9/MachineAuth)*
