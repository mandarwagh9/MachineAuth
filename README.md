<h1 align="center">
  <br>
  🔐 MachineAuth
  <br>
</h1>

<p align="center">
  <strong>Self-hosted OAuth 2.0 authentication for AI agents and machine-to-machine communication.</strong>
</p>

<p align="center">
  <a href="https://github.com/mandarwagh9/MachineAuth/stargazers">
    <img src="https://img.shields.io/github/stars/mandarwagh9/MachineAuth?style=flat-square&color=ffd700" alt="Stars">
  </a>
  <a href="https://github.com/mandarwagh9/MachineAuth/issues">
    <img src="https://img.shields.io/github/issues/mandarwagh9/MachineAuth?style=flat-square" alt="Issues">
  </a>
  <a href="https://github.com/mandarwagh9/MachineAuth/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/mandarwagh9/MachineAuth?style=flat-square" alt="License">
  </a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.21+">
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=black" alt="React 18">
  <img src="https://img.shields.io/badge/TypeScript-5.3-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript">
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#features">Features</a> •
  <a href="#admin-dashboard">Dashboard</a> •
  <a href="#api-reference">API</a> •
  <a href="#sdks">SDKs</a> •
  <a href="#deployment">Deploy</a>
</p>

---

## What is MachineAuth?

MachineAuth is a **self-hosted OAuth 2.0 server** purpose-built for authenticating AI agents and machines. Instead of sharing long-lived API keys, your agents authenticate using **OAuth 2.0 Client Credentials** and receive short-lived **RS256-signed JWT tokens**.

Think of it as Auth0, but for bots — lightweight, self-hosted, zero external dependencies.

### Why MachineAuth?

| Problem | MachineAuth Solution |
|---------|---------------------|
| Sharing long-lived API keys | Short-lived JWTs with configurable expiry |
| No credential rotation | One-click rotation, zero downtime |
| No visibility into agent activity | Per-agent usage tracking, audit logs, metrics |
| Complex auth infrastructure | Single binary, JSON file storage for dev, Postgres for prod |
| No webhook notifications | Built-in webhook system with retry & delivery tracking |
| Multi-tenant headaches | Native organization & team scoping with JWT claims |

---

## Quick Start

```bash
git clone https://github.com/mandarwagh9/MachineAuth.git
cd MachineAuth
go run ./cmd/server
```

Server starts on `http://localhost:8080`. No database needed — uses JSON file storage by default.

```bash
# 1. Create an agent
curl -s -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read", "write"]}' | jq .

# 2. Get a token
curl -s -X POST http://localhost:8080/oauth/token \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" | jq .

# 3. Use the token
curl -s -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8080/api/agents/me | jq .
```

