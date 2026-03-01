// ── Configuration ────────────────────────────────────────────────

export interface ClientOptions {
  baseUrl?: string;
  clientId?: string;
  clientSecret?: string;
  defaultToken?: string;
  fetchImpl?: typeof fetch;
  timeoutMs?: number;
  maxRetries?: number;
  retryBaseMs?: number;
  onRequest?: (url: string, init: RequestInit) => void;
  onResponse?: (url: string, status: number, body: unknown) => void;
}

// ── Token types ─────────────────────────────────────────────────

export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in?: number;
  refresh_token?: string;
  scope?: string;
  issued_at?: string;
}

export interface IntrospectionResponse {
  active: boolean;
  scope?: string;
  client_id?: string;
  token_type?: string;
  exp?: number;
  iat?: number;
  nbf?: number;
  sub?: string;
  aud?: string | string[];
  iss?: string;
  jti?: string;
}

// ── Agent types ─────────────────────────────────────────────────

export interface Agent {
  id: string;
  name: string;
  scopes: string[];
  client_id: string;
  organization_id?: string;
  team_id?: string;
  public_key?: string;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
  expires_at?: string | null;
  token_count?: number;
  refresh_count?: number;
  last_activity_at?: string;
}

export interface CreateAgentRequest {
  name: string;
  scopes?: string[];
  organization_id?: string;
  team_id?: string;
  expires_in?: number;
}

export interface AgentUsage {
  agent: Agent;
  organization_id?: string;
  token_count: number;
  refresh_count: number;
  rotation_history: Record<string, unknown>[];
}

// ── Organization types ──────────────────────────────────────────

export interface Organization {
  id: string;
  name: string;
  slug: string;
  owner_email?: string;
  jwt_issuer?: string;
  jwt_expiry_secs?: number;
  allowed_origins?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface CreateOrganizationRequest {
  name: string;
  slug: string;
  owner_email: string;
}

export interface UpdateOrganizationRequest {
  name?: string;
  jwt_issuer?: string;
  jwt_expiry_secs?: number;
  allowed_origins?: string[];
}

// ── Team types ──────────────────────────────────────────────────

export interface Team {
  id: string;
  name: string;
  organization_id: string;
  description?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateTeamRequest {
  name: string;
  description?: string;
}

// ── API Key types ───────────────────────────────────────────────

export interface APIKey {
  id: string;
  name: string;
  organization_id: string;
  key_prefix?: string;
  team_id?: string;
  is_active?: boolean;
  created_at?: string;
  expires_at?: string;
}

export interface CreateAPIKeyRequest {
  name: string;
  team_id?: string;
  expires_in?: number;
}

// ── Webhook types ───────────────────────────────────────────────

export interface WebhookConfig {
  id: string;
  name: string;
  url: string;
  events: string[];
  secret?: string;
  organization_id?: string;
  team_id?: string;
  is_active?: boolean;
  max_retries?: number;
  retry_backoff_base?: number;
  created_at?: string;
  updated_at?: string;
  last_tested_at?: string;
  consecutive_fails?: number;
}

export interface CreateWebhookRequest {
  name: string;
  url: string;
  events: string[];
  max_retries?: number;
  retry_backoff_base?: number;
}

export interface UpdateWebhookRequest {
  name?: string;
  url?: string;
  events?: string[];
  is_active?: boolean;
  max_retries?: number;
  retry_backoff_base?: number;
}

export interface WebhookDelivery {
  id: string;
  webhook_config_id: string;
  event: string;
  status: string;
  payload?: string;
  headers?: string;
  attempts?: number;
  last_attempt_at?: string;
  last_error?: string;
  next_retry_at?: string;
  created_at?: string;
}

export interface WebhookTestResult {
  success: boolean;
  status_code?: number;
  error?: string;
}

// ── Error ───────────────────────────────────────────────────────

const DEFAULT_TIMEOUT_MS = 10_000;

export class MachineAuthError extends Error {
  status: number;
  body: unknown;

