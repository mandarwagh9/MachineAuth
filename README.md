<h1 align="center">
  <br>
  <img src="https://raw.githubusercontent.com/mandarwagh9/MachineAuth/main/.github/logo.png" alt="MachineAuth" width="200">
  <br>
  MachineAuth
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
  <a href="https://github.com/mandarwagh9/MachineAuth/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/mandarwagh9/MachineAuth.yml?style=flat-square" alt="Build">
  </a>
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-features">Features</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-api-reference">API Reference</a> •
  <a href="#-deployment">Deployment</a> •
  <a href="#-security">Security</a> •
  <a href="#-contributing">Contributing</a>
</p>

---

## 🔐 Why MachineAuth?

As AI agents become autonomous, they need **secure, programmatic authentication** - just like developers use API keys, but built specifically for machines.

### The Problem

Your AI agent needs to access protected APIs, but:
- ❌ Username/password doesn't work for machines
- ❌ API keys are hard to rotate and audit
- ❌ Existing OAuth flows are designed for humans, not agents
- ❌ No way to verify which agent is making requests

### The Solution

MachineAuth implements the **OAuth 2.0 Client Credentials flow** - the industry standard for machine-to-machine (M2M) authentication.

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   AI Agent  │────▶│ MachineAuth  │────▶│  Your API        │
│             │     │   Server     │     │  Service        │
└─────────────┘     └──────────────┘     └─────────────────┘
                           │
                    ┌──────┴──────┐
                    │  JSON File   │
                    │  Storage     │
                    └─────────────┘
```

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🔑 **OAuth 2.0 Client Credentials** | Industry-standard M2M authentication flow |
| 📜 **JWT Token Generation** | RS256 signed tokens with configurable expiry |
| 🔓 **JWKS Endpoint** | Public key distribution for token verification |
| 👥 **Agent Management** | Create, list, and manage agent credentials |
| 🛡️ **Scope-Based Permissions** | Fine-grained access control |
| 📁 **Zero-Database** | JSON file storage - no external dependencies |
| 🚀 **Self-Hosted** | Full control - deploy anywhere |
| ⚡ **Blazing Fast** | Written in Go for minimal latency |
| 🔒 **Secure** | Bcrypt hashed secrets, RS256 tokens |

---

## 🚀 Quick Start

### One-Command Setup

```bash
# Clone and run
git clone https://github.com/mandarwagh9/MachineAuth.git
cd MachineAuth
go run server-main.go
```

Server starts on `http://localhost:8081`

That's it! No database, no dependencies.

---

## 📖 Usage

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
  "agent": {
    "id": "...",
    "name": "my-ai-agent",
    "client_id": "a1b2c3d4-e5f6-...",
    "scopes": ["read:data", "write:data"]
  },
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
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleS0xIn0...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### 3. Access Protected Resources

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8081/api/secret
```

**Response:**
```html
<html>
<head><title>Secret Page</title></head>
<body>
  <h1>MachineAuth Protected Page</h1>
  <p>Success! Your agent authenticated with JWT.</p>
</body>
</html>
```

---

## 📋 API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Service info |
| `/health` | GET | Health check |
| `/oauth/token` | POST | Get JWT access token |
| `/.well-known/jwks.json` | GET | Public keys for verification |
| `/api/agents` | POST | Create new agent |
| `/api/agents` | GET | List all agents |
| `/api/secret` | GET | Protected endpoint (demo) |
| `/api/verify` | GET | Verify token & return claims |

---

## 💻 Integration Examples

### Python

```python
import requests

class MachineAuth:
    def __init__(self, base_url: str, client_id: str, client_secret: str):
        self.base_url = base_url
        self.client_id = client_id
        self.client_secret = client_secret
        self.token = None
    
    def get_token(self) -> str:
        """Get JWT access token."""
        response = requests.post(
            f"{self.base_url}/oauth/token",
            data={
                "grant_type": "client_credentials",
                "client_id": self.client_id,
                "client_secret": self.client_secret
            }
        )
        response.raise_for_status()
        self.token = response.json()["access_token"]
        return self.token
    
    def call_api(self, endpoint: str) -> dict:
        """Call protected API endpoint."""
        if not self.token:
            self.get_token()
        
        response = requests.get(
            f"{self.base_url}{endpoint}",
            headers={"Authorization": f"Bearer {self.token}"}
        )
        return response.json()