**Live demo:** [https://auth.writesomething.fun](https://auth.writesomething.fun)

---

## Features

### Core Authentication
- **OAuth 2.0 Client Credentials** — Industry-standard M2M authentication flow
- **RS256 JWT Tokens** — Asymmetric signing with auto-generated RSA keys
- **Token Introspection** — Validate tokens via RFC 7662 compliant endpoint
- **Token Revocation** — Invalidate tokens instantly via RFC 7009
- **Refresh Tokens** — Renew access without re-authenticating
- **JWKS Endpoint** — Public key discovery at `/.well-known/jwks.json`

### Agent Management
- **Full CRUD** — Create, list, view, update, delete agents
- **Credential Rotation** — Rotate client secrets with zero downtime
- **Scoped Access** — Fine-grained scopes per agent
- **Usage Tracking** — Token count, refresh count, last activity per agent
- **Agent Self-Service** — Agents manage their own lifecycle via JWT auth
- **Activation Control** — Deactivate/reactivate agents without deletion

### Multi-Tenant
- **Organizations** — Isolated tenant environments with unique slugs
- **Teams** — Group agents under teams within organizations
- **Org-Scoped Agents** — Agents belong to orgs, JWT claims include `org_id`/`team_id`
- **API Keys** — Per-organization API key management

### Webhooks
- **Event Notifications** — Real-time HTTP callbacks for agent/token events
- **Delivery Tracking** — Full delivery history with status, attempts, errors
- **Automatic Retries** — Exponential backoff with configurable retry count
- **Webhook Testing** — Send test payloads to verify endpoint connectivity
- **Background Workers** — Async delivery processing (configurable worker count)

### Operations
- **Health Checks** — `/health` and `/health/ready` endpoints
- **Metrics** — Token/agent/error statistics at `/metrics`
- **Audit Logging** — Track all agent and token operations
- **CORS** — Configurable cross-origin settings
- **Graceful Shutdown** — Clean shutdown on SIGINT/SIGTERM
- **Zero-DB Mode** — JSON file storage for development (no database needed)

---

## Admin Dashboard

MachineAuth ships with a full admin UI built with **React 18 + TypeScript + Tailwind CSS**.

### Pages

| Page | Description |
|------|-------------|
| **Dashboard** | Real-time metrics, health status, system overview |
| **Agents** | Browse, search, filter agents; view details, rotate credentials |
| **Agent Detail** | Credentials, scopes, usage stats, rotation, deactivation |
| **Organizations** | Multi-tenant org management with teams and agents |
| **Token Tools** | Generate, introspect, and revoke tokens from the UI |
| **Webhooks** | Create, manage, test webhooks; view delivery history |
| **Metrics** | Detailed token issuance, refresh, revocation statistics |

### Running the Dashboard

```bash
cd web
npm install
npm run dev
```

Open `http://localhost:3000` — proxies API calls to the Go backend on port 8081.

**Default credentials:** `admin@example.com` / `changeme`

> ⚠️ Change `ADMIN_EMAIL` and `ADMIN_PASSWORD` env vars before deploying to production.

### Production Build

```bash
cd web
npm run build    # Output in web/dist/
npm run start    # Serve with built-in static server on port 3000
```

---

## API Reference

### OAuth 2.0 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/oauth/token` | Issue access + refresh token |
| `POST` | `/oauth/introspect` | Validate and inspect a token |
| `POST` | `/oauth/revoke` | Revoke an access token |
| `POST` | `/oauth/refresh` | Refresh an access token |
| `GET` | `/.well-known/jwks.json` | Public key set (JWKS) |

### Agent Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/agents` | List all agents |
| `POST` | `/api/agents` | Create a new agent |
| `GET` | `/api/agents/{id}` | Get agent details |
| `DELETE` | `/api/agents/{id}` | Delete an agent |
| `POST` | `/api/agents/{id}/rotate` | Rotate agent credentials |
| `POST` | `/api/agents/{id}/deactivate` | Deactivate an agent |

### Agent Self-Service (JWT Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/agents/me` | Get own profile |
| `GET` | `/api/agents/me/usage` | Get own usage statistics |
| `POST` | `/api/agents/me/rotate` | Rotate own credentials |
| `POST` | `/api/agents/me/deactivate` | Deactivate own account |
| `POST` | `/api/agents/me/reactivate` | Reactivate own account |
| `DELETE` | `/api/agents/me/delete` | Delete own account |

### Webhooks

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/webhooks` | List webhooks |
| `POST` | `/api/webhooks` | Create webhook |
| `GET` | `/api/webhooks/{id}` | Get webhook details |
| `PUT` | `/api/webhooks/{id}` | Update webhook |
| `DELETE` | `/api/webhooks/{id}` | Delete webhook |
| `POST` | `/api/webhooks/{id}/test` | Send test delivery |
| `GET` | `/api/webhooks/{id}/deliveries` | Get delivery history |
| `GET` | `/api/webhook-events` | List available event types |

### Organizations & Teams

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/organizations` | List organizations |
| `POST` | `/api/organizations` | Create organization |
| `GET` | `/api/organizations/{id}` | Get organization |
| `PUT` | `/api/organizations/{id}` | Update organization |
| `DELETE` | `/api/organizations/{id}` | Delete organization |
| `GET` | `/api/organizations/{id}/teams` | List teams |
| `POST` | `/api/organizations/{id}/teams` | Create team |
| `GET` | `/api/organizations/{id}/agents` | List org agents |
| `POST` | `/api/organizations/{id}/agents` | Create org agent |

### API Keys

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/organizations/{id}/api-keys` | List API keys |
| `POST` | `/api/organizations/{id}/api-keys` | Create API key |
| `DELETE` | `/api/organizations/{id}/api-keys/{key_id}` | Revoke API key |

### System

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | Service info + version |
| `GET` | `/health` | Health check |
| `GET` | `/health/ready` | Readiness check (includes agent count) |
| `GET` | `/metrics` | Token/agent/error metrics |
| `POST` | `/api/verify` | Verify JWT and return agent info |

---

## Usage Examples

### Create an Agent

```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read", "write"]}'
```

```json
{
  "agent": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-agent",
    "client_id": "cid_a1b2c3d4e5f6",
    "scopes": ["read", "write"],
    "is_active": true,
    "created_at": "2026-03-01T12:00:00Z"
  },
  "client_secret": "cs_xK9mPqR...",
  "message": "Save this client_secret - it will not be shown again!"
}
```

> ⚠️ The `client_secret` is only returned once. Store it securely.

### Get an Access Token

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=cid_a1b2c3d4e5f6" \
  -d "client_secret=cs_xK9mPqR..."
```

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "rt_8f14e45f..."
}
```

### Introspect a Token

```bash
curl -X POST http://localhost:8080/oauth/introspect \
  -d "token=eyJhbGciOiJSUzI1NiIs..."
