import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { WebhookService } from '@/services/webhook'
import { WEBHOOK_EVENTS } from '@/types/webhook'
import type { WebhookConfig } from '@/types/webhook'
import './webhook-form.css'

export function WebhookForm() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const isEdit = Boolean(id)

  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [selectedEvents, setSelectedEvents] = useState<string[]>([])
  const [maxRetries, setMaxRetries] = useState('10')
  const [retryBackoffBase, setRetryBackoffBase] = useState('2')
  const [loading, setLoading] = useState(false)
  const [loadingWebhook, setLoadingWebhook] = useState(isEdit)
  const [error, setError] = useState<string | null>(null)
  const [createdWebhook, setCreatedWebhook] = useState<{
    id: string
    secret: string
  } | null>(null)

  useEffect(() => {
    if (isEdit && id) {
      loadWebhook(id)
    }
  }, [id])

  const loadWebhook = async (webhookId: string) => {
    try {
      setLoadingWebhook(true)
      const webhook: WebhookConfig = await WebhookService.get(webhookId)
      setName(webhook.name)
      setUrl(webhook.url)
      setSelectedEvents(webhook.events)
      setMaxRetries(String(webhook.max_retries))
      setRetryBackoffBase(String(webhook.retry_backoff_base))
    } catch (err) {
      setError('Failed to load webhook')
      console.error(err)
    } finally {
      setLoadingWebhook(false)
    }
  }

  const toggleEvent = (event: string) => {
    setSelectedEvents((prev) =>
      prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event]
    )
  }

  const selectAllEvents = () => {
    setSelectedEvents([...WEBHOOK_EVENTS])
  }

  const clearAllEvents = () => {
    setSelectedEvents([])
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      setError('Name is required')
      return
    }
    if (!url.trim()) {
      setError('URL is required')
      return
    }
    if (selectedEvents.length === 0) {
      setError('At least one event must be selected')
      return
    }

    try {
      setLoading(true)
      setError(null)

      if (isEdit && id) {
        await WebhookService.update(id, {
          name: name.trim(),
          url: url.trim(),
          events: selectedEvents,
          max_retries: parseInt(maxRetries, 10),
          retry_backoff_base: parseInt(retryBackoffBase, 10),
        })
        navigate(`/webhooks/${id}`)
      } else {
        const response = await WebhookService.create({
          name: name.trim(),
          url: url.trim(),
          events: selectedEvents,
          max_retries: parseInt(maxRetries, 10) || undefined,
          retry_backoff_base: parseInt(retryBackoffBase, 10) || undefined,
        })
        setCreatedWebhook({
          id: response.webhook.id,
          secret: response.secret,
        })
      }
    } catch (err) {
      setError(isEdit ? 'Failed to update webhook' : 'Failed to create webhook')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  if (loadingWebhook) {
    return <div className="loading">Loading webhook...</div>
  }

  if (createdWebhook) {
    return (
      <div className="webhook-form">
        <div className="success-card">
          <h2>Webhook Created Successfully</h2>
          <p className="warning">
            Copy your webhook secret now. You won't be able to see it again!
          </p>
          <p className="hint">
            Use this secret to verify the HMAC-SHA256 signature in the{' '}
            <code>X-Webhook-Signature-256</code> header.
          </p>

          <div className="credential-field">
            <label>Webhook Secret</label>
            <code className="secret">{createdWebhook.secret}</code>
          </div>

          <div className="actions">
            <button
              onClick={() => navigate(`/webhooks/${createdWebhook.id}`)}
              className="btn btn-primary"
            >
              View Webhook
            </button>
            <button
              onClick={() => navigate('/webhooks')}
              className="btn btn-secondary"
            >
              Back to Webhooks
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="webhook-form">
      <div className="form-card">
        <h2>{isEdit ? 'Edit Webhook' : 'Create New Webhook'}</h2>

        <form onSubmit={handleSubmit}>
          {error && <div className="error-message">{error}</div>}

          <div className="form-field">
            <label htmlFor="name">Webhook Name</label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Webhook"
              disabled={loading}
            />
          </div>

          <div className="form-field">
            <label htmlFor="url">Endpoint URL</label>
            <input
              id="url"
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/webhooks"
              disabled={loading}
            />
            <span className="hint">
              The URL that will receive webhook POST requests.
            </span>
          </div>

          <div className="form-field">
            <label>Events</label>
            <div className="events-selector">
              <div className="events-actions">
                <button
                  type="button"
                  onClick={selectAllEvents}
                  className="btn btn-sm btn-secondary"
                >
                  Select All
                </button>
                <button
                  type="button"
                  onClick={clearAllEvents}
                  className="btn btn-sm btn-secondary"
                >
                  Clear All
                </button>
                <span className="events-count">
                  {selectedEvents.length} selected
                </span>
              </div>
              <div className="events-grid">
                {WEBHOOK_EVENTS.map((event) => (
                  <label key={event} className="event-checkbox">
                    <input
                      type="checkbox"
                      checked={selectedEvents.includes(event)}
                      onChange={() => toggleEvent(event)}
                      disabled={loading}
                    />
                    <span className="event-label">{event}</span>
                  </label>
                ))}
              </div>
            </div>
          </div>

          <div className="form-row">
            <div className="form-field">
              <label htmlFor="maxRetries">Max Retries</label>
              <input
                id="maxRetries"
                type="number"
                value={maxRetries}
                onChange={(e) => setMaxRetries(e.target.value)}
                min="0"
                max="20"
                disabled={loading}
              />
              <span className="hint">
                Number of retry attempts after initial failure (0-20).
              </span>
            </div>

            <div className="form-field">
              <label htmlFor="retryBackoff">Retry Backoff Base (seconds)</label>
              <input
                id="retryBackoff"
                type="number"
                value={retryBackoffBase}
                onChange={(e) => setRetryBackoffBase(e.target.value)}
                min="1"
                max="60"
                disabled={loading}
              />
              <span className="hint">
                Base for exponential backoff between retries.
              </span>
            </div>
          </div>

          <div className="form-actions">
            <button
              type="button"
              onClick={() => navigate('/webhooks')}
              className="btn btn-secondary"
              disabled={loading}
            >
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading
                ? isEdit
                  ? 'Saving...'
                  : 'Creating...'
                : isEdit
                  ? 'Save Changes'
                  : 'Create Webhook'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
