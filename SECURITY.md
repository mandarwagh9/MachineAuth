# Security Policy

## Reporting Vulnerabilities

If you discover a security vulnerability, please **do not open a public GitHub issue**. 

Instead, report it responsibly:

1. **Email**: Send details to the maintainer directly
2. **Private Security Report**: Use GitHub's private vulnerability reporting (if available)

## What to Include

When reporting, please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes (optional)

## Response Timeline

- We aim to acknowledge reports within 48 hours
- We will provide updates on remediation progress

## Supported Versions

| Version | Supported |
|--------|-----------|
| 1.0.x | ✅ |

## Security Best Practices for Deployment

- Use HTTPS/TLS in production
- Rotate credentials regularly
- Set appropriate token expiry times
- Restrict CORS origins to specific domains
- Monitor the `/metrics` endpoint for anomalies
- Keep Go version updated
- Run behind a reverse proxy (nginx, Caddy, etc.)

## Authentication Flow

MachineAuth uses OAuth 2.0 Client Credentials flow:

1. Agent requests token with `client_id` + `client_secret`
2. Server validates credentials
3. Server issues JWT (RS256 signed) + refresh token
4. Agent uses JWT to access protected resources
5. JWT expires (default: 1 hour)
6. Agent uses refresh token to get new access token

## Token Security

- Access tokens: Short-lived (configurable, default 1 hour)
- Refresh tokens: Long-lived (configurable, default 7 days)
- Tokens are signed with RS256 (asymmetric)
- Token revocation supported via `/oauth/revoke`
