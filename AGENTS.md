# AGENTS.md - Machine Authentication Platform

## Project Overview

MachineAuth is a self-hosted AI Agent Authentication SaaS platform that provides OAuth 2.0-based authentication for AI agents and machine-to-machine communication. Customers deploy this software in their own infrastructure.

## Tech Stack

- **Backend**: Go 1.21+
- **Frontend**: React 18 + TypeScript
- **Database**: PostgreSQL 15+
- **Deployment**: Docker + Kubernetes

## Build, Lint, and Test Commands

### Backend (Go)

```bash
# Install dependencies
go mod download

# Build the server
go build -o bin/server ./cmd/server

# Run the server
go run ./cmd/server

# Run tests
go test -v ./...

# Run a single test
go test -v -run TestAgentService_CreateAgent ./internal/services/...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...
go vet ./...
```

### Frontend (React)

```bash
# Install dependencies
cd web && npm install

# Build for production
cd web && npm run build

# Run development server
cd web && npm run dev

# Run tests
cd web && npm test

# Run a single test file
cd web && npm test -- --testPathPattern=AgentList.test.tsx

# Run tests in watch mode
cd web && npm test -- --watch

# Lint
cd web && npm run lint

# Type check
cd web && npm run typecheck
```

## Code Style Guidelines

### Go (Backend)

#### Imports
- Use standard library first, then third-party, then internal
- Group imports: stdlib, external, internal
- Use `goimports` or let IDE handle this

```go
import (
    "context"
    "encoding/json"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"

    "agentauth/internal/config"
    "agentauth/internal/models"
)
```

#### Naming Conventions
- **Packages**: lowercase, short (e.g., `services`, `handlers`, `models`)
- **Types**: PascalCase (e.g., `AgentService`, `TokenRequest`)
- **Functions/Variables**: camelCase (e.g., `createAgent`, `clientID`)
- **Constants**: PascalCase for exported, snake_case for private (e.g., `DefaultTokenExpiry`, `maxTokenLifetime`)
- **Interfaces**: Add `er` suffix for single-method interfaces (e.g., `Reader`, `Writer`)

#### Error Handling
- Return errors with context using `fmt.Errorf("failed to %s: %w", action, err)`
- Use sentinel errors for known conditions
- Wrap errors with `%w` to preserve the original error chain

```go
if err != nil {
    return nil, fmt.Errorf("failed to create agent: %w", err)
}
```

#### Types
- Use precise types (e.g., `time.Duration` over `int`)
- Define custom types for domain concepts
- Use pointers for optional values or when mutation is needed

```go
type Agent struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    ClientID    string    `json:"client_id"`
    Scopes      []string  `json:"scopes,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}
```

#### Logging
- Use structured logging with context
- Include request IDs for traceability
- Log at appropriate levels (debug, info, warn, error)

```go
log.Info("agent created", "agent_id", agent.ID, "client_id", agent.ClientID)
```

### React/TypeScript (Frontend)

#### Imports
- Use absolute imports when configured, otherwise relative
- Order: React/libraries → internal components/utils → styles/types

```typescript
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

import { AgentService } from '@/services/agent';
import { Agent } from '@/types';
import { Button } from '@/components/ui';
```

#### Naming Conventions
- **Components**: PascalCase (e.g., `AgentList`, `TokenGenerator`)
- **Functions/Variables**: camelCase (e.g., `handleSubmit`, `agentList`)
- **Constants**: UPPER_SNAKE_CASE for true constants
- **Files**: kebab-case for components (e.g., `agent-list.tsx`)

#### Types
- Define explicit types for all props and state
- Use interfaces for objects, types for unions/aliases
- Avoid `any`, use `unknown` when type is truly unknown

```typescript
interface Agent {
  id: string;
  name: string;
  clientId: string;
  scopes: string[];
  createdAt: string;
  expiresAt?: string;
}