```

```json
{
  "active": true,
  "client_id": "cid_a1b2c3d4e5f6",
  "scope": "read write",
  "token_type": "Bearer",
  "exp": 1709308800,
  "iat": 1709305200
}
```

### Refresh a Token

```bash
curl -X POST http://localhost:8080/oauth/refresh \
  -d "refresh_token=rt_8f14e45f..." \
  -d "client_id=cid_a1b2c3d4e5f6" \
  -d "client_secret=cs_xK9mPqR..."
```

### Revoke a Token

```bash
curl -X POST http://localhost:8080/oauth/revoke \
  -d "token=eyJhbGciOiJSUzI1NiIs..."
```

### Rotate Agent Credentials

```bash
curl -X POST http://localhost:8080/api/agents/{agent_id}/rotate
```

Returns a new `client_secret` — the old one is immediately invalidated.

### Agent Self-Service

Agents can manage themselves using their JWT token:

```bash
# View own profile
curl -H "Authorization: Bearer eyJ..." http://localhost:8080/api/agents/me

# Check usage stats
curl -H "Authorization: Bearer eyJ..." http://localhost:8080/api/agents/me/usage

# Rotate own credentials
curl -X POST -H "Authorization: Bearer eyJ..." http://localhost:8080/api/agents/me/rotate

# Deactivate own account
curl -X POST -H "Authorization: Bearer eyJ..." http://localhost:8080/api/agents/me/deactivate

# Delete own account permanently
curl -X DELETE -H "Authorization: Bearer eyJ..." http://localhost:8080/api/agents/me/delete
```

### Webhooks

```bash
# Create a webhook
curl -X POST http://localhost:8080/api/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Prod Notifications",
    "url": "https://example.com/webhooks",
    "events": ["agent.created", "agent.deleted", "token.issued"],
    "max_retries": 5
  }'

# Test it
curl -X POST http://localhost:8080/api/webhooks/{id}/test \
  -H "Content-Type: application/json" \
  -d '{"event": "webhook.test"}'

# Check delivery history
curl http://localhost:8080/api/webhooks/{id}/deliveries
```

### Organizations & API Keys

```bash
# Create an organization
curl -X POST http://localhost:8080/api/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Acme Corp", "slug": "acme", "owner_email": "admin@acme.com"}'

# Create a team
curl -X POST http://localhost:8080/api/organizations/{org_id}/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "Engineering", "description": "Engineering team"}'

# Create an API key
curl -X POST http://localhost:8080/api/organizations/{org_id}/api-keys \
  -H "Content-Type: application/json" \
  -d '{"name": "production-key", "expires_in": 86400}'
```

API keys can be used in place of JWT tokens:

```bash
curl -H "Authorization: Bearer sk_1cF4CG1RE..." http://localhost:8080/api/verify
```

---

## JWT Token Claims

Tokens issued by MachineAuth include rich claims for authorization decisions:

```json
{
  "iss": "https://auth.yourdomain.com",
  "sub": "cid_a1b2c3d4e5f6",
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "org_id": "org-uuid",
  "team_id": "team-uuid",
  "scope": ["read", "write"],
  "jti": "unique-token-id",
  "exp": 1709308800,
  "iat": 1709305200
}
```

Validate tokens using the public key from `/.well-known/jwks.json`.

---

## SDKs

Official client libraries for TypeScript and Python.

### TypeScript

```bash
npm install @machineauth/sdk
```

```typescript
import { MachineAuthClient } from '@machineauth/sdk'

