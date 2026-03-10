-- 003_oauth_authorization_code.sql — OAuth 2.0 Authorization Code + PKCE support

-- ── OAuth Authorization Codes ───────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id              TEXT NOT NULL,
    user_id                TEXT NOT NULL,
    organization_id        TEXT NOT NULL,
    redirect_uri           TEXT NOT NULL,
    scope                  TEXT,
    code_challenge         VARCHAR(128),
    code_challenge_method  VARCHAR(10),
    code                   TEXT NOT NULL UNIQUE,
    expires_at            TIMESTAMPTZ NOT NULL,
    used                  BOOLEAN NOT NULL DEFAULT false,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oauth_codes_code ON oauth_authorization_codes(code);
CREATE INDEX IF NOT EXISTS idx_oauth_codes_client_id ON oauth_authorization_codes(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_codes_expires ON oauth_authorization_codes(expires_at);

-- ── Add OAuth fields to agents table ────────────────────────────────────────

ALTER TABLE agents ADD COLUMN IF NOT EXISTS redirect_uris TEXT[];
ALTER TABLE agents ADD COLUMN IF NOT EXISTS grant_types TEXT[];
ALTER TABLE agents ADD COLUMN IF NOT EXISTS client_type VARCHAR(20) DEFAULT 'confidential';
