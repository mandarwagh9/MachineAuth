-- 002_multi_tenancy.sql — Org members, org signing keys, and org-scoped columns

-- ── Org Members (user<->org membership with role) ──────────────────────

CREATE TABLE IF NOT EXISTS org_members (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID         NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    organization_id TEXT         NOT NULL,
    role            VARCHAR(50)  NOT NULL DEFAULT 'member',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, organization_id)
);

CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON org_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_org_id  ON org_members(organization_id);

-- ── Org Signing Keys (per-org RSA keys for JWT signing) ────────────────

CREATE TABLE IF NOT EXISTS org_signing_keys (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id TEXT         NOT NULL,
    key_id          VARCHAR(255) NOT NULL,
    public_key_pem  TEXT         NOT NULL,
    private_key_pem TEXT         NOT NULL,
    algorithm       VARCHAR(50)  NOT NULL DEFAULT 'RS256',
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_org_signing_keys_org_id ON org_signing_keys(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_signing_keys_active ON org_signing_keys(organization_id, is_active) WHERE is_active = true;

-- ── Add organization_id to admin_users for default org binding ─────────

DO $$ BEGIN
    ALTER TABLE admin_users ADD COLUMN organization_id TEXT DEFAULT '';
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- ── Add organization_id to audit_logs for org-scoped audit ─────────────

DO $$ BEGIN
    ALTER TABLE audit_logs ADD COLUMN organization_id TEXT DEFAULT '';
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_audit_logs_org_id ON audit_logs(organization_id);
