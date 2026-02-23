<img src="https://img.shields.io/badge/MachineAuth-OAuth2.0%20for%20AI%20Agents-blue?style=for-the-badge" alt="MachineAuth">

# MachineAuth

<p align="center">
  <strong>Secure OAuth 2.0 authentication for AI agents and machine-to-machine communication.</strong>
</p>

<p align="center">
  <a href="https://github.com/mandarwagh9/MachineAuth/stargazers"><img src="https://img.shields.io/github/stars/mandarwagh9/MachineAuth?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/mandarwagh9/MachineAuth/issues"><img src="https://img.shields.io/github/issues/mandarwagh9/MachineAuth?style=flat-square" alt="Issues"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/mandarwagh9/MachineAuth?style=flat-square" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/mandarwagh9/MachineAuth"><img src="https://goreportcard.com/badge/github.com/mandarwagh9/MachineAuth?style=flat-square" alt="Go Report"></a>
</p>

---

## Why MachineAuth?

As AI agents become autonomous, they need **secure, programmatic authentication** - just like developers use API keys, but built for machines.

MachineAuth provides a self-hosted OAuth 2.0 server specifically designed for:
- 🤖 **AI Agents** - Authenticate and authorize autonomous agents
- 🔄 **M2M Communication** - Machine-to-machine authentication
- 🔐 **API Security** - Issue and validate JWT tokens for your services

### The Problem

Your AI agent needs to access protected APIs, but:
- Username/password doesn't work for machines
- API keys are hard to rotate and audit
- Existing OAuth flows are designed for humans

### The Solution

MachineAuth implements the **OAuth 2.0 Client Credentials flow** - the standard for machine-to-machine authentication.

---

## Features

| Feature | Description |
|---------|-------------|
| 🔑 **OAuth 2.0 Client Credentials** | Industry-standard M2M authentication |
| 📜 **JWT Token Generation** | RS256 signed tokens with configurable expiry |
| 🔓 **JWKS Endpoint** | Public key distribution for token verification |
| 👥 **Agent Management** | Create, list, and manage agent credentials |
| 🛡️ **Scope-Based Permissions** | Fine-grained access control |
| 📁 **Zero-Database** | JSON file storage - no external dependencies |
| 🚀 **Self-Hosted** | Full control - deploy anywhere |
| ⚡ **Blazing Fast** | Written in Go for minimal latency |

---

## Quick Start

### Run Locally

```bash
# Clone the repository
git clone https://github.com/mandarwagh9/MachineAuth.git
cd MachineAuth

# Run the server
go run server-main.go
```

Server starts on `http://localhost:8081`

---

## Usage

### 1. Create an Agent

```bash
curl -X POST http://localhost:8081/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-ai-agent",
    "scopes": ["read:data", "write:data"]
  }'
```

**Response:**
```json
{
  "agent": {...},
  "client_id": "a1b2c3d4-e5f6-...",
  "client_secret": "x9y8z7w6-v5u4t3-..."
}
```

> ⚠️ **Important:** Save `client_id` and `client_secret` - the secret is only shown once!

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
  "expires_in": 3600
}
```

### 3. Access Protected Resources

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8081/api/secret
```

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Service info |
| `/health` | GET | Health check |
| `/oauth/token` | POST | Get JWT access token |
| `/.well-known/jwks.json` | GET | Public keys for verification |
| `/api/agents` | POST | Create new agent |
| `/api/agents` | GET | List all agents |
| `/api/secret` | GET | Protected endpoint demo |
| `/api/verify` | GET | Verify token & return claims |

---

## Production Deployment

### Systemd Service (Linux)

```ini
# /etc/systemd/system/machineauth.service
[Unit]
Description=MachineAuth - AI Agent Authentication
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

### Docker

```yaml
# docker-compose.yml
services:
  machineauth:
    build: .
    ports:
      - "8081:8081"
    volumes:
      - ./data:/opt/machineauth
    restart: unless-stopped
```

---

## Integrating with Your AI Agent

### Python Example

```python
import requests

# Get token
def get_token(client_id, client_secret):
    response = requests.post(
        "https://your-auth-server.com/oauth/token",
        data={
            "grant_type": "client_credentials",
            "client_id": client_id,
            "client_secret": client_secret
        }
    )
    return response.json()["access_token"]

# Use token
def call_protected_api(token):
    response = requests.get(
        "https://your-api.com/protected",
        headers={"Authorization": f"Bearer {token}"}
    )
    return response.json()
```

### Node.js Example

```javascript
// Get token
async function getToken(clientId, clientSecret) {
  const response = await fetch('https://your-auth-server.com/oauth/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'client_credentials',
      client_id: clientId,
      client_secret: clientSecret
    })
  });
  const data = await response.json();
  return data.access_token;
}
```

---

## Security Considerations

- ✅ Client secrets are hashed with bcrypt
- ✅ JWT tokens signed with RS256 (asymmetric)
- ✅ JWKS endpoint for public key distribution
- ✅ Token expiry configurable (default: 1 hour)
- ⚠️ **Use HTTPS in production**
- ⚠️ **Rotate credentials regularly**

---

## Live Demo

Try it out now:

```bash
# Create agent
curl -X POST https://auth.writesomething.fun/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "test-agent", "scopes": ["read"]}'

# Get token (use credentials from above)

# Verify token
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://auth.writesomething.fun/api/verify
```

**Live URL:** https://auth.writesomething.fun

---

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   AI Agent  │────▶│ MachineAuth  │────▶│  Your API       │
│             │     │   Server     │     │  Service        │
└─────────────┘     └──────────────┘     └─────────────────┘
                           │
                    ┌──────┴──────┐
                    │  JSON File  │
                    │  Storage    │
                    └─────────────┘
```

---

## Tech Stack

- **Language:** Go 1.21+
- **Auth:** OAuth 2.0 + JWT (RS256)
- **Storage:** JSON file (no database required)
- **Deployment:** Docker, systemd

---

## Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) first.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  Built with ❤️ for the AI agent ecosystem
</p>
