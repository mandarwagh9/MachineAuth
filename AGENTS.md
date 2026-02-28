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
# Install dependencies
go mod download

# Build server
go build -o bin/server ./cmd/server

# Run server
go run ./cmd/server

# Run all tests
go test -v ./...

# Run single test (when tests exist)
go test -v -run TestFunctionName ./internal/services/...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Lint (requires golangci-lint)
golangci-lint run

# Format
go fmt ./... && go vet ./...
```

### Frontend (React)

```bash
# Install dependencies
cd web && npm install

# Development (port 3000)
cd web && npm run dev

# Production build
cd web && npm run build

# Lint
cd web && npm run lint

# TypeScript type check
cd web && npx tsc --noEmit
```

## Code Style Guidelines

### Go (Backend)

**Imports**: stdlib → external → internal. Use `goimports`.

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

**Naming**: Packages lowercase, exported PascalCase, unexported camelCase. Constants PascalCase.

**Error Handling**: Wrap with context using `fmt.Errorf` with `%w`:
```go
if err != nil {
	return nil, fmt.Errorf("failed to create agent: %w", err)
}
```

**HTTP Handlers**: Return structured JSON responses, use proper HTTP status codes:
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(models.AgentResponse{Agent: *agent})
```

**Types**: Use precise types, pointers for optional values. Embed `uuid.UUID` directly:
```go
type Agent struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
```

**Logging**: Use `log.Printf` for errors with context:
```go
log.Printf("failed to create agent: %v", err)
```

### React/TypeScript (Frontend)

**Imports**: Use absolute imports with `@/` alias.

```typescript
import { useState, useEffect } from 'react';
import { AgentService } from '@/services/api';
import { Agent } from '@/types';
```

**Naming**: Components PascalCase, functions camelCase, files kebab-case.

**Types**: Define explicit types, avoid `any`.
```typescript
interface Agent {
  id: string;
  name: string;
  client_id: string;
  scopes: string[];
}
```

**Error Handling**: Try/catch with user-friendly messages.

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
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── pages/          # Page components
│   │   ├── services/       # API services
│   │   ├── types/          # TypeScript types
│   │   └── App.tsx
│   └── package.json
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
- API baseURL via `VITE_API_URL` env var
- Admin credentials: `admin` / `admin` (change in web/src/App.tsx before deployment)
- Alternate Go entry point: `go run server-main.go` (runs on port 8081)

## Key Conventions

### Backend Layer Architecture
- **Handlers** (`internal/handlers/`): HTTP request/response handling, validation
- **Services** (`internal/services/`): Business logic, orchestration
- **Models** (`internal/models/`): API request/response types, JSON serialization
- **DB** (`internal/db/`): Database operations, storage abstraction

### Frontend State Management
- Use React hooks (`useState`, `useEffect`) for local state
- Use `react-hook-form` for form handling
- Use `sonner` for toast notifications

### JSON Field Naming
- Backend uses `snake_case` for JSON (e.g., `client_id`, `expires_at`)
- Frontend can use either convention; match API response
