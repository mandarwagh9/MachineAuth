import { useState, useEffect } from 'react'
import { useNavigate, useParams, Link } from 'react-router-dom'
import { ArrowLeft, Webhook, AlertTriangle, CheckCircle, Copy, RefreshCw } from 'lucide-react'
import { WebhookService } from '@/services/webhook'
import { WEBHOOK_EVENTS } from '@/types/webhook'
import type { WebhookConfig } from '@/types/webhook'
import { toast } from 'sonner'

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
  const [createdWebhook, setCreatedWebhook] = useState<{ id: string; secret: string } | null>(null)
  const [copiedSecret, setCopiedSecret] = useState(false)

  useEffect(() => {
    if (isEdit && id) loadWebhook(id)
  }, [id])

  const loadWebhook = async (webhookId: string) => {
    try {
      const webhook: WebhookConfig = await WebhookService.get(webhookId)
      setName(webhook.name)
      setUrl(webhook.url)
      setSelectedEvents(webhook.events)
      setMaxRetries(String(webhook.max_retries))
      setRetryBackoffBase(String(webhook.retry_backoff_base))
    } catch {
      toast.error('Failed to load webhook')
      navigate('/webhooks')
    } finally {
      setLoadingWebhook(false)
    }
  }

  const toggleEvent = (event: string) => {
    setSelectedEvents((prev) =>
      prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event]
    )
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) { setError('Name is required'); return }
    if (!url.trim()) { setError('URL is required'); return }
    if (selectedEvents.length === 0) { setError('At least one event must be selected'); return }

    setLoading(true)
    setError(null)
    try {
      if (isEdit && id) {
        await WebhookService.update(id, {
          name: name.trim(),
          url: url.trim(),
          events: selectedEvents,
          max_retries: parseInt(maxRetries, 10),
          retry_backoff_base: parseInt(retryBackoffBase, 10),
        })
        toast.success('Webhook updated')
        navigate(`/webhooks/${id}`)
      } else {
        const response = await WebhookService.create({
          name: name.trim(),
          url: url.trim(),
          events: selectedEvents,
          max_retries: parseInt(maxRetries, 10) || undefined,
          retry_backoff_base: parseInt(retryBackoffBase, 10) || undefined,
        })
        setCreatedWebhook({ id: response.webhook.id, secret: response.secret })
      }
    } catch {
      setError(isEdit ? 'Failed to update webhook' : 'Failed to create webhook')
    } finally {
      setLoading(false)
    }
  }

  const copySecret = () => {
    if (!createdWebhook) return
    navigator.clipboard.writeText(createdWebhook.secret)
    setCopiedSecret(true)
    toast.success('Secret copied')
    setTimeout(() => setCopiedSecret(false), 2000)
  }

  if (loadingWebhook) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="h-64 bg-slate-200 rounded-xl" />
        </div>
      </div>
    )
  }

  if (createdWebhook) {
    return (
      <div className="p-8 max-w-2xl">
        <div className="bg-white border border-slate-200 rounded-xl p-8 text-center">
          <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <CheckCircle className="w-8 h-8 text-green-600" />
          </div>
          <h2 className="text-xl font-bold text-slate-900">Webhook Created!</h2>
          <p className="text-slate-500 mt-2 mb-6">
            Save your webhook secret now — it won't be shown again.
          </p>

          <div className="bg-amber-50 border border-amber-200 rounded-xl p-4 text-left mb-6">
            <div className="flex items-start gap-3 mb-3">
              <AlertTriangle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm font-semibold text-amber-800">One-time secret</p>
                <p className="text-sm text-amber-700">
                  Use this to verify the <code className="bg-amber-100 px-1 rounded text-xs">X-Webhook-Signature-256</code> header on incoming requests.
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-sm font-mono bg-white border border-amber-200 rounded-lg px-3 py-2 text-slate-800 break-all">
                {createdWebhook.secret}
              </code>
              <button
                onClick={copySecret}
                className="p-2 text-amber-600 hover:text-amber-800 flex-shrink-0"
              >
                {copiedSecret ? <CheckCircle className="w-5 h-5 text-green-500" /> : <Copy className="w-5 h-5" />}
              </button>
            </div>
          </div>

          <div className="flex gap-3 justify-center">
            <button
              onClick={() => navigate(`/webhooks/${createdWebhook.id}`)}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
            >
              View Webhook
            </button>
            <button
              onClick={() => navigate('/webhooks')}
              className="px-4 py-2 border border-slate-300 text-slate-700 rounded-lg text-sm font-medium hover:bg-slate-50 transition-colors"
            >
              Back to Webhooks
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-2xl">
      <div className="mb-8">
        <Link to="/webhooks" className="inline-flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-700 mb-4">
          <ArrowLeft className="w-4 h-4" />
          Back to Webhooks
        </Link>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center">
            <Webhook className="w-5 h-5 text-indigo-600" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-slate-900">{isEdit ? 'Edit Webhook' : 'Create Webhook'}</h1>
            <p className="text-slate-500 text-sm">{isEdit ? 'Update webhook configuration' : 'Configure a new event endpoint'}</p>
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg">
            <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        <div className="bg-white border border-slate-200 rounded-xl p-6 space-y-4">
          <h2 className="font-semibold text-slate-900">Basic Information</h2>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Webhook Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Webhook"
              disabled={loading}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Endpoint URL <span className="text-red-500">*</span>
            </label>
            <input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/webhooks"
              disabled={loading}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
            />
            <p className="text-xs text-slate-500 mt-1">The URL that will receive POST requests for events.</p>
          </div>
        </div>

        <div className="bg-white border border-slate-200 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-900">Events <span className="text-red-500">*</span></h2>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setSelectedEvents([...WEBHOOK_EVENTS])}
                className="text-xs text-blue-600 hover:underline"
              >
                Select all
              </button>
              <span className="text-slate-300">|</span>
              <button
                type="button"
                onClick={() => setSelectedEvents([])}
                className="text-xs text-slate-500 hover:underline"
              >
                Clear
              </button>
              <span className="text-xs text-slate-500 ml-2">{selectedEvents.length} selected</span>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            {WEBHOOK_EVENTS.map((event) => (
              <label key={event} className="flex items-center gap-2 p-2.5 rounded-lg hover:bg-slate-50 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedEvents.includes(event)}
                  onChange={() => toggleEvent(event)}
                  disabled={loading}
                  className="w-4 h-4 text-blue-600 border-slate-300 rounded"
                />
                <span className="text-sm text-slate-700 font-medium">{event}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="bg-white border border-slate-200 rounded-xl p-6">
          <h2 className="font-semibold text-slate-900 mb-4">Retry Settings</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Max Retries</label>
              <input
                type="number"
                value={maxRetries}
                onChange={(e) => setMaxRetries(e.target.value)}
                min="0" max="20"
                disabled={loading}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              />
              <p className="text-xs text-slate-500 mt-1">0–20 attempts</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Backoff Base (s)</label>
              <input
                type="number"
                value={retryBackoffBase}
                onChange={(e) => setRetryBackoffBase(e.target.value)}
                min="1" max="60"
                disabled={loading}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              />
              <p className="text-xs text-slate-500 mt-1">Exponential backoff base</p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {loading && <RefreshCw className="w-4 h-4 animate-spin" />}
            {loading ? (isEdit ? 'Saving...' : 'Creating...') : (isEdit ? 'Save Changes' : 'Create Webhook')}
          </button>
          <button
            type="button"
            onClick={() => navigate('/webhooks')}
            disabled={loading}
            className="px-5 py-2.5 border border-slate-300 text-slate-700 rounded-lg text-sm font-medium hover:bg-slate-50 disabled:opacity-50 transition-colors"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}

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
