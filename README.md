<h1 align="center">
  <br>
  MachineAuth
  <br>
</h1>

<p align="center">
  <strong>Secure OAuth 2.0 authentication for AI agents and machine-to-machine communication.</strong>
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
</p>

---

## Table of Contents

- [About](#about)
- [Demo](#demo)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

---

## About

MachineAuth is a self-hosted OAuth 2.0 server for authenticating **AI agents** and machines. 

What is an AI agent in this context? A software bot (like OpenCLAW, Claude Code, etc.) that makes API calls to access protected resources. Instead of sharing long-lived API keys, your agents can authenticate using OAuth 2.0 Client Credentials and receive short-lived JWT tokens.

**Why?**
- 🔑 No more sharing API keys
- ⏱️ Short-lived tokens (configurable)
- 🔄 Easy credential rotation
- 🛡️ Industry-standard security

---

## Demo

```bash
# Create an agent
$ curl -X POST http://localhost:8081/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read"]}'
  
# Response: {"client_id": "abc-123", "client_secret": "secret-xyz", ...}

# Get token
$ curl -X POST http://localhost:8081/oauth/token \
  -d "grant_type=client_credentials" \
  -d "client_id=abc-123" \
  -d "client_secret=secret-xyz"

# Response: {"access_token": "eyJ...", "expires_in": 3600, "refresh_token": "..."}

# Access protected resource
$ curl -H "Authorization: Bearer eyJ..." http://localhost:8081/api/verify

# Response: {"secret_code": "AGENT-AUTH-2026-XK9M", "message": "Your agent is NOT hallucinating!"}
```

**Live Demo:** https://auth.writesomething.fun

---

## Quick Start

```bash
# Clone and run
git clone https://github.com/mandarwagh9/MachineAuth.git
cd MachineAuth
go run ./cmd/server
```

Server runs on `http://localhost:8080`

That's it! No database, no dependencies. Uses JSON file storage by default.

---

## Admin UI

MachineAuth includes a beautiful admin dashboard built with React + TailwindCSS.

### Quick Start

```bash
cd web
npm install
npm run dev
```

Access the admin UI at `http://localhost:3000`

### Login Credentials

- **Username:** `admin`
- **Password:** `admin`

> ⚠️ Change the password in `web/src/App.tsx` before deploying!

### Features

- 📊 **Dashboard** - Real-time metrics and statistics
- 👥 **Agents** - Create, view, rotate, deactivate agents
- 📈 **Metrics** - Token issuance, refresh, revocation stats
- 🔐 **Secure** - Basic authentication required

### Building for Production

```bash
cd web
npm run build
```

The built files will be in `web/dist/`

---

## Installation

### Requirements

- **Go 1.21+** - [Install Go](https://go.dev/doc/install)

### Build from Source

```bash
# Clone
git clone https://github.com/mandarwagh9/MachineAuth.git
cd MachineAuth

# Build
go build -o machineauth server-main.go

# Run
./machineauth
```

### Pre-built Binaries

Download from [Releases](https://github.com/mandarwagh9/MachineAuth/releases) (coming soon).

---

## Features

| Feature | Description |
|---------|-------------|
| 🔑 OAuth 2.0 Client Credentials | Industry-standard M2M authentication |
| 📜 JWT Tokens | RS256 signed, configurable expiry |
| 🔄 Refresh Tokens | Get new access without re-authenticating |
| 🔍 Token Introspection | Validate tokens via `/oauth/introspect` |
| 🚫 Token Revocation | Invalidate tokens via `/oauth/revoke` |
| 🔐 Agent Rotation | Rotate credentials via API |
| 🤖 Agent Self-Service | Agents manage own accounts via API |
| 📊 Usage Tracking | Track tokens, refreshes, activity per agent |
| 📈 Per-Agent Metrics | Agents can view their own usage statistics |
| 📊 Metrics | Track tokens issued, revoked, etc. |
| 🏢 Multi-Tenant | Organizations and Teams support |
| 🌐 CORS | Configurable cross-origin settings |
| 📁 Zero-DB | JSON file storage - no external dependencies |

---

## Usage

### 1. Create an Agent

```bash
curl -X POST http://localhost:8081/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read", "write"]}'
```

**Response:**
```json
{
  "client_id": "abc-123",
  "client_secret": "xyz-789",
  "message": "Save this client_secret - it will not be shown again!"
}
```

> ⚠️ Save the `client_id` and `client_secret` - the secret is only shown once!

### 2. Get Access Token

```bash
curl -X POST http://localhost:8081/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "uuid-refresh-token"
}
```

### 3. Use the Token

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/verify
```

### 4. Introspect Token

```bash
curl -X POST http://localhost:8081/oauth/introspect \
  -d "token=YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "active": true,
  "scope": "read write",
  "client_id": "abc-123",
  "exp": 1234567890,
  "iat": 1234567890,
  "token_type": "Bearer"
}
```

### 5. Refresh Token

```bash
curl -X POST http://localhost:8081/oauth/refresh \
  -d "refresh_token=YOUR_REFRESH_TOKEN" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

### 6. Revoke Token

```bash
curl -X POST http://localhost:8081/oauth/revoke \
  -d "token=YOUR_ACCESS_TOKEN"
```

---

## API Keys

MachineAuth supports API keys for simpler machine-to-machine authentication.

### Create an API Key

```bash
curl -X POST http://localhost:8081/api/organizations/{org_id}/api-keys \
  -H "Content-Type: application/json" \
  -d '{"name": "production-key", "expires_in": 86400}'
```

**Response:**
```json
{
  "api_key": {
    "id": "uuid",
    "organization_id": "org-uuid",
    "name": "production-key",
    "prefix": "sk_1cF4CG1RE",
    "is_active": true,
    "created_at": "2026-03-01T12:00:00Z"
  },
  "key": "sk_1cF4CG1REhHtXrrxWGcI5jlTX92zeWRyw7goSmopwAs"
}
```

> ⚠️ Save the `key` - it will not be shown again!

### List API Keys

```bash
curl http://localhost:8081/api/organizations/{org_id}/api-keys
```

### Revoke API Key

```bash
curl -X DELETE http://localhost:8081/api/organizations/{org_id}/api-keys/{key_id}
```

### Using API Keys

Use the API key in the `Authorization` header:

```bash
curl -H "Authorization: Bearer sk_1cF4CG1REhHtXrrxWGcI5jlTX92zeWRyw7goSmopwAs" \
  http://localhost:8081/api/verify
```

### API Key Model

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `organization_id` | string | Organization UUID |
| `team_id` | string (optional) | Team UUID |
| `name` | string | Key name |
| `prefix` | string | First 12 chars (for identification) |
| `is_active` | boolean | Whether key is active |
| `expires_at` | timestamp (optional) | Expiration time |
| `last_used_at` | timestamp (optional) | Last usage time |
| `created_at` | timestamp | Creation time |

---

## Multi-Tenant (Organizations & Teams)

### Create an Organization

```bash
curl -X POST http://localhost:8081/api/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Acme Corp", "slug": "acme", "owner_email": "admin@acme.com"}'
```

### Create a Team

```bash
curl -X POST http://localhost:8081/api/organizations/{org_id}/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "Engineering", "description": "Engineering team"}'
```

### Create Agent in Organization

```bash
curl -X POST http://localhost:8081/api/organizations/{org_id}/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read", "write"]}'
```

### List Agents in Organization

```bash
curl http://localhost:8081/api/organizations/{org_id}/agents
```

### JWT Token Claims

Tokens include `org_id` and `team_id` for multi-tenant access control:

```json
{
  "iss": "https://auth.example.com",
  "sub": "client_id",
  "agent_id": "agent-uuid",
  "org_id": "organization-uuid",
  "team_id": "team-uuid (optional)",
  "scope": ["read", "write"],
  "jti": "token-uuid",
  "exp": 1234567890,
  "iat": 1234567890
}
```

---

## API Reference

| Endpoint | Method | Description | Example |
|----------|--------|-------------|---------|
| `/` | GET | Service info | `curl localhost:8081/` |
| `/health` | GET | Health check | `curl localhost:8081/health` |
| `/oauth/token` | POST | Get access token | [See above](#2-get-access-token) |
| `/oauth/introspect` | POST | Validate token | `curl -d "token=..." localhost:8081/oauth/introspect` |
| `/oauth/revoke` | POST | Revoke token | `curl -d "token=..." localhost:8081/oauth/revoke` |
| `/oauth/refresh` | POST | Refresh access token | `curl -d "refresh_token=..." localhost:8081/oauth/refresh` |
| `/.well-known/jwks.json` | GET | Public keys | `curl localhost:8081/.well-known/jwks.json` |
| `/api/agents` | GET | List agents | `curl localhost:8081/api/agents` |
| `/api/agents` | POST | Create agent | [See above](#1-create-an-agent) |
| `/api/agents/{id}` | GET | Agent details | `curl localhost:8081/api/agents/{id}` |
| `/api/agents/{id}/rotate` | POST | Rotate credentials | `curl -X POST localhost:8081/api/agents/{id}/rotate` |
| `/api/agents/{id}/deactivate` | POST | Deactivate agent | `curl -X POST localhost:8081/api/agents/{id}/deactivate` |
| `/api/agents/me` | GET | Get own profile | `curl -H "Authorization: Bearer ..." localhost:8081/api/agents/me` |
| `/api/agents/me/usage` | GET | Get own usage stats | `curl -H "Authorization: Bearer ..." localhost:8081/api/agents/me/usage` |
| `/api/agents/me/rotate` | POST | Rotate own credentials | `curl -X POST -H "Authorization: Bearer ..." localhost:8081/api/agents/me/rotate` |
| `/api/agents/me/deactivate` | POST | Deactivate own account | `curl -X POST -H "Authorization: Bearer ..." localhost:8081/api/agents/me/deactivate` |
| `/api/agents/me/reactivate` | POST | Reactivate own account | `curl -X POST -H "Authorization: Bearer ..." localhost:8081/api/agents/me/reactivate` |
| `/api/agents/me/delete` | DELETE | Delete own account | `curl -X DELETE -H "Authorization: Bearer ..." localhost:8081/api/agents/me/delete` |
| `/api/verify` | GET | Verify token | `curl -H "Authorization: Bearer ..." localhost:8081/api/verify` |
| `/metrics` | GET | Server metrics | `curl localhost:8081/metrics` |

### Organizations API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/organizations` | GET | List organizations |
| `/api/organizations` | POST | Create organization |
| `/api/organizations/{id}` | GET | Get organization |
| `/api/organizations/{id}` | PUT | Update organization |
| `/api/organizations/{id}` | DELETE | Delete organization |
| `/api/organizations/{id}/agents` | GET | List agents in organization |
| `/api/organizations/{id}/agents` | POST | Create agent in organization |

### Teams API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/organizations/{id}/teams` | GET | List teams |
| `/api/organizations/{id}/teams` | POST | Create team |

### API Keys API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/organizations/{id}/api-keys` | GET | List API keys |
| `/api/organizations/{id}/api-keys` | POST | Create API key |
| `/api/organizations/{id}/api-keys/{key_id}` | DELETE | Revoke API key |

### Agent Self-Service API

Agents can manage their own accounts using JWT authentication:

```bash
# Get own profile
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/agents/me

# Get usage statistics (tokens issued, refreshes, activity)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/agents/me/usage

# Rotate credentials
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/agents/me/rotate

# Deactivate account
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/agents/me/deactivate

# Delete account
curl -X DELETE -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/agents/me/delete
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request |
| 401 | Unauthorized (invalid credentials/token) |
| 404 | Not found |
| 405 | Method not allowed |

---

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8081 | Server port |
| `ISSUER` | https://auth.writesomething.fun | Token issuer URL |
| `ACCESS_TOKEN_EXPIRY` | 3600 | Access token TTL (seconds) |
| `REFRESH_TOKEN_EXPIRY` | 604800 | Refresh token TTL (7 days) |
| `CORS_ORIGINS` | * | Allowed origins (comma-separated) |
| `ENABLE_METRICS` | true | Enable `/metrics` endpoint |

Example:
```bash
export PORT=8081
export ISSUER=https://auth.yourdomain.com
export ACCESS_TOKEN_EXPIRY=1800
./machineauth
```

---

## Deployment

### Docker

```yaml
version: '3.8'

services:
  machineauth:
    image: machineauth
    ports:
      - "8081:8081"
    volumes:
      - ./data:/opt/machineauth
    environment:
      - ISSUER=https://auth.yourdomain.com
      - ACCESS_TOKEN_EXPIRY=3600
      - REFRESH_TOKEN_EXPIRY=604800
    restart: unless-stopped
```

```bash
docker-compose up -d
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
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable machineauth
sudo systemctl start machineauth
```

---

## Security

### Best Practices

- ✅ Use HTTPS in production (reverse proxy)
- ✅ Rotate credentials regularly
- ✅ Set appropriate token expiry
- ✅ Restrict CORS origins
- ✅ Monitor `/metrics` endpoint

### Reporting Vulnerabilities

Please email security concerns directly instead of opening public issues.

See [SECURITY.md](SECURITY.md) for details.

---

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

1. Fork the repo
2. Create a feature branch
3. Make changes and test
4. Submit a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE).

---

## Links

- [Live Demo](https://auth.writesomething.fun)
- [Report Issues](https://github.com/mandarwagh9/MachineAuth/issues)

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=mandarwagh9/MachineAuth&type=date&legend=top-left)](https://www.star-history.com/#mandarwagh9/MachineAuth&type=date&legend=top-left)

---

<p align="center">
  Built for the AI agent ecosystem 🤖
</p>
