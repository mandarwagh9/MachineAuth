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

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-features">Features</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-configuration">Configuration</a> •
  <a href="#-contributing">Contributing</a>
</p>

---

## What is MachineAuth?

Self-hosted OAuth 2.0 server for authenticating AI agents and machines. Give your AI agents secure API access without sharing long-lived API keys.

**One sentence:** Secure authentication for AI agents using OAuth 2.0 Client Credentials flow.

---

## 🚀 Quick Start

```bash
# Clone and run
git clone https://github.com/mandarwagh9/MachineAuth.git
cd MachineAuth
go run server-main.go
```

Server runs on `http://localhost:8081`

---

## 📖 Usage

### 1. Create an Agent

```bash
curl -X POST http://localhost:8081/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read", "write"]}'
```

Save the `client_id` and `client_secret` - secret is only shown once!

### 2. Get Access Token

```bash
curl -X POST http://localhost:8081/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

Returns:
```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "uuid..."
}
```

### 3. Use the Token

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/verify
```

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| OAuth 2.0 Client Credentials | Industry-standard M2M auth |
| JWT Tokens | RS256 signed, configurable expiry |
| Refresh Tokens | Get new access tokens without re-auth |
| Token Introspection | Validate tokens via `/oauth/introspect` |
| Token Revocation | Invalidate tokens via `/oauth/revoke` |
| Agent Rotation | Rotate credentials via `/api/agents/{id}/rotate` |
| Metrics | Track tokens issued, revoked, etc. |
| CORS | Configurable cross-origin settings |
| Zero-DB | JSON file storage - no dependencies |

---

## ⚙️ Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8081 | Server port |
| `ISSUER` | https://auth.writesomething.fun | Token issuer |
| `ACCESS_TOKEN_EXPIRY` | 3600 | Access token TTL (seconds) |
| `REFRESH_TOKEN_EXPIRY` | 604800 | Refresh token TTL (7 days) |
| `CORS_ORIGINS` | * | Allowed origins |
| `ENABLE_METRICS` | true | Enable /metrics endpoint |

---

## 📡 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Service info |
| `/health` | GET | Health check |
| `/oauth/token` | POST | Get access token |
| `/oauth/introspect` | POST | Validate token |
| `/oauth/revoke` | POST | Revoke token |
| `/oauth/refresh` | POST | Refresh access token |
| `/.well-known/jwks.json` | GET | Public keys |
| `/api/agents` | GET/POST | List/create agents |
| `/api/agents/{id}` | GET | Agent details |
| `/api/agents/{id}/rotate` | POST | Rotate credentials |
| `/api/agents/{id}/deactivate` | POST | Deactivate agent |
| `/metrics` | GET | Server metrics |

---

## 🐳 Docker

```yaml
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
```

---

## 🤝 Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

1. Fork the repo
2. Create a feature branch
3. Make changes and test
4. Submit a Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 🔗 Links

- [Live Demo](https://auth.writesomething.fun)
- [Report Issues](https://github.com/mandarwagh9/MachineAuth/issues)

---

<p align="center">
  Built for the AI agent ecosystem
</p>
