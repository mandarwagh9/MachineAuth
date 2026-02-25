import axios from 'axios'
import type {
  Agent,
  CreateAgentRequest,
  CreateAgentResponse,
  AgentsListResponse,
  TokenRequest,
  TokenResponse,
  Metrics,
  HealthCheck,
} from '@/types'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

export const AgentService = {
  list: async (): Promise<AgentsListResponse> => {
    const response = await api.get<AgentsListResponse>('/agents')
    return response.data
  },

  get: async (id: string): Promise<Agent> => {
    const response = await api.get<{ agent: Agent }>(`/agents/${id}`)
    return response.data.agent
  },

  create: async (data: CreateAgentRequest): Promise<CreateAgentResponse> => {
    const response = await api.post<CreateAgentResponse>('/agents', data)
    return response.data
  },

  rotate: async (id: string): Promise<{ client_id: string; client_secret: string; message: string }> => {
    const response = await api.post<{ client_id: string; client_secret: string; message: string }>(`/agents/${id}/rotate`)
    return response.data
  },

  deactivate: async (id: string): Promise<{ status: string; agent_id: string }> => {
    const response = await api.post<{ status: string; agent_id: string }>(`/agents/${id}/deactivate`)
    return response.data
  },
}

export const TokenService = {
  request: async (data: TokenRequest): Promise<TokenResponse> => {
    const response = await axios.post<TokenResponse>('/oauth/token', data, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })
    return response.data
  },

  refresh: async (refreshToken: string): Promise<TokenResponse> => {
    const response = await axios.post<TokenResponse>('/oauth/refresh', 
      new URLSearchParams({ refresh_token: refreshToken }),
      {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      }
    )
    return response.data
  },

  introspect: async (token: string): Promise<{ active: boolean }> => {
    const response = await axios.post<{ active: boolean }>('/oauth/introspect',
      new URLSearchParams({ token }),
      { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
    )
    return response.data
  },

  revoke: async (token: string): Promise<void> => {
    await axios.post('/oauth/revoke',
      new URLSearchParams({ token }),
      { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
    )
  },

  getJWKS: async (): Promise<unknown> => {
    const response = await axios.get('/.well-known/jwks.json')
    return response.data
  },
}

export const MetricsService = {
  get: async (): Promise<Metrics> => {
    const response = await axios.get<Metrics>('/metrics')
    return response.data
  },
}

export const HealthService = {
  check: async (): Promise<HealthCheck> => {
    const response = await axios.get<HealthCheck>('/health')
    return response.data
  },

  ready: async (): Promise<HealthCheck> => {
    const response = await axios.get<HealthCheck>('/health/ready')
    return response.data
  },
}

export default api
