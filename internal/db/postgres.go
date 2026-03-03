package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB implements the Database interface using PostgreSQL via pgx.
type PostgresDB struct {
	pool *pgxpool.Pool
}

func connectPostgres(databaseURL string) (Database, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &PostgresDB{pool: pool}, nil
}

// Pool exposes the underlying pool for health checks.
func (p *PostgresDB) Pool() *pgxpool.Pool { return p.pool }

func (p *PostgresDB) Close() error {
	p.pool.Close()
	return nil
}

// ═══════════════════════════════════════════════════════════════════════
// Agents
// ═══════════════════════════════════════════════════════════════════════

const agentCols = `id::text, organization_id, team_id, name, description,
	tags, metadata, client_id, client_secret_hash, scopes, public_key,
	status, is_active, created_at, updated_at, expires_at,
	token_count, refresh_count, last_activity_at, last_token_issued_at,
	rotation_history`

func scanAgent(row pgx.Row) (*Agent, error) {
	var a Agent
	var metaJSON, rotJSON []byte
	err := row.Scan(
		&a.ID, &a.OrganizationID, &a.TeamID,
		&a.Name, &a.Description, &a.Tags, &metaJSON,
		&a.ClientID, &a.ClientSecretHash, &a.Scopes, &a.PublicKey,
		&a.Status, &a.IsActive, &a.CreatedAt, &a.UpdatedAt, &a.ExpiresAt,
		&a.TokenCount, &a.RefreshCount, &a.LastActivityAt, &a.LastTokenIssuedAt,
		&rotJSON,
	)
	if err != nil {
		return nil, err
	}
	if metaJSON != nil {
		_ = json.Unmarshal(metaJSON, &a.Metadata)
	}
	if rotJSON != nil {
		_ = json.Unmarshal(rotJSON, &a.RotationHistory)
	}
	if a.Tags == nil {
		a.Tags = []string{}
	}
	if a.Scopes == nil {
		a.Scopes = []string{}
	}
	return &a, nil
}

func scanAgents(rows pgx.Rows) ([]Agent, error) {
	var agents []Agent
	for rows.Next() {
		var a Agent
		var metaJSON, rotJSON []byte
		err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.TeamID,
			&a.Name, &a.Description, &a.Tags, &metaJSON,
			&a.ClientID, &a.ClientSecretHash, &a.Scopes, &a.PublicKey,
			&a.Status, &a.IsActive, &a.CreatedAt, &a.UpdatedAt, &a.ExpiresAt,
			&a.TokenCount, &a.RefreshCount, &a.LastActivityAt, &a.LastTokenIssuedAt,
			&rotJSON,
		)
		if err != nil {
			return nil, err
		}
		if metaJSON != nil {
			_ = json.Unmarshal(metaJSON, &a.Metadata)
		}
		if rotJSON != nil {
			_ = json.Unmarshal(rotJSON, &a.RotationHistory)
		}
		if a.Tags == nil {
			a.Tags = []string{}
		}
		if a.Scopes == nil {
			a.Scopes = []string{}
		}
		agents = append(agents, a)
	}
	if agents == nil {
		agents = []Agent{}
	}
	return agents, nil
}

