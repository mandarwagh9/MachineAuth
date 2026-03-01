export interface WebhookConfig {
  id: string
  organization_id?: string
  team_id?: string
  name: string
  url: string
  events: string[]
  is_active: boolean
  max_retries: number
  retry_backoff_base: number
  created_at: string
  updated_at: string
  last_tested_at?: string
  consecutive_fails: number
}

export interface WebhookDelivery {
  id: string
  webhook_config_id: string
  event: string
  payload: string
  headers?: string
  status: string
  attempts: number
  last_attempt_at?: string
  last_error?: string
  next_retry_at?: string
  created_at: string
}

export interface CreateWebhookRequest {
  name: string
  url: string
  events: string[]
  max_retries?: number
  retry_backoff_base?: number
}

export interface UpdateWebhookRequest {
  name?: string
  url?: string
  events?: string[]
  is_active?: boolean
  max_retries?: number
  retry_backoff_base?: number
}

export interface CreateWebhookResponse {
  webhook: WebhookConfig
  secret: string
}

export interface WebhookResponse {
  webhook: WebhookConfig
}

export interface WebhooksListResponse {
  webhooks: WebhookConfig[]
}

export interface WebhookDeliveryResponse {
  delivery: WebhookDelivery
}

export interface WebhookDeliveriesListResponse {
  deliveries: WebhookDelivery[]
}

export interface TestWebhookRequest {
  event?: string
  payload?: string
}

export interface TestWebhookResponse {
  success: boolean
  status_code: number
  error?: string
}

// Webhook event type constants
export const WEBHOOK_EVENTS = [
  'agent.created',
  'agent.deleted',
  'agent.updated',
  'agent.credentials_rotated',
  'token.issued',
  'token.validation_success',
  'token.validation_failed',
  'webhook.created',
  'webhook.updated',
  'webhook.deleted',
  'webhook.test',
] as const

export type WebhookEventType = (typeof WEBHOOK_EVENTS)[number]

export const DELIVERY_STATUSES = {
  pending: 'Pending',
  delivered: 'Delivered',
  failed: 'Failed',
  retrying: 'Retrying',
  dead: 'Dead Letter',
} as const

export type DeliveryStatus = keyof typeof DELIVERY_STATUSES
