import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MachineAuthClient, MachineAuthError } from './index';
import type {
  TokenResponse,
  Agent,
  Organization,
  Team,
  APIKey,
  WebhookConfig,
  WebhookDelivery,
} from './index';

function mockFetch(body: unknown, status = 200): typeof fetch {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    text: () => Promise.resolve(body != null ? JSON.stringify(body) : ''),
  }) as unknown as typeof fetch;
}

function client(fetchImpl: typeof fetch, token?: string): MachineAuthClient {
  return new MachineAuthClient({
    baseUrl: 'http://test:8081',
    clientId: 'cid',
    clientSecret: 'csec',
    defaultToken: token,
    fetchImpl,
  });
}

// ── Token operations ────────────────────────────────────────────

describe('Token operations', () => {
  it('clientCredentialsToken sets default token', async () => {
    const f = mockFetch({ access_token: 'tok-1', token_type: 'Bearer', expires_in: 3600 });
    const c = client(f);
    const tok = await c.clientCredentialsToken();
    expect(tok.access_token).toBe('tok-1');
    expect(f).toHaveBeenCalledOnce();
  });

  it('refreshToken sets default token', async () => {
    const f = mockFetch({ access_token: 'tok-2', token_type: 'Bearer' });
    const c = client(f);
    const tok = await c.refreshToken({ refreshToken: 'rt-1' });
    expect(tok.access_token).toBe('tok-2');
  });

  it('introspect returns response', async () => {
    const f = mockFetch({ active: true, scope: 'read' });
    const result = await client(f).introspect('tok');
    expect(result.active).toBe(true);
  });

  it('revoke calls API', async () => {
    const f = mockFetch({ status: 'revoked' });
    await client(f).revoke('tok');
    expect(f).toHaveBeenCalledOnce();
  });

  it('throws without credentials', async () => {
    const c = new MachineAuthClient({ fetchImpl: mockFetch({}) });
    await expect(c.clientCredentialsToken()).rejects.toThrow('clientId and clientSecret');
  });
});

// ── Agent management ────────────────────────────────────────────

describe('Agent management', () => {
  const agentData = { id: 'a-1', name: 'test', client_id: 'c', scopes: ['read'] };

  it('listAgents returns Agent[]', async () => {
    const f = mockFetch({ agents: [agentData] });
    const agents = await client(f, 'tok').listAgents();
    expect(agents).toHaveLength(1);
    expect(agents[0].id).toBe('a-1');
  });

  it('createAgent sends body', async () => {
    const f = mockFetch({ agent: agentData, client_secret: 'sec' });
    const result = await client(f, 'tok').createAgent({ name: 'test' });
    expect(result).toHaveProperty('client_secret', 'sec');
  });

  it('getAgent returns Agent', async () => {
    const f = mockFetch({ agent: agentData });
    const agent = await client(f, 'tok').getAgent('a-1');
    expect(agent.id).toBe('a-1');
  });

  it('deleteAgent calls API', async () => {
    const f = mockFetch(null, 204);
    await client(f, 'tok').deleteAgent('a-1');
    expect(f).toHaveBeenCalledOnce();
  });

  it('rotateAgent returns secret', async () => {
    const f = mockFetch({ client_secret: 'new-sec' });
    const result = await client(f, 'tok').rotateAgent('a-1');
    expect(result).toHaveProperty('client_secret', 'new-sec');
  });
});

// ── Self-service ────────────────────────────────────────────────

describe('Self-service', () => {
  it('getMe returns Agent', async () => {
    const f = mockFetch({ agent: { id: 'me', name: 'self', client_id: 'c', scopes: [] } });
    const agent = await client(f, 'tok').getMe();
    expect(agent.id).toBe('me');
  });

  it('getMyUsage returns usage data', async () => {
    const f = mockFetch({ agent: { id: 'me', name: 'self', client_id: 'c' }, token_count: 42 });
    const usage = await client(f, 'tok').getMyUsage();
    expect(usage.token_count).toBe(42);
  });

  it('rotateMe returns secret', async () => {
    const f = mockFetch({ client_secret: 'rotated' });
    const result = await client(f, 'tok').rotateMe();
    expect(result).toHaveProperty('client_secret', 'rotated');
  });

  it('deactivateMe returns message', async () => {
    const f = mockFetch({ message: 'deactivated' });
    const result = await client(f, 'tok').deactivateMe();
    expect(result).toHaveProperty('message');
  });

  it('reactivateMe returns message', async () => {
    const f = mockFetch({ message: 'reactivated' });
    const result = await client(f, 'tok').reactivateMe();
    expect(result).toHaveProperty('message');
  });

  it('deleteMe calls API', async () => {
    const f = mockFetch(null, 204);
    await client(f, 'tok').deleteMe();
    expect(f).toHaveBeenCalledOnce();
  });
});

// ── Organizations ───────────────────────────────────────────────

