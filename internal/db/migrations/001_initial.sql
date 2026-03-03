-- 001_initial.sql — MachineAuth schema
-- All tables needed by the platform.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ── Agents ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS agents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id TEXT        NOT NULL DEFAULT '',
    team_id         TEXT,
    name            VARCHAR(255) NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    tags            TEXT[]      DEFAULT '{}',
    metadata        JSONB       DEFAULT '{}',
    client_id       VARCHAR(255) UNIQUE NOT NULL,
    client_secret_hash VARCHAR(255) NOT NULL,
    scopes          TEXT[]      DEFAULT '{}',
    public_key      TEXT,
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    token_count     INTEGER     NOT NULL DEFAULT 0,
    refresh_count   INTEGER     NOT NULL DEFAULT 0,
    last_activity_at    TIMESTAMPTZ,
    last_token_issued_at TIMESTAMPTZ,
    rotation_history JSONB      DEFAULT '[]'
);

-- ── Admin Users ────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS admin_users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Audit Logs ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS audit_logs (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id   UUID,
    action     VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45)  DEFAULT '',
    user_agent TEXT         DEFAULT '',
    details    TEXT         DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Refresh Tokens ─────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY,
    agent_id   UUID         NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- ── Revoked Tokens ─────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS revoked_tokens (
    jti     VARCHAR(255) PRIMARY KEY,
    expires TIMESTAMPTZ  NOT NULL
);

-- ── Metrics (singleton row) ────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS metrics (
    id               INTEGER PRIMARY KEY DEFAULT 1,
    tokens_refreshed BIGINT NOT NULL DEFAULT 0,
    tokens_revoked   BIGINT NOT NULL DEFAULT 0
);

INSERT INTO metrics (id, tokens_refreshed, tokens_revoked)
VALUES (1, 0, 0)
ON CONFLICT (id) DO NOTHING;

-- ── Organizations ──────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS organizations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) UNIQUE NOT NULL,
    owner_email     VARCHAR(255) NOT NULL,
    jwt_issuer      TEXT         DEFAULT '',
    jwt_expiry_secs INTEGER      DEFAULT 3600,
    allowed_origins TEXT         DEFAULT '',
    plan            VARCHAR(50)  DEFAULT 'free',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Teams ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS teams (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id TEXT         NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT         DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── API Keys ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS api_keys (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id TEXT         NOT NULL,
    team_id         TEXT,
    name            VARCHAR(255) NOT NULL,
    key_hash        VARCHAR(255) NOT NULL,
    prefix          VARCHAR(20)  NOT NULL,
    last_used_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Webhook Configs ────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS webhook_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id   TEXT         DEFAULT '',
    team_id           TEXT         DEFAULT '',
    name              VARCHAR(255) NOT NULL,
    url               TEXT         NOT NULL,
    secret            VARCHAR(255) NOT NULL,
    events            TEXT[]       NOT NULL DEFAULT '{}',
    is_active         BOOLEAN      NOT NULL DEFAULT true,
    max_retries       INTEGER      NOT NULL DEFAULT 10,
    retry_backoff_base INTEGER     NOT NULL DEFAULT 2,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_tested_at    TIMESTAMPTZ,
    consecutive_fails INTEGER      NOT NULL DEFAULT 0
);

-- ── Webhook Deliveries ─────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_config_id UUID         NOT NULL REFERENCES webhook_configs(id) ON DELETE CASCADE,
    event             VARCHAR(100) NOT NULL,
    payload           TEXT         NOT NULL DEFAULT '',
    headers           TEXT         DEFAULT '',
    status            VARCHAR(50)  NOT NULL DEFAULT 'pending',
    attempts          INTEGER      NOT NULL DEFAULT 0,
    last_attempt_at   TIMESTAMPTZ,
    last_error        TEXT         DEFAULT '',
    next_retry_at     TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Indexes ────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_agents_client_id      ON agents(client_id);
CREATE INDEX IF NOT EXISTS idx_agents_org_id         ON agents(organization_id);
CREATE INDEX IF NOT EXISTS idx_agents_status         ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_name_lower     ON agents(LOWER(name));

CREATE INDEX IF NOT EXISTS idx_audit_logs_agent_id   ON audit_logs(agent_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action     ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created    ON audit_logs(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_agent  ON refresh_tokens(agent_id);
CREATE INDEX IF NOT EXISTS idx_revoked_tokens_exp    ON revoked_tokens(expires);

CREATE INDEX IF NOT EXISTS idx_teams_org_id          ON teams(organization_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_org_id       ON api_keys(organization_id);

CREATE INDEX IF NOT EXISTS idx_wh_deliveries_config  ON webhook_deliveries(webhook_config_id);
CREATE INDEX IF NOT EXISTS idx_wh_deliveries_status  ON webhook_deliveries(status);

-- ── Schema Migrations Tracking ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS schema_migrations (
    version   INTEGER PRIMARY KEY,
    name      VARCHAR(255) NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