# Usage
auth = MachineAuth(
    base_url="https://auth.writesomething.fun",
    client_id="YOUR_CLIENT_ID",
    client_secret="YOUR_CLIENT_SECRET"
)

token = auth.get_token()
result = auth.call_api("/api/verify")
print(result)
```

### Node.js

```javascript
class MachineAuth {
  constructor(baseUrl, clientId, clientSecret) {
    this.baseUrl = baseUrl;
    this.clientId = clientId;
    this.clientSecret = clientSecret;
    this.token = null;
  }

  async getToken() {
    const response = await fetch(`${this.baseUrl}/oauth/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'client_credentials',
        client_id: this.clientId,
        client_secret: this.clientSecret
      })
    });
    const data = await response.json();
    this.token = data.access_token;
    return this.token;
  }

  async callApi(endpoint) {
    if (!this.token) await this.getToken();
    
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    return response.json();
  }
}

// Usage
const auth = new MachineAuth(
  'https://auth.writesomething.fun',
  'YOUR_CLIENT_ID',
  'YOUR_CLIENT_SECRET'
);

const token = await auth.getToken();
const result = await auth.callApi('/api/verify');
console.log(result);
```

### Go

```go
package main

import (
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "io/ioutil"
)

func main() {
    data := url.Values{}
    data.Set("grant_type", "client_credentials")
    data.Set("client_id", "YOUR_CLIENT_ID")
    data.Set("client_secret", "YOUR_CLIENT_SECRET")

    resp, _ := http.Post(
        "https://auth.writesomething.fun/oauth/token",
        "application/x-www-form-urlencoded",
        strings.NewReader(data.Encode()),
    )
    defer resp.Body.Close()
    
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

---

## 🐳 Deployment

### Docker

```yaml
# docker-compose.yml
version: '3.8'

services:
  machineauth:
    build: .
    ports:
      - "8081:8081"
    volumes:
      - ./data:/opt/machineauth
    restart: unless-stopped
    environment:
      - PORT=8081
```

```bash
docker-compose up -d
```

### Systemd (Linux)

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
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable machineauth
sudo systemctl start machineauth
```

### Build Binary

```bash
# Build for current platform
go build -o machineauth server-main.go

# Cross-compile for Linux AMD64
GOOS=linux GOARCH=amd64 go build -o machineauth server-main.go
```

---

## 🔒 Security

| Consideration | Status |
|---------------|--------|
| Client secrets hashed with bcrypt | ✅ |
| JWT signed with RS256 | ✅ |
| JWKS endpoint for verification | ✅ |
| Configurable token expiry | ✅ (default: 1 hour) |
| HTTPS recommended for production | ⚠️ |

### Production Checklist

- [ ] Use HTTPS/TLS termination
- [ ] Rotate credentials regularly
- [ ] Use strong token expiry policies
- [ ] Monitor token usage
- [ ] Keep Go version updated

---

## 🗺️ Roadmap

- [ ] Refresh token support
- [ ] Token revocation
- [ ] Agent groups/roles
- [ ] Rate limiting
- [ ] Audit logging
- [ ] PostgreSQL storage option
- [ ] Admin UI

---

## 🤝 Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) first.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 🧪 Live Demo

Try it out right now:

```bash
# Create agent
curl -X POST https://auth.writesomething.fun/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "demo-agent", "scopes": ["read"]}'

# Get token (use credentials from response)

# Verify
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://auth.writesomething.fun/api/verify
```

**Live URL:** [https://auth.writesomething.fun](https://auth.writesomething.fun)

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  Built with ❤️ for the AI agent ecosystem
</p>
