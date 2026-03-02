import axios from 'axios'
import type { TokenRequest, TokenResponse, Metrics, HealthCheck, Agent, AgentUsage, CreateOrganizationRequest, CreateTeamRequest, CreateAPIKeyRequest, CreateAPIKeyResponse } from '@/types'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'https://auth.writesomething.fun'

const api = axios.create({
  baseURL: API_BASE_URL + '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

export const AgentService = {
  list: async () => {
    const response = await api.get('/agents')
    return response.data
  },

  get: async (id: string) => {
    const response = await api.get(`/agents/${id}`)
    return response.data.agent
  },

  create: async (data: unknown) => {
    const response = await api.post('/agents', data)
    return response.data
  },

  rotate: async (id: string) => {
    const response = await api.post(`/agents/${id}/rotate`)
    return response.data
  },

  deactivate: async (id: string) => {
    const response = await api.post(`/agents/${id}/deactivate`)
    return response.data
  },
}

export const AgentSelfService = {
  getMe: async (token: string): Promise<{ agent: Agent }> => {
    const response = await api.get('/agents/me', {
      headers: { Authorization: `Bearer ${token}` },
    })
    return response.data
  },

  getUsage: async (token: string): Promise<AgentUsage> => {
    const response = await api.get('/agents/me/usage', {
      headers: { Authorization: `Bearer ${token}` },
    })
    return response.data
  },

  rotateCredentials: async (token: string): Promise<{ client_secret: string }> => {
    const response = await api.post('/agents/me/rotate', {}, {
      headers: { Authorization: `Bearer ${token}` },
    })
    return response.data
  },

  deactivate: async (token: string): Promise<{ message: string }> => {
    const response = await api.post('/agents/me/deactivate', {}, {
      headers: { Authorization: `Bearer ${token}` },
    })
    return response.data
  },

  reactivate: async (token: string): Promise<{ message: string }> => {
    const response = await api.post('/agents/me/reactivate', {}, {
      headers: { Authorization: `Bearer ${token}` },
    })
    return response.data
  },

  delete: async (token: string): Promise<void> => {
    await api.delete('/agents/me/delete', {
      headers: { Authorization: `Bearer ${token}` },
    })
  },
}

export const TokenService = {
  request: async (data: TokenRequest): Promise<TokenResponse> => {
    const response = await axios.post<TokenResponse>(API_BASE_URL + '/oauth/token', data, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })
    return response.data
  },

  refresh: async (refreshToken: string): Promise<TokenResponse> => {
    const response = await axios.post<TokenResponse>(API_BASE_URL + '/oauth/refresh', 
      new URLSearchParams({ refresh_token: refreshToken }),
      {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      }
    )
    return response.data
  },

  introspect: async (token: string): Promise<{ active: boolean }> => {
    const response = await axios.post<{ active: boolean }>(API_BASE_URL + '/oauth/introspect',
      new URLSearchParams({ token }),
      { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
    )
    return response.data
  },

  revoke: async (token: string): Promise<void> => {
    await axios.post(API_BASE_URL + '/oauth/revoke',
      new URLSearchParams({ token }),
      { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
    )
  },

  getJWKS: async (): Promise<unknown> => {
    const response = await axios.get(API_BASE_URL + '/.well-known/jwks.json')
    return response.data
  },
}

export const MetricsService = {
  get: async (): Promise<Metrics> => {
    const response = await api.get<Metrics>('/stats')
    return response.data
  },
}

export const HealthService = {
  check: async (): Promise<HealthCheck> => {
    const response = await axios.get<HealthCheck>(API_BASE_URL + '/health')
    return response.data
  },

  ready: async (): Promise<HealthCheck> => {
    const response = await axios.get<HealthCheck>(API_BASE_URL + '/health/ready')
    return response.data
  },
}

export const OrganizationService = {
  list: async () => {
    const response = await api.get('/organizations')
    return response.data
  },

  create: async (data: CreateOrganizationRequest) => {
    const response = await api.post('/organizations', data)
    return response.data
  },

  get: async (id: string) => {
    const response = await api.get(`/organizations/${id}`)
    return response.data
  },

  update: async (id: string, data: Partial<CreateOrganizationRequest>) => {
    const response = await api.patch(`/organizations/${id}`, data)
    return response.data
  },

  delete: async (id: string) => {
    await api.delete(`/organizations/${id}`)
  },

  listTeams: async (orgId: string) => {
    const response = await api.get(`/organizations/${orgId}/teams`)
    return response.data
  },

  createTeam: async (orgId: string, data: CreateTeamRequest) => {
    const response = await api.post(`/organizations/${orgId}/teams`, data)
    return response.data
  },

  listAgents: async (orgId: string) => {
    const response = await api.get(`/organizations/${orgId}/agents`)
    return response.data
  },

  createAgent: async (orgId: string, data: unknown) => {
    const response = await api.post(`/organizations/${orgId}/agents`, data)
    return response.data
  },

  listAPIKeys: async (orgId: string) => {
    const response = await api.get(`/organizations/${orgId}/api-keys`)
    return response.data
  },

  createAPIKey: async (orgId: string, data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> => {
    const response = await api.post(`/organizations/${orgId}/api-keys`, data)
    return response.data
  },

  deleteAPIKey: async (orgId: string, keyId: string) => {
    await api.delete(`/organizations/${orgId}/api-keys/${keyId}`)
  },
}

export default api
