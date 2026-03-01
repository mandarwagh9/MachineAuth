# Multi-Tenant Architecture Design

## Current State
- Single-tenant: all agents belong to the system
- No organization/team concept

## Target State
- **Organization** - Top-level tenant (company, team)
- **Team** - Group of agents within an organization  
- **Agent** - Machine identity (belongs to a team)
- **API Keys** - Alternative to client credentials

## Data Model

### Organization
```go
type Organization struct {
    ID             uuid.UUID  `json:"id"`
    Name           string     `json:"name"`
    Slug           string     `json:"slug"`           // unique identifier for URLs
    OwnerEmail     string     `json:"owner_email"`
    JWTIssuer      string     `json:"jwt_issuer"`     // custom JWT issuer
    JWTExpirySecs  int        `json:"jwt_expiry_secs"`
    AllowedOrigins string     `json:"allowed_origins"`
    Plan           string     `json:"plan"`           // free, pro, enterprise
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}
```

### Team
```go
type Team struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    Name           string     `json:"name"`
    Description    string     `json:"description"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}
```

### Agent (Updated)
```go
type Agent struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`  // NEW
    TeamID         *uuid.UUID `json:"team_id"`          // NEW
    Name           string     `json:"name"`
    ClientID       string     `json:"client_id"`
    // ... existing fields
}
```

### APIKey (NEW - Alternative to client credentials)
```go
type APIKey struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    TeamID         *uuid.UUID `json:"team_id"`
    Name           string     `json:"name"`
    KeyHash        string     `json:"key_hash"`        // hash of sk_live_xxx
    Prefix         string     `json:"prefix"`         // sk_live_xxx (first 12 chars)
    LastUsedAt     *time.Time `json:"last_used_at"`
    ExpiresAt      *time.Time `json:"expires_at"`
    IsActive       bool       `json:"is_active"`
    CreatedAt      time.Time  `json:"created_at"`
}
```

## API Endpoints

### Organizations
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/organizations | List user's organizations |
| POST | /api/organizations | Create organization |
| GET | /api/organizations/:id | Get organization |
| PUT | /api/organizations/:id | Update organization |
| DELETE | /api/organizations/:id | Delete organization |

### Teams
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/organizations/:org/teams | List teams |
| POST | /api/organizations/:org/teams | Create team |
| GET | /api/teams/:id | Get team |
| PUT | /api/teams/:id | Update team |
| DELETE | /api/teams/:id | Delete team |

### Agents (Updated)
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/teams/:team/agents | List agents in team |
| POST | /api/teams/:team/agents | Create agent in team |
| GET | /api/agents/:id | Get agent (any team) |
| PUT | /api/agents/:id | Update agent |
| DELETE | /api/agents/:id | Delete agent |

### API Keys
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/organizations/:org/api-keys | List API keys |
| POST | /api/organizations/:org/api-keys | Create API key |
| DELETE | /api/organizations/:org/api-keys/:id | Revoke API key |

## JWT Claims (Updated)
```json
{
    "iss": "https://auth.agentauth.io",
    "sub": "client_id",
    "agent_id": "agent_uuid",
    "org_id": "organization_uuid",     // NEW
    "team_id": "team_uuid",            // NEW (optional)
    "scope": ["read", "write"],
    "jti": "token_uuid",
    "exp": 1234567890,
    "iat": 1234567890
}
```

## Authentication Flow

### Using Client Credentials (Existing)
1. Agent calls `/oauth/token` with client_id + client_secret
2. Server validates credentials + checks org/team membership
3. Returns JWT with org_id and team_id claims

### Using API Key (NEW)
1. Agent calls API with `Authorization: Bearer sk_live_xxx`
2. Server hashes the key, looks up in APIKeys table
3. Validates organization and team membership
4. Returns JWT with org_id and team_id claims

## Implementation Priority

1. **Phase 1**: Add Organization model + CRUD
2. **Phase 2**: Add Team model + CRUD  
3. **Phase 3**: Update Agent to belong to org/team
4. **Phase 4**: Add API Keys support
5. **Phase 5**: Add organization switcher to Admin UI
