# MachineAuth  Agent Auth SaaS Platform Plan

> **"Auth0 for AI Agents"** — A managed SaaS that any AI agent platform (OpenAI, CrewAI, Anthropic) can plug into for identity, credentials, tokens, and access control.

**Created:** March 3, 2026  
**Current Version:** 2.12.61  
**Status:** Wave 1 (Security Hardening) complete, SaaS buildout starting

---

## Table of Contents

- [Current State](#current-state)
- [Target State](#target-state)
- [Build Sequence](#build-sequence)
  - [Layer 0: PostgreSQL — The Foundation](#layer-0-postgresql--the-foundation)
  - [Layer 1: Real Multi-Tenancy](#layer-1-real-multi-tenancy)
  - [Layer 2: OIDC Compliance](#layer-2-oidc-compliance--act-as-an-identity-provider)
  - [Layer 3: Billing & Plan Enforcement](#layer-3-billing--plan-enforcement)
  - [Layer 4: Agent Framework Adapters](#layer-4-agent-framework-adapters)
  - [Layer 5: Developer Portal & Docs](#layer-5-developer-portal--documentation)
  - [Layer 6: Cloud Infrastructure](#layer-6-cloud-infrastructure)
  - [Layer 7: Advanced Features](#layer-7-advanced-features-post-launch)
- [Wave 1 Completed Work](#wave-1-completed-work)
- [Pricing Model](#pricing-model)
- [Key Decisions](#key-decisions)
- [Verification Criteria](#verification-criteria)
- [SaaS Readiness Scorecard](#saas-readiness-scorecard)

---

## Current State

### What Works Today

| Capability | Status | Grade |
|---|---|---|
| OAuth2 Core (client_credentials, refresh, revoke, introspect) | Solid | **B+** |
| JWT/RS256 + JWKS | Working | **C+** |
| Agent CRUD + self-service | Working (update, metadata, tags, status added in Wave 1) | **B** |
| Admin JWT auth | Working (Wave 1 — was plaintext, now JWT-based) | **B+** |
| All admin routes protected | Working (Wave 1 — was unauthenticated) | **A** |
| Rate limiting | Working (Wave 1 — IP-based fixed window) | **B** |
| Security headers | Working (Wave 1 — HSTS, nosniff, etc.) | **A** |
| Brute-force protection | Working (Wave 1 — 5 failures  5min lockout) | **B+** |
| Pagination + filtering | Working (Wave 1 — search, status, org filter) | **B** |
| Audit log query API | Working (Wave 1 — filter by agent/action/date) | **B** |
| Webhook system | Comprehensive (HMAC-SHA256, retry, delivery tracking) | **A-** |
| SDKs (TypeScript + Python) | Full API coverage | **A** |
| Multi-tenancy | Structural only — models exist, no enforcement | **D+** |
| Billing / plans | `Plan` field exists, never checked | **F** |
| OIDC compliance | JWKS only, no discovery | **F** |
| Database | JSON file only, no real Postgres driver | **F** |
| Agent framework integrations | None | **F** |
| Deployment | Single VPS, manual scripts | **D** |

### Critical Architecture Gaps

1. **`db.Connect()` always returns JSON DB** — even with a `postgres://` URL, it falls through to JSON. No PostgreSQL driver code exists.
2. **`RunMigrations()` is a no-op** — returns `nil`.
3. **All list queries are global** — `ListAgents()`, `ListWebhooks()`, `ListAuditLogs()` return all data across all orgs.
4. **API keys exist but have no middleware** — `sk_*` keys are in the DB but no handler validates them for auth.
5. **Per-org signing keys unused** — `Organization.JWTIssuer` and `JWTExpirySecs` fields are stored but ignored in token generation.
6. **Single RSA key pair** — all tenants share one signing key, stored as PEM on disk.
7. **In-memory everything** — rate limiter, brute-force state, and revoked token list are all in-memory, preventing horizontal scaling.
8. **Zero references to any AI agent framework** in the entire codebase.

---

## Target State

A managed cloud platform where:

1. **Platform teams** (building with OpenAI Agents SDK, CrewAI, Anthropic) sign up, create an org, and get API keys
2. **Each org** has isolated agents, signing keys, webhooks, and audit logs
3. **Framework-specific SDKs** let developers add auth to their agents in <5 minutes
4. **OIDC compliance** means MachineAuth can be configured as a trusted IdP in AWS IAM, GCP, K8s, Vercel
5. **MAU-based billing** with Stripe charges per active agent per month, gated by feature tiers
6. **Enterprise customers** get dedicated tenant isolation, BYOK, and data residency

---

## Build Sequence

Each layer is prerequisite for the next. Don't skip ahead.

### Layer 0: PostgreSQL — The Foundation

> *Nothing else works at SaaS scale on a JSON file.*

| # | Step | File(s) | Details |
|---|------|---------|---------|
| 1 | **Implement PostgreSQL driver** | `internal/db/db.go` | Add `connectPostgres()` path when `DATABASE_URL` starts with `postgres://`. Implement all ~40 DB methods against SQL using `pgx` or `database/sql` + `lib/pq`. |
| 2 | **Schema migrations** | `internal/db/migrations/*.sql` | Build migration runner using embedded SQL files. Tables: `agents`, `admin_users`, `organizations`, `teams`, `api_keys`, `audit_logs`, `refresh_tokens`, `revoked_tokens`, `webhook_configs`, `webhook_deliveries`, plus new: `subscriptions`, `usage_events`, `org_signing_keys`. |
| 3 | **Connection pooling + health** | `internal/db/db.go`, `cmd/server/main.go` | Add pool metrics to Prometheus. Validate DB connectivity in `/health/ready`. |
| 4 | **Keep JSON DB for dev** | `internal/db/db.go` | `DATABASE_URL=json:machineauth.json` continues to work for local development. |

**Verify:** `DATABASE_URL=postgres://...`  all CRUD operations work. `go test -v ./...` passes against both backends.

---

### Layer 1: Real Multi-Tenancy

> *Every query, every response, every key — scoped to an organization.*

| # | Step | File(s) | Details |
|---|------|---------|---------|
| 5 | **Org-scoped middleware** | `internal/middleware/auth.go` | Extend `AdminAuth` to extract `org_id` from admin JWT. Inject into context. Every handler reads it. |
| 6 | **Org-scoped queries** | `internal/db/db.go`, all service files | Refactor every `List*()`, `Get*()`, `Create*()` to accept and filter by `organization_id`. |
| 7 | **API key auth middleware** | `internal/middleware/auth.go` | Accept `Authorization: Bearer sk_...` or `X-API-Key: sk_...`. Validate against DB. Inject `org_id` + `team_id`. This is how customers' backends authenticate. |
| 8 | **Per-org admin users + RBAC** | `internal/db/db.go`, `internal/services/admin.go`, `internal/models/models.go` | New `org_members` table: `user_id`, `org_id`, `role` (owner/admin/member/viewer). Sign-up creates org + owner atomically. |
| 9 | **Per-org signing keys** | `internal/services/token.go` | Generate unique RSA key pair per org. Store in `org_signing_keys` table (encrypted at rest). JWKS returns all active keys. Each org's tokens use their own `iss` claim. |
| 10 | **Per-org webhook scoping** | `internal/services/webhook.go` | `ListWebhooks()`, `ListActiveWebhooksForEvent()`, and `TriggerEvent()` filter by org. Agent event  only that org's webhooks fire. |

**Verify:** Admin for org-A lists agents  sees only org-A agents. API key from org-B  403 on org-A resources. Cross-org data leakage test.

---

### Layer 2: OIDC Compliance — Act as an Identity Provider

> *External systems (AWS, K8s, Vercel) can trust MachineAuth as an IdP.*

| # | Step | File(s) | Details |
|---|------|---------|---------|
| 11 | **OIDC Discovery** | `cmd/server/main.go`, new handler | `GET /.well-known/openid-configuration` returning standard discovery document: `issuer`, `token_endpoint`, `jwks_uri`, `grant_types_supported`, `scopes_supported`, etc. |
| 12 | **ID token issuance** | `internal/services/token.go` | When `scope` includes `openid`, include `id_token` JWT in response with OIDC claims: `sub`, `iss`, `aud`, `iat`, `exp`, `auth_time`, `name`, `org_id`, `metadata`. |
| 13 | **UserInfo endpoint** | `internal/handlers/auth.go` | `GET /oauth/userinfo` (access-token-protected): returns agent profile (`sub`, `name`, `org_id`, `team_id`, `scopes`, `metadata`, `status`). |
| 14 | **Per-org OIDC** | Multiple | Each org gets `/.well-known/{org_slug}/openid-configuration` with their own issuer URL and signing keys. Enables per-customer OIDC configuration in AWS IAM/GCP/K8s. |

**Verify:** `curl /.well-known/openid-configuration`  valid OIDC discovery document. Configure MachineAuth as OIDC provider in AWS IAM  agents can assume roles.

---

### Layer 3: Billing & Plan Enforcement

> *MAU-based + flat tiers, enforced at every API call.*

| # | Step | Details |
|---|------|---------|
| 15 | **Plan definitions** | Config table `plans` — see [Pricing Model](#pricing-model) below. |
| 16 | **Stripe integration** | `internal/services/billing.go`: customer creation on org signup, Checkout for subscriptions, webhook handler for `invoice.paid`/`subscription.updated`/`subscription.deleted`. Store `stripe_customer_id` on Organization. |
| 17 | **Usage metering** | Track per-org: active agents/month (any agent that issued a token), total tokens, webhook deliveries. Report to Stripe as metered usage. |
| 18 | **Plan enforcement middleware** | Before creating agents, issuing tokens, creating webhooks: check plan limits. `402 Payment Required` when exceeded. Warning headers before hard cutoff. |
| 19 | **Self-service billing portal** | Frontend page: current plan, usage graph, invoices, upgrade/downgrade. Stripe Customer Portal for payment methods. |

**Verify:** Free org  11th agent returns 402. Stripe webhook  subscription cancel  org degraded.

---

### Layer 4: Agent Framework Adapters

> *The SDKs that make MachineAuth plug-and-play for target platforms.*

| # | Adapter | Package | What It Does |
|---|---------|---------|--------------|
| 20 | **OpenAI Agents SDK** | `@machineauth/openai-agents` | Credential provider for OpenAI's Agents SDK. Auto-manages per-agent identity. Token injection into function calls. Token refresh. |
| 21 | **CrewAI** | `machineauth[crewai]` (Python) | `MachineAuthCredentialProvider` for CrewAI agent config. Per-crew-member identity. Token injection into tool calls. Audit trail mapping. |
| 22 | **Anthropic tool-use** | `@machineauth/anthropic` | Auth wrapper for Claude tool_use. Agent-scoped tool permissions via scopes. JWT verification in tool servers. |
| 23 | **Generic middleware** | `@machineauth/middleware` | Framework-agnostic HTTP middleware (Express, ASGI, Go net/http) that verifies MachineAuth JWTs, extracts agent identity + scopes. |
| 24 | **SDK auto-refresh** | `sdk/typescript/`, `sdk/python/` | Background token refresh before expiry in existing SDKs. Token cache with TTL. |
| 25 | **Go SDK** | `sdk/go/` | Implement the spec from `docs/GO-SDK.md`. Full API coverage with context plumbing. |

**Verify:** CrewAI crew with `MachineAuthCredentialProvider`  each agent authenticates. OpenAI agent calls tool with MachineAuth token  tool verifies JWT.

---

### Layer 5: Developer Portal & Documentation

| # | Step | Details |
|---|------|---------|
| 26 | **OpenAPI 3.1 spec** | Complete spec for all endpoints (OAuth, OIDC, admin, agents, orgs, webhooks, audit, billing). |
| 27 | **Documentation site** | Mintlify or Docusaurus. Quickstarts per framework, concept docs, API reference, SDK references. |
| 28 | **Interactive API explorer** | Swagger UI at `/api/docs` in admin dashboard. |
| 29 | **Onboarding flow** | Create org  create agent  copy credentials  pick framework  show code snippet  done. |
| 30 | **Webhook payload schemas** | JSON Schema for every webhook event payload with examples. |

---

### Layer 6: Cloud Infrastructure

| # | Step | Details |
|---|------|---------|
| 31 | **Kubernetes deployment** | Helm chart: API pods (HPA), managed PostgreSQL, Redis (distributed rate limiting + token revocation cache + brute-force state), separate webhook worker deployment. |
| 32 | **Multi-region** | Primary region control plane, read replicas, CDN for JWKS endpoint (high-frequency, cacheable). |
| 33 | **Dedicated tenant isolation** | Enterprise: dedicated DB schema/database, dedicated signing keys, BYOK (import RSA/EC keys), dedicated compute (K8s namespace), data residency (EU/US). |
| 34 | **Secrets management** | Move RSA keys from PEM files to Vault/KMS. Encrypt per-org keys at rest in DB. |
| 35 | **Observability** | Wire all Prometheus metrics (many declared but never updated in `metrics.go`). Grafana dashboards. OpenTelemetry tracing. |
| 36 | **CI/CD** | Automated test  build  deploy pipeline. Replace manual `deploy.sh`. |

**Verify:** `helm install machineauth`  working cluster. Load test: 10k concurrent token issuances  p99 < 100ms.

---

### Layer 7: Advanced Features (Post-Launch)

| # | Feature | Details |
|---|---------|---------|
| 37 | **Actions/scripting engine** | Embedded JS (goja) for custom logic at auth events. `pre-token-issuance` scripts can modify JWT claims. Auth0-style `api.accessToken.setCustomClaim()`. |
| 38 | **DPoP + mTLS** | Proof-of-possession for agents. `PublicKey` field exists on agents but is unused. Implement RFC 9449 (DPoP) and RFC 8705 (mTLS cert-bound tokens). |
| 39 | **Workload Identity Federation** | Accept AWS STS, GCP OIDC, Azure MSI tokens as agent identity proof. Map external workload identities to MachineAuth agents. No static secrets needed. |
| 40 | **Agent-to-agent delegation** | Token exchange (RFC 8693): Agent A requests scoped token on behalf of Agent B. Enables agent collaboration with proper authz chains. |
| 41 | **Anomaly detection** | Rules or ML-based: abnormal token volume, new IP ranges, scope escalation, activity outside time windows. Auto-suspend + alert. |
| 42 | **Compliance** | SOC 2 Type II, ISO 27001, GDPR DPA. Required for enterprise customers. |

---

## Wave 1 Completed Work

> These changes are already merged as of March 3, 2026.

### Files Created
- `internal/services/admin.go` — AdminService: JWT-based admin sessions, bcrypt password hashing, auto-creates default admin on startup
- `internal/middleware/security.go` — Rate limiter (IP-based fixed window), security headers, request body size limiter (1MB)
- `internal/handlers/audit.go` — Audit log query handler with filtering & pagination

### Files Modified

**Models** (`internal/models/models.go`):
- `Agent` now has `Description`, `Tags`, `Metadata map[string]interface{}`, `Status` (enum: active/inactive/suspended/pending/expired)
- Added `UpdateAgentRequest`, `PaginationParams`, `Pagination`, `AuditLogQuery`, `AuditLogsListResponse`
- Added `AdminUser`, `AdminLoginRequest` (supports both `email` and legacy `username`), `AdminTokenResponse` (returns JWT)
- `AgentsListResponse` now includes optional `Pagination`

**DB layer** (`internal/db/db.go`):
- `Agent` struct extended with `Description`, `Tags`, `Metadata`, `Status`
- Added `AdminUser` storage + CRUD methods
- Added `ListAuditLogs` with filtering (agent_id, action, ip, date range) + pagination
- Added `ListAgentsPaginated` with search, status/org filter, sorting

**Agent service** (`internal/services/agent.go`):
- `Create` populates new fields
- Added `Update` method (partial updates via `UpdateAgentRequest`)
- Added `ListPaginated` with filtering

**Auth handler** (`internal/handlers/auth.go`):
- `AdminLogin` now issues JWT via AdminService (was returning `{success: true}`)
- Built-in brute-force protection: 5 failed attempts  5min lockout (admin login + client_credentials)

**Agents handler** (`internal/handlers/agents.go`):
- `PUT/PATCH /api/agents/:id`  `updateAgent` handler
- `GET /api/agents?page=&limit=&q=&status=&org_id=&sort=`  paginated listing

**Main** (`cmd/server/main.go`):
- All `/api/*` admin routes wrapped with `AdminAuth` middleware
- OAuth endpoints rate-limited (30/min per IP)
- Admin login rate-limited (10/min per IP)
- API endpoints rate-limited (120/min per IP)
- Middleware chain: SecurityHeaders  BodyLimit(1MB)  Logging  CORS  Mux

### Admin Login Flow (New)
```
POST /api/auth/login  {"email":"admin@example.com","password":"changeme"}
 200 {"success":true,"access_token":"eyJhbG...","expires_in":28800,"role":"owner"}

GET /api/agents  Authorization: Bearer eyJhbG...
 200 {"agents":[...],"pagination":{"total":5,"page":1,"limit":20,"total_pages":1}}

GET /api/agents  (no token)
 401 {"error":"unauthorized","error_description":"missing authorization header"}
```

---

## Pricing Model

### Tiers

| Feature | Free | Pro ($49/mo) | Enterprise (Custom) |
|---------|------|-------------|-------------------|
| Active agents | 10 | 100 | Unlimited |
| Token issuance | 1k/day | 50k/day | Unlimited |
| Webhooks | 2 | 20 | Unlimited |
| Custom claims (Actions) |  |  |  |
| DPoP / mTLS |  | DPoP | Both |
| Dedicated signing keys |  |  |  |
| OIDC IdP |  |  |  |
| Framework SDKs |  |  |  |
| Audit log retention | 7 days | 90 days | 1 year |
| Support | Community | Email | Priority + Slack |
| Data residency |  |  |  (EU/US) |
| BYOK (bring your own key) |  |  |  |
| Dedicated compute |  |  |  |

### MAU Overage

$0.01 per active agent per month above plan limit.
"Active agent" = any agent that issued or refreshed a token during the billing month.

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Build order | Postgres  Multi-tenancy  OIDC  Billing  Adapters  Docs  Infra | Each layer is prerequisite for the next |
| Database | PostgreSQL first, JSON stays for dev | Can't run multi-tenant SaaS on a JSON file |
| Billing model | MAU + flat tiers | Feature-gated tiers (like Auth0), usage-based overage on agent count |
| Target platforms | OpenAI Agents SDK, CrewAI, Anthropic | Dominant agent ecosystems. Generic middleware covers the rest |
| Hosting model | Managed cloud + dedicated tenant for enterprise | Standard multi-tenant for free/pro, isolated for enterprise |
| Admin auth | JWT sessions (done in Wave 1) | Enables RBAC, expiry, proper session management |
| DPoP + mTLS both | DPoP for HTTP-layer (most agents), mTLS for hardware-secured (IoT/infra) | Different security profiles for different agent types |
| Scripting engine | Embedded JS via goja | Synchronous in token pipeline, no network latency, familiar Auth0-style API |
| OIDC before federation | OIDC discovery + ID tokens first | Required infrastructure before external IdP federation |

---

## Verification Criteria

### Layer 0 (Postgres)
- `DATABASE_URL=postgres://...`  all CRUD operations work
- `go test -v ./...` passes against both JSON and Postgres backends
- Migration runner creates all tables from scratch

### Layer 1 (Multi-tenancy)
- Admin for org-A lists agents  sees only org-A agents
- API key from org-B  403 on org-A resources
- Cross-org data leakage test: create entities in 2 orgs, verify strict isolation

### Layer 2 (OIDC)
- `curl /.well-known/openid-configuration`  valid OIDC discovery document
- Configure MachineAuth as OIDC provider in AWS IAM  agents can assume AWS roles
- Token with `scope=openid`  response includes `id_token`

### Layer 3 (Billing)
- Free plan org  11th agent creation returns 402
- Stripe webhook  subscription cancellation  org degraded to free
- Usage dashboard shows accurate MAU count

### Layer 4 (Adapters)
- CrewAI crew with `MachineAuthCredentialProvider`  each agent authenticates, audit trail shows
- OpenAI agent calls tool with MachineAuth token  tool verifies JWT
- Generic middleware on Express app  agent identity extracted from JWT

### Layer 6 (Infra)
- `helm install machineauth`  working cluster with 3 API replicas, Redis, Postgres
- Load test: 10k concurrent token issuances  p99 < 100ms
- JWKS endpoint served via CDN with <10ms latency globally

---

## SaaS Readiness Scorecard

*Update this as layers are completed.*

| Dimension | Current Score | Target | Blocking Layer |
|-----------|:---:|:---:|---|
| OAuth 2.0 Core | 8/10 | 10/10 | Layer 2 (OIDC) |
| Agent Identity Management | 8/10 | 10/10 | Layer 1 (multi-tenancy) |
| Multi-Tenancy | 3/10 | 10/10 | **Layer 0 (Postgres)** |
| Billing / Plans | 1/10 | 9/10 | Layer 3 |
| API Keys for External Auth | 3/10 | 9/10 | Layer 1 (middleware) |
| SDK Completeness | 6/10 | 10/10 | Layers 4-5 |
| Documentation | 6/10 | 10/10 | Layer 5 |
| Deployment / Infra | 2/10 | 9/10 | **Layer 0 (Postgres)**, Layer 6 |
| OIDC / Federation | 2/10 | 10/10 | Layer 2 |
| Webhook System | 6/10 | 9/10 | Layer 1 (org scoping) |
| Agent Framework Integration | 0/10 | 9/10 | Layer 4 |

---

## File Reference

Key files in the current codebase:

| Purpose | Path |
|---------|------|
| Server entry + route wiring | `cmd/server/main.go` |
| Models (Agent, Org, Webhook, etc.) | `internal/models/models.go` |
| Database layer (JSON, needs Postgres) | `internal/db/db.go` |
| Agent service (CRUD, update, paginate) | `internal/services/agent.go` |
| Token service (JWT, JWKS, refresh) | `internal/services/token.go` |
| Admin service (JWT sessions) | `internal/services/admin.go` |
| Auth handler (login, token, introspect) | `internal/handlers/auth.go` |
| Agent handler (CRUD, update) | `internal/handlers/agents.go` |
| Audit handler (query API) | `internal/handlers/audit.go` |
| Webhook service + delivery worker | `internal/services/webhook.go`, `webhook_worker.go` |
| Auth middleware (JWT, Admin, future API key) | `internal/middleware/auth.go` |
| Security middleware (rate limit, headers) | `internal/middleware/security.go` |
| Config | `internal/config/config.go` |
| TypeScript SDK | `sdk/typescript/src/index.ts` |
| Python SDK | `sdk/python/machineauth/client.py` |
| Frontend | `web/src/App.tsx` |
| API docs (comprehensive) | `soul.md` |
| Multi-tenant design doc | `docs/MULTITENANT.md` |
| Go SDK spec (not implemented) | `docs/GO-SDK.md` |
