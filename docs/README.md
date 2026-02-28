# MachineAuth Documentation

## Getting Started

- [Quick Start](../README.md) - Basic setup and usage
- [API Reference](../README.md#api-reference) - Complete API documentation

## Testing

- [Hallucination Test](HALLUCINATION-TEST.md) - Test that proves AI agents can authenticate and access real data
- [Demo Website](DEMO-WEBSITE.md) - Interactive demo using MachineAuth for authentication

## Architecture

See [AGENTS.md](../AGENTS.md) for development guidelines and code conventions.

## Key Features

### Agent Self-Service API

Agents can manage their own accounts:

```bash
# Get own profile
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8081/api/agents/me

# Get usage stats
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8081/api/agents/me/usage

# Rotate credentials
curl -X POST -H "Authorization: Bearer TOKEN" \
  http://localhost:8081/api/agents/me/rotate

# Deactivate account
curl -X POST -H "Authorization: Bearer TOKEN" \
  http://localhost:8081/api/agents/me/deactivate

# Delete account
curl -X DELETE -H "Authorization: Bearer TOKEN" \
  http://localhost:8081/api/agents/me/delete
```

### Usage Tracking

Full usage metrics per agent:

- Token count (cumulative)
- Refresh count
- Last activity timestamp
- Last token issued timestamp
- Credential rotation history

## Related Projects

- [test-hallucination.js](../test-hallucination.js) - Automated hallucination test
- [demo-website](../demo-website/) - Demo application