const client = new MachineAuthClient({
  baseUrl: 'https://auth.yourdomain.com',
  clientId: 'cid_a1b2c3d4e5f6',
  clientSecret: 'cs_xK9mPqR...',
})

// Get a token
const token = await client.getToken({ scope: 'read write' })

// List agents
const agents = await client.listAgents()

// Self-service
const me = await client.getMe()
```

### Python

```bash
pip install machineauth
```

```python
from machineauth import MachineAuthClient

client = MachineAuthClient(
    base_url="https://auth.yourdomain.com",
    client_id="cid_a1b2c3d4e5f6",
    client_secret="cs_xK9mPqR...",
)

# Get a token
token = client.get_token(scope="read write")

# Async support
from machineauth import AsyncMachineAuthClient

async_client = AsyncMachineAuthClient(...)
token = await async_client.get_token(scope="read write")
```

See [sdk/README.md](sdk/README.md) for the full SDK documentation.

---

## Configuration

All configuration via environment variables (or `.env` file):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port |
| `ENV` | `development` | Environment (`development` / `production`) |
| `DATABASE_URL` | `json:machineauth.json` | Database connection string |
| `JWT_SIGNING_ALGORITHM` | `RS256` | JWT signing algorithm |
| `JWT_KEY_ID` | `key-1` | JWKS key identifier |
| `JWT_ACCESS_TOKEN_EXPIRY` | `3600` | Access token TTL in seconds (1 hour) |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins (comma-separated) |
| `REQUIRE_HTTPS` | `false` | Enforce HTTPS redirects |
| `ADMIN_EMAIL` | `admin@example.com` | Admin dashboard email |
| `ADMIN_PASSWORD` | `changeme` | Admin dashboard password |
| `WEBHOOK_WORKER_COUNT` | `3` | Concurrent webhook delivery workers |
| `WEBHOOK_MAX_RETRIES` | `10` | Max delivery retry attempts |
| `WEBHOOK_TIMEOUT_SECS` | `10` | Webhook HTTP request timeout |

### Database Options

```bash
# JSON file (default, zero deps, great for dev)
DATABASE_URL=json:machineauth.json

# PostgreSQL (recommended for production)
DATABASE_URL=postgresql://user:pass@localhost:5432/machineauth
```

### Example `.env`

```env
PORT=8080
ENV=production
DATABASE_URL=postgresql://machineauth:secret@db:5432/machineauth
JWT_ACCESS_TOKEN_EXPIRY=1800
ALLOWED_ORIGINS=https://dashboard.yourdomain.com
ADMIN_PASSWORD=your-secure-password
WEBHOOK_WORKER_COUNT=5
```

---

## Deployment

### Docker Compose (Recommended)

```bash
git clone https://github.com/mandarwagh9/MachineAuth.git
cd MachineAuth
docker-compose up -d
```

This starts three services:

| Service | Port | Description |
|---------|------|-------------|
| **postgres** | 5432 | PostgreSQL 15 database |
| **server** | 8080 | Go API server |
| **web** | 80 | React admin dashboard |

### Docker (Server Only)

```dockerfile
# docker/Dockerfile.server — Multi-stage build
FROM golang:1.21-alpine AS builder
# ... builds to /server

FROM alpine:3.19
COPY --from=builder /server .
EXPOSE 8080
CMD ["./server"]
```

```bash
docker build -f docker/Dockerfile.server -t machineauth .
docker run -p 8080:8080 -e DATABASE_URL=json:/data/machineauth.json machineauth
```

### Build from Source

```bash
# Requirements: Go 1.21+
go build -o machineauth ./cmd/server
./machineauth
```

### Systemd (Linux)

```ini
[Unit]
Description=MachineAuth - OAuth 2.0 for AI Agents
After=network.target