func (p *PostgresDB) CreateAgent(agent Agent) error {
	ctx := context.Background()
	metaJSON, _ := json.Marshal(agent.Metadata)
	rotJSON, _ := json.Marshal(agent.RotationHistory)
	if agent.Tags == nil {
		agent.Tags = []string{}
	}
	if agent.Scopes == nil {
		agent.Scopes = []string{}
	}
	_, err := p.pool.Exec(ctx, `INSERT INTO agents
		(id, organization_id, team_id, name, description, tags, metadata,
		 client_id, client_secret_hash, scopes, public_key, status, is_active,
		 created_at, updated_at, expires_at, token_count, refresh_count,
		 last_activity_at, last_token_issued_at, rotation_history)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		agent.ID, agent.OrganizationID, agent.TeamID,
		agent.Name, agent.Description, agent.Tags, metaJSON,
		agent.ClientID, agent.ClientSecretHash, agent.Scopes, agent.PublicKey,
		agent.Status, agent.IsActive, agent.CreatedAt, agent.UpdatedAt, agent.ExpiresAt,
		agent.TokenCount, agent.RefreshCount, agent.LastActivityAt, agent.LastTokenIssuedAt,
		rotJSON,
	)
	return err
}

func (p *PostgresDB) GetAgentByClientID(clientID string) (*Agent, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+agentCols+` FROM agents WHERE client_id = $1`, clientID)
	return scanAgent(row)
}

func (p *PostgresDB) GetAgentByID(id string) (*Agent, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+agentCols+` FROM agents WHERE id::text = $1`, id)
	a, err := scanAgent(row)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}
	return a, nil
}

func (p *PostgresDB) ListAgents() ([]Agent, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+agentCols+` FROM agents ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAgents(rows)
}

func (p *PostgresDB) ListAgentsByOrganization(orgID string) ([]Agent, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+agentCols+` FROM agents WHERE organization_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAgents(rows)
}

func (p *PostgresDB) ListAgentsByTeam(teamID string) ([]Agent, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+agentCols+` FROM agents WHERE team_id = $1 ORDER BY created_at DESC`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAgents(rows)
}

func (p *PostgresDB) UpdateAgent(id string, updateFn func(*Agent) error) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx,
		`SELECT `+agentCols+` FROM agents WHERE id::text = $1 FOR UPDATE`, id)
	agent, err := scanAgent(row)
	if err != nil {
		return fmt.Errorf("agent not found")
	}
	if err := updateFn(agent); err != nil {
		return err
	}
	agent.UpdatedAt = time.Now()
	metaJSON, _ := json.Marshal(agent.Metadata)
	rotJSON, _ := json.Marshal(agent.RotationHistory)
	if agent.Tags == nil {
		agent.Tags = []string{}
	}
	if agent.Scopes == nil {
		agent.Scopes = []string{}
	}
	_, err = tx.Exec(ctx, `UPDATE agents SET
		organization_id=$2, team_id=$3, name=$4, description=$5,
		tags=$6, metadata=$7, client_secret_hash=$8, scopes=$9,
		public_key=$10, status=$11, is_active=$12, updated_at=$13,
		expires_at=$14, token_count=$15, refresh_count=$16,
		last_activity_at=$17, last_token_issued_at=$18, rotation_history=$19
		WHERE id::text = $1`,
		id, agent.OrganizationID, agent.TeamID, agent.Name, agent.Description,
		agent.Tags, metaJSON, agent.ClientSecretHash, agent.Scopes,
		agent.PublicKey, agent.Status, agent.IsActive, agent.UpdatedAt,
		agent.ExpiresAt, agent.TokenCount, agent.RefreshCount,
		agent.LastActivityAt, agent.LastTokenIssuedAt, rotJSON,
	)
	if err != nil {
		return fmt.Errorf("update agent: %w", err)
	}
	return tx.Commit(ctx)
}

func (p *PostgresDB) DeleteAgent(id string) error {
	tag, err := p.pool.Exec(context.Background(),
		`DELETE FROM agents WHERE id::text = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("agent not found")
	}
	return nil
}

func (p *PostgresDB) CountAgents() (int, error) {
	var count int
	err := p.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM agents`).Scan(&count)
	return count, err
}

func (p *PostgresDB) ListAgentsPaginated(search, status, orgID, sort string, page, limit int) ([]Agent, int, error) {
	ctx := context.Background()
	var conds []string
	var args []interface{}
	idx := 1

	if search != "" {
		conds = append(conds, fmt.Sprintf(
			"(name ILIKE $%d OR client_id ILIKE $%d OR description ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}
	if status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, status)
		idx++
	}
	if orgID != "" {
		conds = append(conds, fmt.Sprintf("organization_id = $%d", idx))
		args = append(args, orgID)
		idx++
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	// Count.
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM agents "+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Sort.
	orderBy := "ORDER BY created_at DESC"
	if sort == "name" {
		orderBy = "ORDER BY LOWER(name) ASC"
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	q := fmt.Sprintf("SELECT %s FROM agents %s %s LIMIT $%d OFFSET $%d",
		agentCols, where, orderBy, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	agents, err := scanAgents(rows)
	if err != nil {
		return nil, 0, err
	}
	return agents, total, nil
}

// ═══════════════════════════════════════════════════════════════════════
// Audit Logs
// ═══════════════════════════════════════════════════════════════════════

func (p *PostgresDB) AddAuditLog(l AuditLog) error {
	_, err := p.pool.Exec(context.Background(), `INSERT INTO audit_logs
		(id, agent_id, action, ip_address, user_agent, details, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		l.ID, l.AgentID, l.Action, l.IPAddress, l.UserAgent, l.Details, l.CreatedAt,
	)
	return err
}