  constructor(status: number, message: string, body?: unknown) {
    super(message);
    this.name = 'MachineAuthError';
    this.status = status;
    this.body = body;
  }
}

// ── Client ──────────────────────────────────────────────────────

export class MachineAuthClient {
  private baseUrl: string;
  private clientId?: string;
  private clientSecret?: string;
  private defaultToken?: string;
  private fetchImpl?: typeof fetch;
  private timeoutMs: number;
  private maxRetries: number;
  private retryBaseMs: number;
  private onRequest?: (url: string, init: RequestInit) => void;
  private onResponse?: (url: string, status: number, body: unknown) => void;

  constructor(options: ClientOptions = {}) {
    this.baseUrl = (options.baseUrl ?? 'http://localhost:8081').replace(/\/+$/, '');
    this.clientId = options.clientId;
    this.clientSecret = options.clientSecret;
    this.defaultToken = options.defaultToken;
    this.fetchImpl = options.fetchImpl;
    this.timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS;
    this.maxRetries = options.maxRetries ?? 0;
    this.retryBaseMs = options.retryBaseMs ?? 500;
    this.onRequest = options.onRequest;
    this.onResponse = options.onResponse;
  }

  /** Create a client from MACHINEAUTH_* environment variables (Node.js). */
  static fromEnv(overrides: Partial<ClientOptions> = {}): MachineAuthClient {
    const env = typeof process !== 'undefined' ? process.env : {};
    return new MachineAuthClient({
      baseUrl: env.MACHINEAUTH_BASE_URL ?? 'http://localhost:8081',
      clientId: env.MACHINEAUTH_CLIENT_ID,
      clientSecret: env.MACHINEAUTH_CLIENT_SECRET,
      defaultToken: env.MACHINEAUTH_ACCESS_TOKEN,
      ...overrides,
    });
  }

  withToken(token: string): this {
    this.defaultToken = token;
    return this;
  }

  // ── Token operations ────────────────────────────────────────

  async clientCredentialsToken(params?: {
    clientId?: string;
    clientSecret?: string;
    scope?: string[];
  }): Promise<TokenResponse> {
    const clientId = params?.clientId ?? this.clientId;
    const clientSecret = params?.clientSecret ?? this.clientSecret;
    if (!clientId || !clientSecret) {
      throw new Error('clientId and clientSecret are required');
    }
    const form = new URLSearchParams({
      grant_type: 'client_credentials',
      client_id: clientId,
      client_secret: clientSecret,
    });
    if (params?.scope?.length) {
      form.set('scope', params.scope.join(' '));
    }
    const token = await this.request<TokenResponse>('/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: form.toString(),
    });
    this.defaultToken = token.access_token;
    return token;
  }

