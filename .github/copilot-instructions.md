# Copilot instructions for MachineAuth

## Build, test, lint
- Backend (Go): `go mod download`; `go build -o bin/server ./cmd/server`; run `go run ./cmd/server` (main entry) or `go run server-main.go` (alt port 8081); all tests `go test -v ./...`; single test `go test -v -run TestName ./path/...`; coverage `go test -v -coverprofile=cov.out ./...`; lint `golangci-lint run`; format `go fmt ./... && go vet ./...`.
- Frontend (React/Vite): `cd web && npm install`; dev `npm run dev`; build `npm run build`; lint `npm run lint`; typecheck `npx tsc --noEmit`.

## High-level architecture
- Go backend under `cmd/server` (entry) and `internal/` (config, db, handlers, middleware, models, services, utils); exposes OAuth2 endpoints (`/oauth/token|introspect|revoke`), agent CRUD/rotation, org/team/webhook APIs, readiness/health/metrics, JWKS.
- Storage: PostgreSQL by default; JSON file storage via `DATABASE_URL=json:machineauth.json` for dev; uses snake_case JSON payloads.
- Frontend admin UI in `web/` (React 18 + TypeScript + Vite + Tailwind); proxies `/api`, `/oauth`, `/.well-known`, `/health`, `/metrics` to backend (localhost:8081 in dev).
- SDKs: scaffolds in `sdk/typescript` (npm `@machineauth/sdk`) and `sdk/python` (`machineauth`) with token + agent operations; future phases planned for more languages.

## Key conventions
- Go: import order stdlib → external → internal (`goimports`); wrap errors with context using `fmt.Errorf(... %w ...)`; prefer `context.Context` plumbing; HTTP handlers return structured JSON and set content-type; log via `log.Printf`; types use pointers for optional fields; JSON keys snake_case.
- React/TS: absolute imports via `@/`; components PascalCase, functions camelCase, files kebab-case; avoid `any`; Tailwind with HSL variables; react-hook-form/sonner patterns.
- Environments: default backend port 8080 (alt 8081 for dev), CORS via `ALLOWED_ORIGINS`; admin credentials default `admin/admin` (change before deploy).
