import api from './api'
import type {
  WebhookConfig,
  CreateWebhookRequest,
  CreateWebhookResponse,
  WebhooksListResponse,
  WebhookResponse,
  WebhookDelivery,
  WebhookDeliveriesListResponse,
  TestWebhookRequest,
  TestWebhookResponse,
  UpdateWebhookRequest,
} from '@/types/webhook'

export const WebhookService = {
  list: async (): Promise<WebhookConfig[]> => {
    const response = await api.get<WebhooksListResponse>('/webhooks')
    return response.data.webhooks
  },

  get: async (id: string): Promise<WebhookConfig> => {
    const response = await api.get<WebhookResponse>(`/webhooks/${id}`)
    return response.data.webhook
  },

  create: async (data: CreateWebhookRequest): Promise<CreateWebhookResponse> => {
    const response = await api.post<CreateWebhookResponse>('/webhooks', data)
    return response.data
  },

  update: async (id: string, data: UpdateWebhookRequest): Promise<WebhookConfig> => {
    const response = await api.put<WebhookResponse>(`/webhooks/${id}`, data)
    return response.data.webhook
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/webhooks/${id}`)
  },

  test: async (id: string, data?: TestWebhookRequest): Promise<TestWebhookResponse> => {
    const response = await api.post<TestWebhookResponse>(`/webhooks/${id}/test`, data || {})
    return response.data
  },

  getDeliveries: async (webhookId: string): Promise<WebhookDelivery[]> => {
    const response = await api.get<WebhookDeliveriesListResponse>(
      `/webhooks/${webhookId}/deliveries`
    )
    return response.data.deliveries
  },

  getEvents: async (): Promise<string[]> => {
    const response = await api.get<{ events: string[] }>('/webhook-events')
    return response.data.events
  },
}