  async refreshToken(params: {
    refreshToken: string;
    clientId?: string;
    clientSecret?: string;
  }): Promise<TokenResponse> {
    const form = new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: params.refreshToken,
    });
    if (params.clientId ?? this.clientId) {
      form.set('client_id', (params.clientId ?? this.clientId)!);
    }
    if (params.clientSecret ?? this.clientSecret) {
      form.set('client_secret', (params.clientSecret ?? this.clientSecret)!);
    }
    const token = await this.request<TokenResponse>('/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: form.toString(),
    });
    this.defaultToken = token.access_token;
    return token;
  }

  async introspect(token: string): Promise<IntrospectionResponse> {
    const form = new URLSearchParams({ token });
    return this.request<IntrospectionResponse>('/oauth/introspect', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: form.toString(),
    });
  }

  async revoke(token: string, tokenTypeHint?: string): Promise<void> {
    const form = new URLSearchParams({ token });
    if (tokenTypeHint) form.set('token_type_hint', tokenTypeHint);
    await this.request<void>('/oauth/revoke', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: form.toString(),
    });
  }

  // ── Agent management ────────────────────────────────────────

  async listAgents(token?: string): Promise<Agent[]> {
    const raw = await this.request<{ agents: Agent[] }>('/api/agents', {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).agents ?? raw;
  }

  async createAgent(body: CreateAgentRequest, token?: string): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('/api/agents', {
      method: 'POST',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  async getAgent(id: string, token?: string): Promise<Agent> {
    const raw = await this.request<{ agent: Agent }>(`/api/agents/${id}`, {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).agent ?? raw;
  }

  async deleteAgent(id: string, token?: string): Promise<void> {
    await this.request<void>(`/api/agents/${id}`, {
      method: 'DELETE',
      headers: this.authHeaders(token),
    });
  }

  async rotateAgent(id: string, token?: string): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>(`/api/agents/${id}/rotate`, {
      method: 'POST',
      headers: this.authHeaders(token),
    });
  }

  // ── Self-service ────────────────────────────────────────────

  async getMe(token?: string): Promise<Agent> {
    const raw = await this.request<{ agent: Agent }>('/api/agents/me', {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).agent ?? raw;
  }

  async getMyUsage(token?: string): Promise<AgentUsage> {
    return this.request<AgentUsage>('/api/agents/me/usage', {
      method: 'GET',
      headers: this.authHeaders(token),
    });
  }

  async rotateMe(token?: string): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('/api/agents/me/rotate', {
      method: 'POST',
      headers: this.authHeaders(token),
    });
  }

  async deactivateMe(token?: string): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('/api/agents/me/deactivate', {
      method: 'POST',
      headers: this.authHeaders(token),
    });
  }

  async reactivateMe(token?: string): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('/api/agents/me/reactivate', {
      method: 'POST',
      headers: this.authHeaders(token),
    });
  }

  async deleteMe(token?: string): Promise<void> {
    await this.request<void>('/api/agents/me', {
      method: 'DELETE',
      headers: this.authHeaders(token),
    });
  }

  // ── Organizations ───────────────────────────────────────────

  async listOrganizations(token?: string): Promise<Organization[]> {
    const raw = await this.request<{ organizations: Organization[] }>('/api/organizations', {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).organizations ?? raw;
  }

  async createOrganization(body: CreateOrganizationRequest, token?: string): Promise<Organization> {
    return this.request<Organization>('/api/organizations', {
      method: 'POST',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  async getOrganization(id: string, token?: string): Promise<Organization> {
    return this.request<Organization>(`/api/organizations/${id}`, {
      method: 'GET',
      headers: this.authHeaders(token),
    });
  }

  async updateOrganization(id: string, body: UpdateOrganizationRequest, token?: string): Promise<Organization> {
    return this.request<Organization>(`/api/organizations/${id}`, {
      method: 'PUT',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  async deleteOrganization(id: string, token?: string): Promise<void> {
    await this.request<void>(`/api/organizations/${id}`, {
      method: 'DELETE',
      headers: this.authHeaders(token),
    });
  }

  // ── Teams ───────────────────────────────────────────────────

  async listTeams(orgId: string, token?: string): Promise<Team[]> {
    const raw = await this.request<{ teams: Team[] }>(`/api/organizations/${orgId}/teams`, {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).teams ?? raw;
  }

  async createTeam(orgId: string, body: CreateTeamRequest, token?: string): Promise<Team> {
    return this.request<Team>(`/api/organizations/${orgId}/teams`, {
      method: 'POST',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  // ── API Keys ────────────────────────────────────────────────

  async listAPIKeys(orgId: string, token?: string): Promise<APIKey[]> {
    const raw = await this.request<{ api_keys: APIKey[] }>(`/api/organizations/${orgId}/api-keys`, {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).api_keys ?? raw;
  }

  async createAPIKey(orgId: string, body: CreateAPIKeyRequest, token?: string): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>(`/api/organizations/${orgId}/api-keys`, {
      method: 'POST',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  async deleteAPIKey(orgId: string, keyId: string, token?: string): Promise<void> {
    await this.request<void>(`/api/organizations/${orgId}/api-keys/${keyId}`, {
      method: 'DELETE',
      headers: this.authHeaders(token),
    });
  }

  // ── Webhooks ────────────────────────────────────────────────

  async listWebhooks(token?: string): Promise<WebhookConfig[]> {
    const raw = await this.request<{ webhooks: WebhookConfig[] }>('/api/webhooks', {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).webhooks ?? raw;
  }

  async createWebhook(body: CreateWebhookRequest, token?: string): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('/api/webhooks', {
      method: 'POST',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  async getWebhook(id: string, token?: string): Promise<WebhookConfig> {
    const raw = await this.request<{ webhook: WebhookConfig }>(`/api/webhooks/${id}`, {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).webhook ?? raw;
  }

  async updateWebhook(id: string, body: UpdateWebhookRequest, token?: string): Promise<WebhookConfig> {
    const raw = await this.request<{ webhook: WebhookConfig }>(`/api/webhooks/${id}`, {
      method: 'PUT',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    return (raw as any).webhook ?? raw;
  }

  async deleteWebhook(id: string, token?: string): Promise<void> {
    await this.request<void>(`/api/webhooks/${id}`, {
      method: 'DELETE',
      headers: this.authHeaders(token),
    });
  }

  async testWebhook(id: string, event: string, payload?: Record<string, unknown>, token?: string): Promise<WebhookTestResult> {
    const body: Record<string, unknown> = { event };
    if (payload) body.payload = payload;
    return this.request<WebhookTestResult>(`/api/webhooks/${id}/test`, {
      method: 'POST',
      headers: { ...this.authHeaders(token), 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  async listDeliveries(webhookId: string, token?: string): Promise<WebhookDelivery[]> {
    const raw = await this.request<{ deliveries: WebhookDelivery[] }>(`/api/webhooks/${webhookId}/deliveries`, {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).deliveries ?? raw;
  }

  async getDelivery(webhookId: string, deliveryId: string, token?: string): Promise<WebhookDelivery> {
    const raw = await this.request<{ delivery: WebhookDelivery }>(`/api/webhooks/${webhookId}/deliveries/${deliveryId}`, {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).delivery ?? raw;
  }

  async listWebhookEvents(token?: string): Promise<string[]> {
    const raw = await this.request<{ events: string[] }>('/api/webhook-events', {
      method: 'GET',
      headers: this.authHeaders(token),
    });
    return (raw as any).events ?? [];
  }

  // ── Internal helpers ────────────────────────────────────────

  private authHeaders(token?: string): Record<string, string> {
    const bearer = token ?? this.defaultToken;
    return bearer ? { Authorization: `Bearer ${bearer}` } : {};
  }

  private async request<T>(path: string, init: RequestInit): Promise<T> {
    const fetchFn = this.fetchImpl ?? (globalThis as any).fetch;
    if (!fetchFn) {
      throw new Error('fetch is not available in this environment; provide fetchImpl');
    }
    const url = `${this.baseUrl}${path}`;
    let lastError: unknown;

    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      if (attempt > 0) {
        const delay = this.retryBaseMs * Math.pow(2, attempt - 1);
        await new Promise(r => setTimeout(r, delay));
      }

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.timeoutMs);
      try {
        this.onRequest?.(url, init);
        const res = await fetchFn(url, {
          ...init,
          signal: controller.signal,
        });
        const text = await res.text();
        let parsed: unknown = undefined;
        if (text) {
          try {
            parsed = JSON.parse(text);
          } catch {
            parsed = text;
          }
        }
        this.onResponse?.(url, res.status, parsed);
        if (!res.ok) {
          const message = typeof parsed === 'string' ? parsed : (parsed as any)?.message ?? res.statusText;
          const err = new MachineAuthError(res.status, message, parsed);
          // Only retry on 429 or 5xx
          if (attempt < this.maxRetries && (res.status === 429 || res.status >= 500)) {
            lastError = err;
            continue;
          }
          throw err;
        }
        return parsed as T;
      } catch (err: any) {
        if (err instanceof MachineAuthError) throw err;
        if (err?.name === 'AbortError') {
          const timeoutErr = new MachineAuthError(408, 'Request timeout');
          if (attempt < this.maxRetries) {
            lastError = timeoutErr;
            continue;
          }
          throw timeoutErr;
        }
        throw err;
      } finally {
        clearTimeout(timeoutId);
      }
    }
    throw lastError;
  }
}
