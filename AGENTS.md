# AGENTS.md - MachineAuth Development Guide

## Project Overview

MachineAuth is a self-hosted AI Agent Authentication SaaS platform (v2.12.61) providing OAuth 2.0-based authentication for AI agents and machine-to-machine communication.

## Tech Stack

- **Backend**: Go 1.21+ (module: `machineauth`)
- **Frontend**: React 18 + TypeScript + Vite + Tailwind CSS
- **Database**: PostgreSQL 15+ (or JSON file for development)
- **SDKs**: `sdk/typescript` (`@machineauth/sdk`) and `sdk/python` (`machineauth`)

## Build, Lint, and Test Commands

### Backend (Go)

```bash
go mod download                    # Install dependencies
go build -o bin/server ./cmd/server # Build server
go run ./cmd/server                 # Run server
go test -v ./...                    # Run all tests
go test -v -run TestName ./path/... # Run single test
go test -v -coverprofile=cov.out ./... # Tests with coverage
golangci-lint run                  # Lint (requires golangci-lint)
go fmt ./... && go vet ./...       # Format
```

### Frontend (React)

```bash
cd web && npm install              # Install dependencies
cd web && npm run dev              # Development (port 3000)
cd web && npm run build            # Production build
cd web && npm run lint             # Lint
cd web && npx tsc --noEmit         # TypeScript check
```

## Code Style Guidelines

### Go (Backend)

**Imports**: stdlib → external → internal (use `goimports`)

```go
import (
	"context"
	"encoding/json"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"machineauth/internal/config"
	"machineauth/internal/models"
)
```

**Naming**: Packages lowercase, exported PascalCase, unexported camelCase, constants PascalCase.

**Error Handling**: Wrap with context using `fmt.Errorf` with `%w`:
```go
if err != nil {
	return nil, fmt.Errorf("failed to create agent: %w", err)
}
```

**HTTP Handlers**: Return structured JSON, proper status codes:
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(models.AgentResponse{Agent: *agent})
```

**Types**: Use precise types, pointers for optional values, embed `uuid.UUID` directly:
```go
type Agent struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
```

**Logging**: Use `log.Printf` with context.

### React/TypeScript (Frontend)

**Imports**: Use absolute imports with `@/` alias.

```typescript
import { useState } from 'react';
import { AgentService } from '@/services/api';
import { Agent } from '@/types';
```

**Naming**: Components PascalCase, functions camelCase, files kebab-case.

**Types**: Define explicit types, avoid `any`.

**Styling**: Tailwind CSS with HSL color variables.

## Project Structure

```
machineauth/
├── cmd/server/main.go      # Entry point
├── internal/
│   ├── config/             # Configuration
│   ├── db/                 # Database layer
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Data models
│   ├── services/           # Business logic
│   └── utils/              # Utilities
├── web/                    # React frontend
│   └── src/
│       ├── components/     # React components
│       ├── pages/          # Page components
│       ├── services/       # API services
│       ├── types/          # TypeScript types
│       └── App.tsx
└── docker-compose.yml
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/oauth/token` | Get access token |
| POST | `/oauth/introspect` | Validate token |
| POST | `/oauth/revoke` | Revoke token |
| GET | `/.well-known/jwks.json` | Public keys |
| GET | `/health` | Health check |
| GET | `/health/ready` | Readiness check |
| GET | `/metrics` | Service metrics |
| GET | `/api/agents` | List agents |
| POST | `/api/agents` | Create agent |
| GET | `/api/agents/:id` | Get agent |
| DELETE | `/api/agents/:id` | Delete agent |
| POST | `/api/agents/:id/rotate` | Rotate credentials |
| POST | `/api/webhooks` | Create webhook |
| GET | `/api/webhooks` | List webhooks |
| GET | `/api/webhooks/:id` | Get webhook |
| PUT | `/api/webhooks/:id` | Update webhook |
| DELETE | `/api/webhooks/:id` | Delete webhook |
| POST | `/api/webhooks/:id/test` | Test webhook |
| GET | `/api/webhooks/:id/deliveries` | Delivery history |
| GET | `/api/webhook-events` | List event types |

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
PORT=8080
ENV=development
DATABASE_URL=postgresql://user:pass@localhost:5432/agentauth
# Dev fallback: DATABASE_URL=json:machineauth.json
JWT_SIGNING_ALGORITHM=RS256
JWT_KEY_ID=key-1
JWT_ACCESS_TOKEN_EXPIRY=3600
ALLOWED_ORIGINS=http://localhost:3000

# Admin
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=changeme

# Webhooks
WEBHOOK_WORKER_COUNT=3
WEBHOOK_MAX_RETRIES=10
WEBHOOK_TIMEOUT_SECS=10
```

## Development Notes

- Frontend proxies `/api`, `/oauth`, `/.well-known`, `/health`, `/metrics` to `localhost:8081`
- Use `@/` for absolute imports (maps to `./src/`)
- Admin credentials: `admin` / `admin` (change before deployment)
- Alternate Go entry: `go run server-main.go` (port 8081)

## Key Conventions

- **Handlers**: HTTP request/response handling, validation (see [`internal/handlers/agents.go`](internal/handlers/agents.go))
- **Services**: Business logic, orchestration (see [`internal/services/`](internal/services/))
- **Models**: API request/response types, JSON serialization (see [`internal/models/models.go`](internal/models/models.go))
- **DB**: Storage abstraction; `db.Connect` returns either PostgreSQL or `JSONDB` based on `DATABASE_URL` prefix (see [`internal/db/db.go`](internal/db/db.go))
- Backend uses `snake_case` JSON (e.g., `client_id`, `expires_at`); `ClientSecretHash` tagged `json:"-"` (never serialized)
- Frontend: React hooks for state, `react-hook-form` for forms, `sonner` for toasts, `axios` for API calls
- Frontend API base URL from `VITE_API_URL` env var; falls back to production URL — set it to `http://localhost:8081` locally
- Router registration: `cmd/server/main.go` wires all routes manually onto `http.NewServeMux`; middleware (`JWTAuth`, `AdminAuth`) wraps handlers via closures
- `middleware.AgentIDKey` context key carries `uuid.UUID`; retrieve with `middleware.GetAgentIDFromContext(ctx)`
- Webhook system: `services.DeliveryWorker` goroutine pool dispatches deliveries; `AuditService.SetWebhookService` injects webhook triggering into audit events
- Version string lives in `cmd/server/main.go` root handler response (`"version":"2.12.61"`)
