# Hallucination Test

This test proves that an AI agent can authenticate with MachineAuth and retrieve real data - not hallucinations.

## What It Tests

1. **Agent Signup** - Create an agent account via API
2. **Authentication** - Get a JWT token using client credentials
3. **Protected Access** - Access a protected endpoint with the token
4. **Verification** - Verify the returned secret matches the expected value

## Running the Test

### Prerequisites

- Node.js 18+ installed
- Access to a MachineAuth server (default: `https://auth.writesomething.fun`)

### Run the Test

```bash
# Using Node.js
node test-hallucination.js

# Or using bash
bash test-hallucination.sh
```

### Custom Server

```bash
MACHINEAUTH_URL=http://localhost:8081 node test-hallucination.js
```

## Expected Output

```
========================================
🤖 MachineAuth Hallucination Test
========================================

Step 1: Agent signing up...
✅ Agent created successfully

Step 2: Agent logging in...
✅ Token obtained successfully

Step 3: Agent accessing protected content...
✅ Protected content accessed

Step 4: Verifying secret is REAL...
✅ SUCCESS! Secret matches expected value
   The agent is NOT hallucinating!

🎉 FULL TEST PASSED
```

## The Secret Code

The test verifies the agent retrieves the exact secret:
```
AGENT-AUTH-2026-XK9M
```

If an AI agent returns this exact code, it proves:
- ✅ The agent can make HTTP requests
- ✅ The agent can parse JSON responses
- ✅ The agent can handle OAuth authentication
- ✅ The agent is NOT hallucinating the response

## Why This Matters

In the "Claude Code age" where AI agents are customers, proving an agent can:
1. Sign up programmatically
2. Authenticate via API
3. Access real protected resources

...is fundamental to autonomous agent workflows.

## API Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/agents` | POST | Create agent account |
| `/oauth/token` | POST | Get access token |
| `/api/verify` | GET | Access protected content |

## Test Files

- `test-hallucination.js` - Node.js version (recommended)
- `test-hallucination.sh` - Bash version