func (p *PostgresDB) ListAuditLogs(agentID, action, ipAddress string, from, to *time.Time, page, limit int) ([]AuditLog, int, error) {
	ctx := context.Background()
	var conds []string
	var args []interface{}
	idx := 1

	if agentID != "" {
		conds = append(conds, fmt.Sprintf("agent_id::text = $%d", idx))
		args = append(args, agentID)
		idx++
	}
	if action != "" {
		conds = append(conds, fmt.Sprintf("action = $%d", idx))
		args = append(args, action)
		idx++
	}
	if ipAddress != "" {
		conds = append(conds, fmt.Sprintf("ip_address = $%d", idx))
		args = append(args, ipAddress)
		idx++
	}
	if from != nil {
		conds = append(conds, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, *from)
		idx++
	}
	if to != nil {
		conds = append(conds, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, *to)
		idx++
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs "+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	q := fmt.Sprintf(`SELECT id::text, agent_id::text, action, ip_address, user_agent, details, created_at
		FROM audit_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.AgentID, &l.Action, &l.IPAddress, &l.UserAgent, &l.Details, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []AuditLog{}
	}
	return logs, total, nil
}

// ═══════════════════════════════════════════════════════════════════════
// Refresh Tokens
// ═══════════════════════════════════════════════════════════════════════