interface AgentListProps {
  agents: Agent[];
  onSelect: (id: string) => void;
}
```

#### React Patterns
- Use functional components with hooks
- Keep components small and focused
- Extract reusable logic into custom hooks
- Use TypeScript generics for flexible components

```typescript
export const AgentList: React.FC<AgentListProps> = ({ agents, onSelect }) => {
  const [loading, setLoading] = useState(false);

  // Custom hook for agent operations
  const { createAgent, deleteAgent } = useAgentService();

  return (
    <div className="agent-list">
      {agents.map(agent => (
        <AgentCard
          key={agent.id}
          agent={agent}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
};
```

#### Error Handling
- Handle API errors with try/catch
- Show user-friendly error messages
- Log errors for debugging

```typescript
try {
  await createAgent(agentData);
} catch (error) {
  if (error instanceof ApiError) {
    setError(error.message);
  } else {
    setError('Failed to create agent');
    console.error(error);
  }
}
```

#### Styling
- Use CSS modules or Tailwind CSS
- Keep styles co-located with components
- Use consistent spacing and typography

## Project Structure

```
agentauth/
├── cmd/
│   └── server/           # Main application entry point
│       └── main.go
├── internal/
│   ├── config/           # Configuration management
│   │   └── config.go
│   ├── db/               # Database layer
│   │   ├── migrations/  # SQL migrations
│   │   └── db.go         # Database connection
│   ├── handlers/         # HTTP handlers
│   │   ├── auth.go       # OAuth endpoints
│   │   └── agents.go     # Agent management
│   ├── middleware/       # HTTP middleware
│   │   ├── auth.go       # JWT validation
│   │   └── logging.go    # Request logging
│   ├── models/           # Data models
│   │   └── models.go
│   ├── services/         # Business logic
│   │   ├── agent.go      # Agent service
│   │   └── token.go      # Token service
│   └── utils/            # Utilities
│       └── crypto.go     # Key generation
├── web/                  # React frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── pages/        # Page components
│   │   ├── services/     # API services
│   │   ├── types/        # TypeScript types
│   │   └── App.tsx
│   └── package.json
├── docker/               # Docker configuration
│   ├── Dockerfile.server
│   └── Dockerfile.web
├── docker-compose.yml
└── README.md
```

## API Endpoints

### OAuth 2.0 Token Endpoint
- `POST /oauth/token` - Get access token (Client Credentials flow)

### Agent Management
- `POST /api/agents` - Create new agent
- `GET /api/agents` - List all agents
- `GET /api/agents/:id` - Get agent details
- `DELETE /api/agents/:id` - Revoke/delete agent
- `POST /api/agents/:id/rotate` - Rotate agent credentials

### JWKS
- `GET /.well-known/jwks.json` - Get public keys for token verification

### Webhooks
- `POST /api/webhooks` - Create new webhook
- `GET /api/webhooks` - List all webhooks
- `GET /api/webhooks/:id` - Get webhook details
- `PUT /api/webhooks/:id` - Update webhook configuration
- `DELETE /api/webhooks/:id` - Delete webhook
- `POST /api/webhooks/:id/test` - Send test delivery
- `GET /api/webhooks/:id/deliveries` - Get delivery history
- `GET /api/webhook-events` - List available event types

### Health
- `GET /health` - Health check endpoint

## Database Schema

### agents table
- `id` UUID PRIMARY KEY
- `name` VARCHAR(255) NOT NULL
- `client_id` VARCHAR(255) UNIQUE NOT NULL
- `client_secret_hash` VARCHAR(255) NOT NULL
- `scopes` TEXT[] DEFAULT '{}'
- `public_key` TEXT (for DPoP/mTLS)
- `is_active` BOOLEAN DEFAULT true
- `created_at` TIMESTAMP
- `updated_at` TIMESTAMP
- `expires_at` TIMESTAMP (optional)

### audit_logs table
- `id` UUID PRIMARY KEY
- `agent_id` UUID REFERENCES agents(id)
- `action` VARCHAR(100)
- `ip_address` VARCHAR(45)
- `user_agent` TEXT
- `created_at` TIMESTAMP

### webhook_configs table
- `id` UUID PRIMARY KEY
- `organization_id` VARCHAR(255)
- `team_id` VARCHAR(255)
- `name` VARCHAR(255) NOT NULL
- `url` TEXT NOT NULL
- `secret` VARCHAR(255) NOT NULL
- `events` TEXT[] NOT NULL
- `is_active` BOOLEAN DEFAULT true
- `max_retries` INTEGER DEFAULT 10
- `retry_backoff_base` INTEGER DEFAULT 2
- `created_at` TIMESTAMP
- `updated_at` TIMESTAMP
- `last_tested_at` TIMESTAMP (optional)
- `consecutive_fails` INTEGER DEFAULT 0

### webhook_deliveries table
- `id` UUID PRIMARY KEY
- `webhook_config_id` UUID REFERENCES webhook_configs(id)
- `event` VARCHAR(100)
- `payload` TEXT
- `headers` TEXT
- `status` VARCHAR(50) (pending, delivered, failed, retrying, dead)
- `attempts` INTEGER DEFAULT 0
- `last_attempt_at` TIMESTAMP (optional)
- `last_error` TEXT
- `next_retry_at` TIMESTAMP (optional)
- `created_at` TIMESTAMP

## Environment Variables

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/agentauth

# JWT
JWT_SIGNING_ALGORITHM=RS256
JWT_KEY_ID=key-1
JWT_ACCESS_TOKEN_EXPIRY=3600

# Security
REQUIRE_HTTPS=false
ALLOWED_ORIGINS=http://localhost:3000

# Admin
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=changeme

# Webhooks
WEBHOOK_WORKER_COUNT=3
WEBHOOK_MAX_RETRIES=10
WEBHOOK_TIMEOUT_SECS=10
```

## Common Development Tasks

### Creating a new agent
1. User calls `POST /api/agents` with name and scopes
2. Server generates unique client_id (UUID) and client_secret
3. Server stores secret hash (bcrypt/argon2)
4. Returns client_id and client_secret (one-time)

### Token issuance flow
1. Agent calls `POST /oauth/token` with client_id and client_secret
2. Server validates credentials
3. Server creates JWT with claims (sub, iss, aud, exp, scope)
4. Returns access_token

### Token validation flow
1. Resource server fetches JWKS from `GET /.well-known/jwks.json`
2. Resource server validates JWT signature
3. Resource server checks claims (exp, iss, aud)
4. Resource server checks scopes

## Testing Strategy

- **Unit tests**: Test individual services and utilities
- **Integration tests**: Test HTTP handlers with test database
- **E2E tests**: Test complete user flows (frontend)
- Use table-driven tests for Go
- Use React Testing Library for component tests
