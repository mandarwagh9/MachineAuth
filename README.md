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
  <a href="https://goreportcard.com/report/github.com/mandarwagh9/MachineAuth">
    <img src="https://img.shields.io/goreportcard/g/mandarwagh9/MachineAuth?style=flat-square" alt="Go Report">
  </a>
  <img src="https://img.shields.io/github/v/release/mandarwagh9/MachineAuth?style=flat-square" alt="Version">
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
go run server-main.go
```

Server runs on `http://localhost:8081`

That's it! No database, no dependencies.

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
| 📊 Metrics | Track tokens issued, revoked, etc. |
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
| `/api/verify` | GET | Verify token | `curl -H "Authorization: Bearer ..." localhost:8081/api/verify` |
| `/metrics` | GET | Server metrics | `curl localhost:8081/metrics` |

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
- [Releases](https://github.com/mandarwagh9/MachineAuth/releases)

---

<p align="center">
  Built for the AI agent ecosystem 🤖
</p>
