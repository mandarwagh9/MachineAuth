# AGENTS.md - MachineAuth Development Guide

## Project Overview

MachineAuth is a self-hosted AI Agent Authentication SaaS platform providing OAuth 2.0-based authentication for AI agents and machine-to-machine communication.

## Tech Stack

- **Backend**: Go 1.21+ (module: `machineauth`)
- **Frontend**: React 18 + TypeScript + Vite + Tailwind CSS
- **Database**: PostgreSQL 15+ (or JSON file for development)

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
```

## Development Notes

- Frontend proxies `/api`, `/oauth`, `/.well-known`, `/health`, `/metrics` to `localhost:8081`
- Use `@/` for absolute imports (maps to `./src/`)
- Admin credentials: `admin` / `admin` (change before deployment)
- Alternate Go entry: `go run server-main.go` (port 8081)

## Key Conventions

- **Handlers**: HTTP request/response handling, validation
- **Services**: Business logic, orchestration
- **Models**: API request/response types, JSON serialization
- **DB**: Database operations, storage abstraction
- Backend uses `snake_case` JSON (e.g., `client_id`, `expires_at`)
- Frontend: React hooks for state, `react-hook-form` for forms, `sonner` for toasts
