export interface Agent {
  id: string;
  name: string;
  description?: string;
  client_id: string;
  scopes: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_seen_at?: string;
  expires_at?: string;
  token_count?: number;
  refresh_count?: number;
  last_activity_at?: string;
  last_token_issued_at?: string;
}

export interface Rotation {
  rotated_at: string;
  rotated_by_ip?: string;
}

export interface AgentUsage {
  agent: Agent;
  token_count: number;
  refresh_count: number;
  last_activity_at?: string;
  last_token_issued_at?: string;
  rotation_history: Rotation[];
}

export interface CreateAgentRequest {
  name: string;
  description?: string;
  scopes: string[];
}

export interface CreateAgentResponse {
  agent: Agent;
  client_id: string;
  client_secret: string;
  message: string;
}

export interface AgentsListResponse {
  agents: Agent[];
  total: number;
}

export interface TokenRequest {
  grant_type: string;
  client_id: string;
  client_secret: string;
}

export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  scope?: string;
}

export interface Metrics {
  requests: number;
  tokens_issued: number;
  tokens_refreshed: number;
  tokens_revoked: number;
  active_tokens: number;
  total_agents: number;
}

export interface HealthCheck {
  status: string;
  timestamp: string;
  agents_count?: number;
  active_tokens?: number;
}

export interface ErrorResponse {
  error: string;
  error_description?: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  owner_email?: string;
  created_at: string;
  updated_at?: string;
}

export interface CreateOrganizationRequest {
  name: string;
  slug: string;
  owner_email?: string;
}

export interface OrganizationsListResponse {
  organizations: Organization[];
  total: number;
}

export interface Team {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  created_at: string;
}

export interface CreateTeamRequest {
  name: string;
  description?: string;
}

export interface TeamsListResponse {
  teams: Team[];
}

export interface APIKey {
  id: string;
  organization_id: string;
  team_id?: string;
  name: string;
  prefix: string;
  is_active: boolean;
  expires_at?: string;
  last_used_at?: string;
  created_at: string;
}

export interface CreateAPIKeyRequest {
  name: string;
  expires_in?: number;
  team_id?: string;
}

export interface CreateAPIKeyResponse {
  api_key: APIKey;
  key: string;
}

export interface APIKeysListResponse {
  api_keys: APIKey[];
}
