import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { WebhookService } from '@/services/webhook'
import type { WebhookConfig } from '@/types/webhook'
import './webhook-list.css'

export function WebhookList() {
  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadWebhooks()
  }, [])

  const loadWebhooks = async () => {
    try {
      setLoading(true)
      const data = await WebhookService.list()
      setWebhooks(data)
      setError(null)
    } catch (err) {
      setError('Failed to load webhooks')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this webhook?')) return

    try {
      await WebhookService.delete(id)
      setWebhooks(webhooks.filter((w) => w.id !== id))
    } catch (err) {
      alert('Failed to delete webhook')
      console.error(err)
    }
  }

  const handleToggleActive = async (webhook: WebhookConfig) => {
    try {
      const updated = await WebhookService.update(webhook.id, {
        is_active: !webhook.is_active,
      })
      setWebhooks(webhooks.map((w) => (w.id === webhook.id ? updated : w)))
    } catch (err) {
      alert('Failed to update webhook')
      console.error(err)
    }
  }

  if (loading) {
    return <div className="loading">Loading webhooks...</div>
  }

  if (error) {
    return <div className="error">{error}</div>
  }

  return (
    <div className="webhook-list">
      <div className="page-header">
        <h2>Webhooks</h2>
        <Link to="/webhooks/new" className="btn btn-primary">
          Create Webhook
        </Link>
      </div>

      {webhooks.length === 0 ? (
        <div className="empty-state">
          <p>No webhooks configured. Create your first webhook to get started.</p>
          <Link to="/webhooks/new" className="btn btn-primary">
            Create Webhook
          </Link>
        </div>
      ) : (
        <table className="webhooks-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>URL</th>
              <th>Events</th>
              <th>Status</th>
              <th>Failures</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {webhooks.map((webhook) => (
              <tr key={webhook.id}>
                <td>
                  <Link to={`/webhooks/${webhook.id}`} className="webhook-name">
                    {webhook.name}
                  </Link>
                </td>
                <td>
                  <code className="webhook-url">{webhook.url}</code>
                </td>
                <td>
                  <div className="events">
                    {webhook.events.slice(0, 3).map((event) => (
                      <span key={event} className="event-badge">
                        {event}
                      </span>
                    ))}
                    {webhook.events.length > 3 && (
                      <span className="event-badge event-more">
                        +{webhook.events.length - 3} more
                      </span>
                    )}
                  </div>
                </td>
                <td>
                  <span
                    className={`status ${webhook.is_active ? 'active' : 'inactive'}`}
                  >
                    {webhook.is_active ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td>
                  <span
                    className={`fail-count ${webhook.consecutive_fails > 0 ? 'has-fails' : ''}`}
                  >
                    {webhook.consecutive_fails}
                  </span>
                </td>
                <td>{new Date(webhook.created_at).toLocaleDateString()}</td>
                <td>
                  <div className="actions">
                    <Link to={`/webhooks/${webhook.id}`} className="btn btn-sm">
                      View
                    </Link>
                    <button
                      onClick={() => handleToggleActive(webhook)}
                      className={`btn btn-sm ${webhook.is_active ? 'btn-warning' : 'btn-success'}`}
                    >
                      {webhook.is_active ? 'Disable' : 'Enable'}
                    </button>
                    <button
                      onClick={() => handleDelete(webhook.id)}
                      className="btn btn-sm btn-danger"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
