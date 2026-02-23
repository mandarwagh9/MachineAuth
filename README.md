# MachineAuth - AI Agent Authentication Platform

Self-hosted OAuth 2.0 authentication for AI agents and machine-to-machine communication.

## Quick Start

### Prerequisites

- Go 1.21+ (for backend)
- Node.js 18+ (for frontend)
- Docker & Docker Compose (for containerized deployment)

### Local Development

1. **Clone the repository**

2. **Run the backend**
   ```bash
   go run server-main.go
   ```

3. **Access the API**
   - API: http://localhost:8081
   - Health: http://localhost:8081/health
   - JWKS: http://localhost:8081/.well-known/jwks.json

### Docker Deployment

```bash
docker-compose up -d
```

## API Endpoints

### Create Agent
```bash
curl -X POST http://localhost:8081/api/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read:data", "write:data"]}'
```

### Get Token
```bash
curl -X POST http://localhost:8081/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET"
```

### Verify Token
```bash
curl http://localhost:8081/.well-known/jwks.json
```

### Access Protected Endpoint
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8081/api/secret
```

## Features

- OAuth 2.0 Client Credentials flow
- JWT token generation with RS256
- JWKS endpoint for public key distribution
- Agent management (create, list, revoke)
- Scope-based permissions
- Self-hosted deployment
- JSON file-based storage (no database required)

## Tech Stack

- **Backend**: Go
- **Storage**: JSON file (for simplicity)
- **Deployment**: Docker

## Configuration

The server uses the following default paths:
- Database: `/opt/machineauth/machineauth.json`
- Port: `8081`

To change these, modify the `dbFile` variable and `Addr` in `server-main.go`.

## Live Demo

The service is currently deployed at: https://auth.writesomething.fun

## License

MIT