func (p *PostgresDB) CreateRefreshToken(rt RefreshToken) error {
	_, err := p.pool.Exec(context.Background(), `INSERT INTO refresh_tokens
		(id, agent_id, token_hash, expires_at, created_at, revoked_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		rt.ID, rt.AgentID, rt.TokenHash, rt.ExpiresAt, rt.CreatedAt, rt.RevokedAt,
	)
	return err
}

func (p *PostgresDB) GetRefreshToken(id string) (*RefreshToken, error) {
	var rt RefreshToken
	err := p.pool.QueryRow(context.Background(),
		`SELECT id::text, agent_id::text, token_hash, expires_at, created_at, revoked_at
		 FROM refresh_tokens WHERE id::text = $1`, id).
		Scan(&rt.ID, &rt.AgentID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found")
	}
	return &rt, nil
}

func (p *PostgresDB) RevokeRefreshToken(id string) error {
	tag, err := p.pool.Exec(context.Background(),
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE id::text = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found")
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════
// Revoked Tokens
// ═══════════════════════════════════════════════════════════════════════

func (p *PostgresDB) AddRevokedToken(rt RevokedToken) error {
	_, err := p.pool.Exec(context.Background(),
		`INSERT INTO revoked_tokens (jti, expires) VALUES ($1,$2)
		 ON CONFLICT (jti) DO NOTHING`,
		rt.JTI, rt.Expires)
	return err
}

func (p *PostgresDB) IsTokenRevoked(jti string) bool {
	var exists bool
	err := p.pool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE jti = $1)`, jti).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// ═══════════════════════════════════════════════════════════════════════
// Metrics
// ═══════════════════════════════════════════════════════════════════════

func (p *PostgresDB) IncrementTokensRefreshed() error {
	_, err := p.pool.Exec(context.Background(),
		`UPDATE metrics SET tokens_refreshed = tokens_refreshed + 1 WHERE id = 1`)
	return err
}

func (p *PostgresDB) IncrementTokensRevoked() error {
	_, err := p.pool.Exec(context.Background(),
		`UPDATE metrics SET tokens_revoked = tokens_revoked + 1 WHERE id = 1`)
	return err
}

func (p *PostgresDB) GetMetrics() Metrics {
	var m Metrics
	_ = p.pool.QueryRow(context.Background(),
		`SELECT tokens_refreshed, tokens_revoked FROM metrics WHERE id = 1`).
		Scan(&m.TokensRefreshed, &m.TokensRevoked)
	return m
}

// ═══════════════════════════════════════════════════════════════════════
// Organizations
// ═══════════════════════════════════════════════════════════════════════

const orgCols = `id::text, name, slug, owner_email, jwt_issuer, jwt_expiry_secs,
	allowed_origins, plan, created_at, updated_at`

func scanOrg(row pgx.Row) (*Organization, error) {
	var o Organization
	err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerEmail, &o.JWTIssuer,
		&o.JWTExpirySecs, &o.AllowedOrigins, &o.Plan, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (p *PostgresDB) CreateOrganization(org Organization) error {
	_, err := p.pool.Exec(context.Background(), `INSERT INTO organizations
		(id, name, slug, owner_email, jwt_issuer, jwt_expiry_secs,
		 allowed_origins, plan, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		org.ID, org.Name, org.Slug, org.OwnerEmail, org.JWTIssuer,
		org.JWTExpirySecs, org.AllowedOrigins, org.Plan, org.CreatedAt, org.UpdatedAt,
	)
	return err
}

func (p *PostgresDB) GetOrganization(id string) (*Organization, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+orgCols+` FROM organizations WHERE id::text = $1`, id)
	o, err := scanOrg(row)
	if err != nil {
		return nil, fmt.Errorf("organization not found")
	}
	return o, nil
}

func (p *PostgresDB) GetOrganizationBySlug(slug string) (*Organization, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+orgCols+` FROM organizations WHERE slug = $1`, slug)
	o, err := scanOrg(row)
	if err != nil {
		return nil, fmt.Errorf("organization not found")
	}
	return o, nil
}

func (p *PostgresDB) ListOrganizations() ([]Organization, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+orgCols+` FROM organizations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, *o)
	}
	if orgs == nil {
		orgs = []Organization{}
	}
	return orgs, nil
}

func (p *PostgresDB) UpdateOrganization(id string, updateFn func(*Organization) error) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT `+orgCols+` FROM organizations WHERE id::text = $1 FOR UPDATE`, id)
	org, err := scanOrg(row)
	if err != nil {
		return fmt.Errorf("organization not found")
	}
	if err := updateFn(org); err != nil {
		return err
	}
	org.UpdatedAt = time.Now()
	_, err = tx.Exec(ctx, `UPDATE organizations SET
		name=$2, slug=$3, owner_email=$4, jwt_issuer=$5, jwt_expiry_secs=$6,
		allowed_origins=$7, plan=$8, updated_at=$9
		WHERE id::text = $1`,
		id, org.Name, org.Slug, org.OwnerEmail, org.JWTIssuer,
		org.JWTExpirySecs, org.AllowedOrigins, org.Plan, org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	return tx.Commit(ctx)
}

func (p *PostgresDB) DeleteOrganization(id string) error {
	tag, err := p.pool.Exec(context.Background(),
		`DELETE FROM organizations WHERE id::text = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("organization not found")
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════
// Teams
// ═══════════════════════════════════════════════════════════════════════

const teamCols = `id::text, organization_id, name, description, created_at, updated_at`

func scanTeam(row pgx.Row) (*Team, error) {
	var t Team
	err := row.Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (p *PostgresDB) CreateTeam(team Team) error {
	_, err := p.pool.Exec(context.Background(), `INSERT INTO teams
		(id, organization_id, name, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		team.ID, team.OrganizationID, team.Name, team.Description, team.CreatedAt, team.UpdatedAt,
	)
	return err
}

func (p *PostgresDB) GetTeam(id string) (*Team, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+teamCols+` FROM teams WHERE id::text = $1`, id)
	t, err := scanTeam(row)
	if err != nil {
		return nil, fmt.Errorf("team not found")
	}
	return t, nil
}

func (p *PostgresDB) ListTeamsByOrganization(orgID string) ([]Team, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+teamCols+` FROM teams WHERE organization_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, *t)
	}
	return teams, nil
}

func (p *PostgresDB) UpdateTeam(id string, updateFn func(*Team) error) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT `+teamCols+` FROM teams WHERE id::text = $1 FOR UPDATE`, id)
	team, err := scanTeam(row)
	if err != nil {
		return fmt.Errorf("team not found")
	}
	if err := updateFn(team); err != nil {
		return err
	}
	team.UpdatedAt = time.Now()
	_, err = tx.Exec(ctx, `UPDATE teams SET name=$2, description=$3, updated_at=$4 WHERE id::text = $1`,
		id, team.Name, team.Description, team.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update team: %w", err)
	}
	return tx.Commit(ctx)
}

func (p *PostgresDB) DeleteTeam(id string) error {
	tag, err := p.pool.Exec(context.Background(),
		`DELETE FROM teams WHERE id::text = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════
// API Keys
// ═══════════════════════════════════════════════════════════════════════

const apiKeyCols = `id::text, organization_id, team_id, name, key_hash, prefix,
	last_used_at, expires_at, is_active, created_at`

func scanAPIKey(row pgx.Row) (*APIKey, error) {
	var k APIKey
	err := row.Scan(&k.ID, &k.OrganizationID, &k.TeamID, &k.Name, &k.KeyHash,
		&k.Prefix, &k.LastUsedAt, &k.ExpiresAt, &k.IsActive, &k.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (p *PostgresDB) CreateAPIKey(key APIKey) error {
	_, err := p.pool.Exec(context.Background(), `INSERT INTO api_keys
		(id, organization_id, team_id, name, key_hash, prefix,
		 last_used_at, expires_at, is_active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		key.ID, key.OrganizationID, key.TeamID, key.Name, key.KeyHash,
		key.Prefix, key.LastUsedAt, key.ExpiresAt, key.IsActive, key.CreatedAt,
	)
	return err
}

func (p *PostgresDB) GetAPIKeyByID(id string) (*APIKey, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+apiKeyCols+` FROM api_keys WHERE id::text = $1`, id)
	k, err := scanAPIKey(row)
	if err != nil {
		return nil, fmt.Errorf("API key not found")
	}
	return k, nil
}

func (p *PostgresDB) GetAPIKeyByKeyHash(keyHash string) (*APIKey, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+apiKeyCols+` FROM api_keys WHERE key_hash = $1 AND is_active = true`, keyHash)
	k, err := scanAPIKey(row)
	if err != nil {
		return nil, fmt.Errorf("API key not found")
	}
	return k, nil
}

func (p *PostgresDB) ListAPIKeysByOrganization(orgID string) ([]APIKey, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+apiKeyCols+` FROM api_keys WHERE organization_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, *k)
	}
	return keys, nil
}

func (p *PostgresDB) UpdateAPIKey(id string, updateFn func(*APIKey) error) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT `+apiKeyCols+` FROM api_keys WHERE id::text = $1 FOR UPDATE`, id)
	key, err := scanAPIKey(row)
	if err != nil {
		return fmt.Errorf("API key not found")
	}
	if err := updateFn(key); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE api_keys SET
		name=$2, key_hash=$3, last_used_at=$4, expires_at=$5, is_active=$6
		WHERE id::text = $1`,
		id, key.Name, key.KeyHash, key.LastUsedAt, key.ExpiresAt, key.IsActive)
	if err != nil {
		return fmt.Errorf("update api key: %w", err)
	}
	return tx.Commit(ctx)
}