describe('Organizations', () => {
  it('listOrganizations returns array', async () => {
    const f = mockFetch({ organizations: [{ id: 'o1', name: 'Org', slug: 'org' }] });
    const orgs = await client(f, 'tok').listOrganizations();
    expect(orgs).toHaveLength(1);
  });

  it('createOrganization returns org', async () => {
    const f = mockFetch({ id: 'o1', name: 'Org', slug: 'org' });
    const org = await client(f, 'tok').createOrganization({ name: 'Org', slug: 'org', owner_email: 'a@b.com' });
    expect(org.id).toBe('o1');
  });

  it('updateOrganization returns updated org', async () => {
    const f = mockFetch({ id: 'o1', name: 'Updated', slug: 'org' });
    const org = await client(f, 'tok').updateOrganization('o1', { name: 'Updated' });
    expect(org.name).toBe('Updated');
  });

  it('deleteOrganization calls API', async () => {
    const f = mockFetch(null, 204);
    await client(f, 'tok').deleteOrganization('o1');
    expect(f).toHaveBeenCalledOnce();
  });
});

// ── Teams ───────────────────────────────────────────────────────

describe('Teams', () => {
  it('listTeams returns array', async () => {
    const f = mockFetch({ teams: [{ id: 't1', name: 'T', organization_id: 'o1' }] });
    const teams = await client(f, 'tok').listTeams('o1');
    expect(teams).toHaveLength(1);
  });

  it('createTeam returns team', async () => {
    const f = mockFetch({ id: 't1', name: 'T', organization_id: 'o1' });
    const team = await client(f, 'tok').createTeam('o1', { name: 'T' });
    expect(team.id).toBe('t1');
  });
});

// ── API Keys ────────────────────────────────────────────────────

describe('API Keys', () => {
  it('listAPIKeys returns array', async () => {
    const f = mockFetch({ api_keys: [{ id: 'k1', name: 'K', organization_id: 'o1' }] });
    const keys = await client(f, 'tok').listAPIKeys('o1');
    expect(keys).toHaveLength(1);
  });

  it('createAPIKey returns key', async () => {
    const f = mockFetch({ api_key: { id: 'k1' }, key: 'raw-key' });
    const result = await client(f, 'tok').createAPIKey('o1', { name: 'K' });
    expect(result).toHaveProperty('key', 'raw-key');
  });

  it('deleteAPIKey calls API', async () => {
    const f = mockFetch(null, 204);
    await client(f, 'tok').deleteAPIKey('o1', 'k1');
    expect(f).toHaveBeenCalledOnce();
  });
});

// ── Webhooks ────────────────────────────────────────────────────

describe('Webhooks', () => {
  it('listWebhooks returns array', async () => {
    const f = mockFetch({ webhooks: [{ id: 'w1', name: 'W', url: 'http://x', events: ['a'] }] });
    const hooks = await client(f, 'tok').listWebhooks();
    expect(hooks).toHaveLength(1);
  });

  it('createWebhook returns with secret', async () => {
    const f = mockFetch({ webhook: { id: 'w1' }, secret: 'sec' });
    const result = await client(f, 'tok').createWebhook({ name: 'W', url: 'http://x', events: ['a'] });
    expect(result).toHaveProperty('secret', 'sec');
  });

  it('getWebhook returns config', async () => {
    const f = mockFetch({ webhook: { id: 'w1', name: 'W', url: 'http://x', events: ['a'] } });
    const hook = await client(f, 'tok').getWebhook('w1');
    expect(hook.id).toBe('w1');
  });

  it('updateWebhook returns updated config', async () => {
    const f = mockFetch({ webhook: { id: 'w1', name: 'Updated', url: 'http://x', events: ['a'] } });
    const hook = await client(f, 'tok').updateWebhook('w1', { name: 'Updated' });
    expect(hook.name).toBe('Updated');
  });

  it('deleteWebhook calls API', async () => {
    const f = mockFetch(null, 204);
    await client(f, 'tok').deleteWebhook('w1');
    expect(f).toHaveBeenCalledOnce();
  });

  it('testWebhook returns result', async () => {
    const f = mockFetch({ success: true, status_code: 200 });
    const result = await client(f, 'tok').testWebhook('w1', 'agent.created');
    expect(result.success).toBe(true);
  });

  it('listDeliveries returns array', async () => {
    const f = mockFetch({ deliveries: [{ id: 'd1', webhook_config_id: 'w1', event: 'a', status: 'delivered' }] });
    const deliveries = await client(f, 'tok').listDeliveries('w1');
    expect(deliveries).toHaveLength(1);
  });

  it('listWebhookEvents returns string[]', async () => {
    const f = mockFetch({ events: ['agent.created', 'agent.deleted'] });
    const events = await client(f, 'tok').listWebhookEvents();
    expect(events).toEqual(['agent.created', 'agent.deleted']);
  });
});

// ── Error handling ──────────────────────────────────────────────

describe('Error handling', () => {
  it('throws MachineAuthError on non-OK response', async () => {
    const f = mockFetch({ message: 'not found' }, 404);
    await expect(client(f, 'tok').getAgent('missing')).rejects.toThrow(MachineAuthError);
  });

  it('error has status and body', async () => {
    const f = mockFetch({ message: 'unauthorized' }, 401);
    try {
      await client(f, 'tok').getMe();
      expect.fail('should throw');
    } catch (e) {
      expect(e).toBeInstanceOf(MachineAuthError);
      expect((e as MachineAuthError).status).toBe(401);
    }
  });
});
