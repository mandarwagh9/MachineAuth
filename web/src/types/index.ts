export interface Agent {
  id: string
  name: string
  client_id: string
  scopes: string[]
  public_key?: string
  is_active: boolean
  created_at: string
  updated_at: string
  expires_at?: string
}

export interface CreateAgentRequest {
  name: string
  scopes?: string[]
  expires_in?: number
}

export interface CreateAgentResponse {
  agent: Agent
  client_secret: string
  client_id: string
}

export interface AgentsListResponse {
  agents: Agent[]
}

export interface TokenRequest {
  grant_type: string
  client_id: string
  client_secret: string
  scope?: string
}

export interface TokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  scope?: string
  issued_at: number
}

export interface ErrorResponse {
  error: string
  error_description?: string
}

export interface AuditLog {
  id: string
  agent_id?: string
  action: string
  ip_address?: string
  user_agent?: string
  created_at: string
}
