import axios from 'axios'
import type {
  Agent,
  CreateAgentRequest,
  CreateAgentResponse,
  AgentsListResponse,
  TokenRequest,
  TokenResponse,
} from '@/types'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

export const AgentService = {
  list: async (): Promise<Agent[]> => {
    const response = await api.get<AgentsListResponse>('/agents')
    return response.data.agents
  },

  get: async (id: string): Promise<Agent> => {
    const response = await api.get<{ agent: Agent }>(`/agents/${id}`)
    return response.data.agent
  },

  create: async (data: CreateAgentRequest): Promise<CreateAgentResponse> => {
    const response = await api.post<CreateAgentResponse>('/agents', data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/agents/${id}`)
  },

  rotate: async (id: string): Promise<{ client_secret: string }> => {
    const response = await api.post<{ client_secret: string }>(`/agents/${id}`, {
      action: 'rotate',
    })
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

  getJWKS: async (): Promise<unknown> => {
    const response = await axios.get('/.well-known/jwks.json')
    return response.data
  },
}

export default api
