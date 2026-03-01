# MachineAuth SDKs

Official client SDKs for MachineAuth covering the full API surface: OAuth2 tokens, agent management, self-service, organizations/teams, API keys, and webhooks.

## TypeScript

```bash
cd sdk/typescript
npm install
npm run build
```

### Usage

```ts
import { MachineAuthClient } from '@machineauth/sdk';

const client = new MachineAuthClient({
  baseUrl: 'http://localhost:8081',
  clientId: 'your-client-id',
  clientSecret: 'your-client-secret',
});

// Get token (auto-stored for subsequent calls)
const token = await client.clientCredentialsToken();

// Agent management
const agents = await client.listAgents();
const created = await client.createAgent({ name: 'my-agent', scopes: ['read'] });
await client.rotateAgent(created.agent.id);

// Self-service (agent calling as itself)
const me = await client.getMe();
const usage = await client.getMyUsage();

// Organizations & teams
const orgs = await client.listOrganizations();
const teams = await client.listTeams(orgs[0].id);

// API keys
const keys = await client.listAPIKeys(orgs[0].id);

// Webhooks
const hooks = await client.listWebhooks();
const events = await client.listWebhookEvents();
```

### Tests

```bash
npm install vitest --save-dev
npx vitest run src/index.test.ts
```

## Python

```bash
cd sdk/python
pip install -e .
```

### Usage

```python
from machineauth import MachineAuthClient

client = MachineAuthClient(
    base_url="http://localhost:8081",
    client_id="your-client-id",
    client_secret="your-client-secret",
)

# Get token (auto-stored for subsequent calls)
token = client.client_credentials_token()

# Agent management
agents = client.list_agents()
created = client.create_agent("my-agent", scopes=["read"])
client.rotate_agent(created["agent"]["id"])

# Self-service
me = client.get_me()
usage = client.get_my_usage()

# Organizations & teams
orgs = client.list_organizations()
teams = client.list_teams(orgs[0].id)

# API keys
keys = client.list_api_keys(orgs[0].id)

# Webhooks
hooks = client.list_webhooks()
events = client.list_webhook_events()
```

### Tests

```bash
pip install pytest
pytest test_client.py -v
```

## API Coverage

Both SDKs cover the complete MachineAuth API:

| Category | Methods |
|----------|---------|
| **Tokens** | clientCredentials, refresh, introspect, revoke |
| **Agents** | list, create, get, delete, rotate |
| **Self-service** | getMe, getMyUsage, rotateMe, deactivateMe, reactivateMe, deleteMe |
| **Organizations** | list, create, get, update, delete |
| **Teams** | list, create |
| **API Keys** | list, create, delete |
| **Webhooks** | list, create, get, update, delete, test, listDeliveries, getDelivery, listEvents |