func (p *PostgresDB) DeleteAPIKey(id string) error {
	tag, err := p.pool.Exec(context.Background(),
		`DELETE FROM api_keys WHERE id::text = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("API key not found")
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════
// Webhooks
// ═══════════════════════════════════════════════════════════════════════

const webhookCols = `id::text, organization_id, team_id, name, url, secret, events,
	is_active, max_retries, retry_backoff_base, created_at, updated_at,
	last_tested_at, consecutive_fails`

func scanWebhook(row pgx.Row) (*WebhookConfig, error) {
	var w WebhookConfig
	err := row.Scan(&w.ID, &w.OrganizationID, &w.TeamID, &w.Name, &w.URL, &w.Secret,
		&w.Events, &w.IsActive, &w.MaxRetries, &w.RetryBackoffBase,
		&w.CreatedAt, &w.UpdatedAt, &w.LastTestedAt, &w.ConsecutiveFails)
	if err != nil {
		return nil, err
	}
	if w.Events == nil {
		w.Events = []string{}
	}
	return &w, nil
}

func (p *PostgresDB) CreateWebhook(webhook WebhookConfig) error {
	if webhook.Events == nil {
		webhook.Events = []string{}
	}
	_, err := p.pool.Exec(context.Background(), `INSERT INTO webhook_configs
		(id, organization_id, team_id, name, url, secret, events,
		 is_active, max_retries, retry_backoff_base, created_at, updated_at,
		 last_tested_at, consecutive_fails)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		webhook.ID, webhook.OrganizationID, webhook.TeamID, webhook.Name,
		webhook.URL, webhook.Secret, webhook.Events, webhook.IsActive,
		webhook.MaxRetries, webhook.RetryBackoffBase, webhook.CreatedAt, webhook.UpdatedAt,
		webhook.LastTestedAt, webhook.ConsecutiveFails,
	)
	return err
}

