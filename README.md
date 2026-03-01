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

3. **Run the frontend**
   ```bash
   cd web && npm install && npm run dev
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
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

### Backend
- OAuth 2.0 Client Credentials flow
- JWT token generation with RS256
- JWKS endpoint for public key distribution
- Agent management (create, list, revoke)
- Credential rotation for agents
- Scope-based permissions
- Audit logging for agent lifecycle events
- Self-hosted deployment
- JSON file-based storage (no database required)

### Frontend (Web Dashboard)
- **Agent Management Dashboard** — View all registered agents in a sortable table with name, client ID, scopes, status, and creation date
- **Agent Creation** — Form to register new agents with name, comma-separated scopes, and optional expiry duration. Displays one-time client credentials on success
- **Agent Detail View** — Inspect individual agent details including ID, client ID, status, scopes, and timestamps
- **Credential Rotation** — Rotate an agent's client secret directly from the detail page with one-time display of the new secret
- **Agent Deletion** — Delete agents with confirmation dialog from the list or detail view
- **Token Generator** — Interactive tool to request OAuth 2.0 access tokens by selecting an agent, entering the client secret, and optionally specifying scopes. Displays the JWT with a copy-to-clipboard button
- **Responsive Layout** — Header navigation with links to Agents and Tokens sections
- **API Proxy** — Vite dev server proxies `/api`, `/oauth`, `/.well-known`, and `/health` to the Go backend

## Tech Stack

- **Backend**: Go 1.21+
- **Frontend**: React 18, TypeScript, React Router 6, Axios, Vite 5
- **Storage**: JSON file (for simplicity)
- **Deployment**: Docker + Nginx reverse proxy

## Configuration

The server uses the following default paths:
- Database: `/opt/machineauth/machineauth.json`
- Port: `8081`

To change these, modify the `dbFile` variable and `Addr` in `server-main.go`.

## Live Demo

The service is currently deployed at: https://auth.writesomething.fun

## License

MIT