[Service]
Type=simple
User=machineauth
WorkingDirectory=/opt/machineauth
ExecStart=/opt/machineauth/machineauth
EnvironmentFile=/opt/machineauth/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable machineauth
sudo systemctl start machineauth
```

---

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌──────────────┐
│  React Admin UI │────▶│   Go Server  │────▶│  PostgreSQL   │
│  (Tailwind CSS) │     │  (net/http)  │     │  or JSON file │
└─────────────────┘     └──────┬───────┘     └──────────────┘
                               │
                    ┌──────────┼──────────┐
                    │          │          │
              ┌─────▼──┐ ┌────▼───┐ ┌────▼────┐
              │ Agents  │ │ Tokens │ │Webhooks │
              │ Service │ │Service │ │ Worker  │
              └────────┘ └────────┘ └─────────┘
```

### Project Structure

```
machineauth/
├── cmd/server/main.go          # Server entry point
├── internal/
│   ├── config/config.go        # Environment configuration
│   ├── db/db.go                # Database layer (Postgres + JSON)
│   ├── handlers/               # HTTP request handlers
│   │   ├── agents.go           # Agent CRUD + self-service
│   │   ├── auth.go             # OAuth2 token endpoints
│   │   └── webhook.go          # Webhook management
│   ├── middleware/              # Logging, CORS
│   ├── models/models.go        # All data types and DTOs
│   ├── services/               # Business logic
│   │   ├── agent.go            # Agent operations
│   │   ├── audit.go            # Audit logging + webhook triggers
│   │   ├── token.go            # JWT creation/validation
│   │   ├── webhook.go          # Webhook CRUD
│   │   └── webhook_worker.go   # Async delivery processing
│   └── utils/crypto.go         # Cryptographic helpers
├── web/                        # React admin dashboard
│   ├── src/
│   │   ├── pages/              # Dashboard, Agents, Tokens, Webhooks, etc.
│   │   ├── components/         # Layout, Sidebar
│   │   ├── services/           # API client (axios)
│   │   └── types/              # TypeScript interfaces
│   └── vite.config.ts          # Vite + proxy config
├── sdk/
│   ├── typescript/             # @machineauth/sdk (npm)
│   └── python/                 # machineauth (pip)
├── docker/                     # Dockerfiles
├── deploy/                     # Deployment scripts
└── docker-compose.yml          # Full-stack compose
```

---

## Security

### Best Practices

- **Use HTTPS** — Always run behind a TLS-terminating reverse proxy in production
- **Rotate credentials** — Use the rotation API regularly; old secrets are invalidated immediately
- **Short token expiry** — Default 1 hour; reduce to 15-30 min for sensitive workloads
- **Restrict CORS** — Set `ALLOWED_ORIGINS` to your specific domains
- **Change admin password** — Default is `changeme`; set `ADMIN_PASSWORD` before deploying
- **Monitor metrics** — Watch `/metrics` for token issuance anomalies
- **Use Postgres in prod** — JSON file storage is for development only

### Reporting Vulnerabilities

Please email security concerns directly rather than opening public issues. See [SECURITY.md](SECURITY.md).

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.21, `net/http`, `golang-jwt/jwt/v5` |
| **Frontend** | React 18, TypeScript 5.3, Vite 5, Tailwind CSS 3.4 |
| **Database** | PostgreSQL 15 (prod) / JSON file (dev) |
| **Auth** | OAuth 2.0 Client Credentials, RS256 JWT |
| **Icons** | Lucide React |
| **Toasts** | Sonner |
| **HTTP Client** | Axios |

---

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Make your changes and add tests
4. Run `go test -v ./...` and `cd web && npm run build`
5. Commit (`git commit -m 'feat: add amazing feature'`)
6. Push (`git push origin feature/amazing`)
7. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

## Links

- **Live Demo:** [https://auth.writesomething.fun](https://auth.writesomething.fun)
- **Issues:** [GitHub Issues](https://github.com/mandarwagh9/MachineAuth/issues)
- **SDK Docs:** [sdk/README.md](sdk/README.md)

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=mandarwagh9/MachineAuth&type=Date)](https://star-history.com/#mandarwagh9/MachineAuth&Date)

---

<p align="center">
  Built for the AI agent ecosystem 🤖
</p>
