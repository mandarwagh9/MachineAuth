import axios from 'axios'
import type { TokenRequest, TokenResponse, Metrics, HealthCheck, Agent, AgentUsage } from '@/types'

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
    const response = await axios.get<Metrics>(API_BASE_URL + '/metrics')
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

export default api
