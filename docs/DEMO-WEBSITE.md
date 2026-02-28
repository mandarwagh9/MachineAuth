# Demo Website

A simple demo application that demonstrates MachineAuth authentication in action.

## What It Demonstrates

- Users can login using MachineAuth credentials
- Protected content is only accessible with valid tokens
- AI agents can authenticate and access protected resources

## Quick Start

```bash
# Navigate to demo directory
cd demo-website

# Install dependencies
npm install

# Start the server
npm start
```

The website will be available at `http://localhost:3001`

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 3001 | Server port |
| `MACHINEAUTH_URL` | https://auth.writesomething.fun | MachineAuth server URL |

Example:
```bash
MACHINEAUTH_URL=http://localhost:8081 npm start
```

## How It Works

### Login Flow

1. User enters their MachineAuth `client_id` and `client_secret`
2. Website calls `/oauth/token` to get an access token
3. Website calls `/api/agents/me` to get agent profile
4. Website calls `/api/verify` to access protected content
5. If successful, displays the secret code

### Protected API

The demo also exposes a protected API endpoint:

```bash
# Get protected content
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:3001/api/protected
```

### Programmatic Agent Creation

```bash
# Create an agent via the demo
curl -X POST http://localhost:3001/api/demo/create-agent \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent", "scopes": ["read", "write"]}'
```

## Testing with AI Agents

This demo is designed to test AI agent hallucination:

1. **Create an agent** via the API or demo endpoint
2. **Login** with the agent's credentials
3. **Access protected content** - the secret code `AGENT-AUTH-2026-XK9M`
4. **Verify** the agent got the exact same secret

If the agent retrieves the exact secret code, it's NOT hallucinating!

## Files

- `server.js` - Express.js server with authentication
- `package.json` - Dependencies