func (p *PostgresDB) GetWebhook(id string) (*WebhookConfig, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+webhookCols+` FROM webhook_configs WHERE id::text = $1`, id)
	w, err := scanWebhook(row)
	if err != nil {
		return nil, fmt.Errorf("webhook not found")
	}
	return w, nil
}

func (p *PostgresDB) ListWebhooks() ([]WebhookConfig, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+webhookCols+` FROM webhook_configs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []WebhookConfig
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, *w)
	}
	if webhooks == nil {
		webhooks = []WebhookConfig{}
	}
	return webhooks, nil
}

func (p *PostgresDB) UpdateWebhook(id string, updateFn func(*WebhookConfig) error) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT `+webhookCols+` FROM webhook_configs WHERE id::text = $1 FOR UPDATE`, id)
	wh, err := scanWebhook(row)
	if err != nil {
		return fmt.Errorf("webhook not found")
	}
	if err := updateFn(wh); err != nil {
		return err
	}
	wh.UpdatedAt = time.Now()
	if wh.Events == nil {
		wh.Events = []string{}
	}
	_, err = tx.Exec(ctx, `UPDATE webhook_configs SET
		name=$2, url=$3, secret=$4, events=$5, is_active=$6,
		max_retries=$7, retry_backoff_base=$8, updated_at=$9,
		last_tested_at=$10, consecutive_fails=$11
		WHERE id::text = $1`,
		id, wh.Name, wh.URL, wh.Secret, wh.Events, wh.IsActive,
		wh.MaxRetries, wh.RetryBackoffBase, wh.UpdatedAt,
		wh.LastTestedAt, wh.ConsecutiveFails,
	)
	if err != nil {
		return fmt.Errorf("update webhook: %w", err)
	}
	return tx.Commit(ctx)
}

