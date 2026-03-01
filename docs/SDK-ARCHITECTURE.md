# MachineAuth SDK Architecture & Rollout

## Purpose
Define a unified, idiomatic SDK architecture across Python, TypeScript/JavaScript, Go, Rust, Java, C#, and later mobile, mirroring the API’s resource hierarchy with consistent configuration, error handling, and token management.

## Scope and API surface
- Resources: agents (create/list/get/delete/rotate), token operations (token, refresh, introspect, revoke), self-service (profile, usage, rotation), organizations/teams, API keys, jwks/health/metrics (as needed for diagnostics).
- Core behaviors: environment-driven configuration with constructor overrides, automatic token acquisition/refresh, revocation support, structured errors, and consistent HTTP defaults (timeouts, retries, user agent, telemetry hooks).

## Cross-language design principles
- Resource-based clients that align to API paths; shared base client handling auth, transport, retries, and serialization.
- Strong typing and validation (Python type hints, TS strict mode, Go/Rust static types); minimize `any`/untagged errors.
- Error taxonomy: transport vs. auth vs. validation vs. server; include HTTP status, machine-readable code, message, request id/trace id.
- Auth & tokens: client credentials first-class; refresh tokens supported; optional DPoP/mTLS left feature-flagged until confirmed.
- Token storage: pluggable (memory by default), with hooks to persist/evict; background refresh when near expiry where idiomatic.
- HTTP defaults: sane timeouts, limited retries with exponential backoff for idempotent calls, configurable max attempts; opt-in tracing/logging hooks.
- Configuration: builder-style per language, supports env var defaults, base URL, timeouts, proxy, user agent, custom headers, scopes.
- Models: generated or shared schema to keep parity across SDKs; keep snake_case JSON to match API.

## Language-specific notes
- **Python (3.9+)**: async/await-first client (httpx recommended), context managers for session/token lifecycle, mypy-compatible types, sync facade optional. Package name TBD (proposed: `machineauth`). Distribute via PyPI with wheel + sdist.
- **TypeScript/JavaScript**: Promise-based API with fetch adapter; Node + browser compatibility; React hooks as optional helper; strict TS config. Package name TBD (proposed: `@machineauth/sdk`). Publish via npm.
- **Go (1.21+)**: Context-aware clients, `%w` error wrapping with sentinel/types, config builder, token store interface, minimal dependencies. Module path TBD (proposed: `github.com/mandarwagh9/machineauth/sdk`).
- **Rust**: Async client (reqwest-based) with feature flags; strongly typed models and error enums; later phase.
- **Java/C#**: Enterprise packaging (Maven/NuGet), builder configs, token lifecycle helpers; later phase.
- **Mobile (Kotlin/Swift)**: Secure token storage, mobile networking constraints; post-stabilization.

## Packaging, versioning, releases (assumptions to confirm)
- Proposed IDs: Python `machineauth`, npm `@machineauth/sdk`, Go module `github.com/mandarwagh9/machineauth/sdk`; semver with shared changelog and per-SDK release notes.
- CI: lint/type-check/test per language; publish pipelines for PyPI/npm and later Maven/NuGet; contract tests against local MachineAuth server.

## Phase deliverables
- **Phase 1 (Python + TS core)**: Scaffold repos/packages, implement agents, tokens, self-service, org/teams, API keys; config builders; error types; token store; basic retries; examples.
- **Phase 2 (advanced features)**: Async/middleware interceptors, richer retries/backoff, tracing/logging hooks, pluggable caches/stores, hardened config (proxy/timeout), expanded examples/docs.
- **Phase 3 (Go + Rust)**: Idiomatic clients mirroring Phase 1/2 surfaces, contexts/error wrapping (Go) and typed enums (Rust).
- **Phase 4 (Java/C#)**: Enterprise SDKs with same surface and publishing pipelines.
- **Phase 5 (mobile)**: Kotlin/Swift variants with secure storage and mobile-friendly networking.

## Quality and docs
- Cross-language spec: resources, request/response models, error catalog, configuration contract kept in repo; drive codegen/tests where feasible.
- Samples per language: minimal token fetch, agent lifecycle, React hook example, Go CLI; include quickstart snippets in README.
- Parity checks: automated conformance tests comparing SDK requests/responses to API fixtures; CI gates on parity drift.

## Pending decisions (needs confirmation)
- Final package/module names per language.
- Minimum runtime targets (Python 3.9+, TS lib/target, browser support matrix).
- Whether to include DPoP/mTLS in v1 or defer behind feature flags.
