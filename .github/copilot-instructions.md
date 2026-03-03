```instructions
# Copilot instructions for MachineAuth

## Build, test, lint
- Backend (Go): `go mod download`; `go build -o bin/server ./cmd/server`; run `go run ./cmd/server` (main entry, port 8080) or `go run server-main.go` (alt port 8081); all tests `go test -v ./...`; single test `go test -v -run TestName ./path/...`; coverage `go test -v -coverprofile=cov.out ./...`; lint `golangci-lint run`; format `go fmt ./... && go vet ./...`.
- Frontend (React/Vite): `cd web && npm install`; dev `npm run dev` (port 3000); build `npm run build`; lint `npm run lint`; typecheck `npx tsc --noEmit`.

## High-level architecture
- Go backend under `cmd/server` (entry) and `internal/` (config, db, handlers, middleware, models, services, utils); exposes OAuth2 endpoints (`/oauth/token|introspect|revoke|refresh`), agent CRUD/rotation, org/team/webhook APIs, readiness/health/metrics, JWKS.
- Storage: PostgreSQL by default; JSON file storage via `DATABASE_URL=json:machineauth.json` for dev; `db.Connect` selects driver based on prefix. Uses snake_case JSON payloads.
- Frontend admin UI in `web/` (React 18 + TypeScript + Vite + Tailwind); proxies `/api`, `/oauth`, `/.well-known`, `/health`, `/metrics` to backend. Set `VITE_API_URL=http://localhost:8081` locally.
- Routes wired manually in `cmd/server/main.go` onto `http.NewServeMux` - no third-party router.
- Webhook delivery worker (`services.DeliveryWorker`) runs as goroutine pool; audit events trigger webhooks via `AuditService.SetWebhookService`.
- SDKs in `sdk/typescript` (`@machineauth/sdk`) and `sdk/python` (`machineauth`).

## Key conventions
- Go: import order stdlib -> external -> internal (`goimports`); wrap errors `fmt.Errorf("context: %w", err)`; `context.Context` plumbing preferred; handlers always set `Content-Type: application/json`; log via `log.Printf`.
- Models: pointers for optional fields; `ClientSecretHash` has `json:"-"` (never serialized); `uuid.UUID` embedded directly (see `internal/models/models.go`).
- Middleware: `middleware.JWTAuth` extracts `uuid.UUID` into `middleware.AgentIDKey` context key; retrieve with `middleware.GetAgentIDFromContext(ctx)` (see `internal/middleware/auth.go`).
- React/TS: absolute imports via `@/`; components PascalCase, functions camelCase, files kebab-case; avoid `any`; Tailwind with HSL variables; `axios` in `web/src/services/api.ts`.
- Environments: backend port 8080 (alt 8081 for dev), CORS via `ALLOWED_ORIGINS`; admin default `admin/admin` (change before deploy).
- Version string in root handler response in `cmd/server/main.go`.
```