func (p *PostgresDB) DeleteWebhook(id string) error {
	tag, err := p.pool.Exec(context.Background(),
		`DELETE FROM webhook_configs WHERE id::text = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

func (p *PostgresDB) ListActiveWebhooksForEvent(event string) ([]WebhookConfig, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+webhookCols+` FROM webhook_configs
		 WHERE is_active = true AND ($1 = ANY(events) OR '*' = ANY(events))`, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []WebhookConfig
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *w)
	}
	return result, nil
}

// ═══════════════════════════════════════════════════════════════════════
// Webhook Deliveries
// ═══════════════════════════════════════════════════════════════════════

const deliveryCols = `id::text, webhook_config_id::text, event, payload, headers,
	status, attempts, last_attempt_at, last_error, next_retry_at, created_at`

func scanDelivery(row pgx.Row) (*WebhookDelivery, error) {
	var d WebhookDelivery
	err := row.Scan(&d.ID, &d.WebhookConfigID, &d.Event, &d.Payload, &d.Headers,
		&d.Status, &d.Attempts, &d.LastAttemptAt, &d.LastError, &d.NextRetryAt, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (p *PostgresDB) AddWebhookDelivery(delivery WebhookDelivery) error {
	_, err := p.pool.Exec(context.Background(), `INSERT INTO webhook_deliveries
		(id, webhook_config_id, event, payload, headers, status, attempts,
		 last_attempt_at, last_error, next_retry_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		delivery.ID, delivery.WebhookConfigID, delivery.Event,
		delivery.Payload, delivery.Headers, delivery.Status, delivery.Attempts,
		delivery.LastAttemptAt, delivery.LastError, delivery.NextRetryAt, delivery.CreatedAt,
	)
	return err
}

func (p *PostgresDB) GetWebhookDelivery(id string) (*WebhookDelivery, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT `+deliveryCols+` FROM webhook_deliveries WHERE id::text = $1`, id)
	d, err := scanDelivery(row)
	if err != nil {
		return nil, fmt.Errorf("webhook delivery not found")
	}
	return d, nil
}

func (p *PostgresDB) UpdateWebhookDelivery(id string, updateFn func(*WebhookDelivery) error) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT `+deliveryCols+` FROM webhook_deliveries WHERE id::text = $1 FOR UPDATE`, id)
	d, err := scanDelivery(row)
	if err != nil {
		return fmt.Errorf("webhook delivery not found")
	}
	if err := updateFn(d); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE webhook_deliveries SET
		status=$2, attempts=$3, last_attempt_at=$4, last_error=$5, next_retry_at=$6
		WHERE id::text = $1`,
		id, d.Status, d.Attempts, d.LastAttemptAt, d.LastError, d.NextRetryAt)
	if err != nil {
		return fmt.Errorf("update delivery: %w", err)
	}
	return tx.Commit(ctx)
}

func (p *PostgresDB) ListWebhookDeliveries(webhookConfigID string) ([]WebhookDelivery, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+deliveryCols+` FROM webhook_deliveries
		 WHERE webhook_config_id::text = $1 ORDER BY created_at DESC`, webhookConfigID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []WebhookDelivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, nil
}

func (p *PostgresDB) ListPendingDeliveries() ([]WebhookDelivery, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+deliveryCols+` FROM webhook_deliveries
		 WHERE status = 'pending'
		    OR (status = 'retrying' AND next_retry_at <= NOW())
		 ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []WebhookDelivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, nil
}

// ═══════════════════════════════════════════════════════════════════════
// Admin Users
// ═══════════════════════════════════════════════════════════════════════

const adminCols = `id::text, email, password_hash, role, created_at, updated_at`

func (p *PostgresDB) CreateAdminUser(user AdminUser) error {
	_, err := p.pool.Exec(context.Background(), `INSERT INTO admin_users
		(id, email, password_hash, role, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (p *PostgresDB) GetAdminUserByEmail(email string) (*AdminUser, error) {
	var u AdminUser
	err := p.pool.QueryRow(context.Background(),
		`SELECT `+adminCols+` FROM admin_users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("admin user not found")
	}
	return &u, nil
}

func (p *PostgresDB) GetAdminUserByID(id string) (*AdminUser, error) {
	var u AdminUser
	err := p.pool.QueryRow(context.Background(),
		`SELECT `+adminCols+` FROM admin_users WHERE id::text = $1`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("admin user not found")
	}
	return &u, nil
}

func (p *PostgresDB) ListAdminUsers() ([]AdminUser, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT `+adminCols+` FROM admin_users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []AdminUser{}
	}
	return users, nil
}
