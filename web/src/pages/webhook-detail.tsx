import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { WebhookService } from '@/services/webhook'
import type { WebhookConfig, WebhookDelivery, TestWebhookResponse } from '@/types/webhook'
import { DELIVERY_STATUSES } from '@/types/webhook'
import type { DeliveryStatus } from '@/types/webhook'
import './webhook-detail.css'

type TabType = 'details' | 'deliveries' | 'test'

export function WebhookDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [webhook, setWebhook] = useState<WebhookConfig | null>(null)
  const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('details')

  // Test state
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<TestWebhookResponse | null>(null)
  const [testEvent, setTestEvent] = useState('webhook.test')
  const [testPayload, setTestPayload] = useState('')

  useEffect(() => {
    if (id) {
      loadWebhook()
    }
  }, [id])

  useEffect(() => {
    if (id && activeTab === 'deliveries') {
      loadDeliveries()
    }
  }, [id, activeTab])

  const loadWebhook = async () => {
    try {
      setLoading(true)
      const data = await WebhookService.get(id!)
      setWebhook(data)
      setError(null)
    } catch (err) {
      setError('Failed to load webhook')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const loadDeliveries = async () => {
    try {
      const data = await WebhookService.getDeliveries(id!)
      setDeliveries(data)
    } catch (err) {
      console.error('Failed to load deliveries', err)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this webhook? This cannot be undone.')) {
      return
    }
    try {
      await WebhookService.delete(id!)
      navigate('/webhooks')
    } catch (err) {
      alert('Failed to delete webhook')
      console.error(err)
    }
  }

  const handleToggleActive = async () => {
    if (!webhook) return
    try {
      const updated = await WebhookService.update(id!, {
        is_active: !webhook.is_active,
      })
      setWebhook(updated)
    } catch (err) {
      alert('Failed to update webhook')
      console.error(err)
    }
  }

  const handleTest = async () => {
    try {
      setTesting(true)
      setTestResult(null)
      const result = await WebhookService.test(id!, {
        event: testEvent,
        payload: testPayload || undefined,
      })
      setTestResult(result)
      // Refresh webhook to update last_tested_at
      loadWebhook()
    } catch (err) {
      setTestResult({
        success: false,
        status_code: 0,
        error: 'Failed to send test request',
      })
      console.error(err)
    } finally {
      setTesting(false)
    }
  }

  if (loading) {
    return <div className="loading">Loading webhook...</div>
  }

  if (error || !webhook) {
    return <div className="error">{error || 'Webhook not found'}</div>
  }

  return (
    <div className="webhook-detail">
      <div className="page-header">
        <button onClick={() => navigate('/webhooks')} className="back-link">
          ← Back to Webhooks
        </button>
      </div>

      <div className="detail-card">
        <div className="detail-header">
          <h2>{webhook.name}</h2>
          <div className="header-actions">
            <span className={`status ${webhook.is_active ? 'active' : 'inactive'}`}>
              {webhook.is_active ? 'Active' : 'Inactive'}
            </span>
          </div>
        </div>

        <div className="tabs">
          <button
            className={`tab ${activeTab === 'details' ? 'active' : ''}`}
            onClick={() => setActiveTab('details')}
          >
            Details
          </button>
          <button
            className={`tab ${activeTab === 'deliveries' ? 'active' : ''}`}
            onClick={() => setActiveTab('deliveries')}
          >
            Deliveries
          </button>
          <button
            className={`tab ${activeTab === 'test' ? 'active' : ''}`}
            onClick={() => setActiveTab('test')}
          >
            Test
          </button>
        </div>

        {activeTab === 'details' && (
          <div className="tab-content">
            <div className="detail-grid">
              <div className="detail-item">
                <label>Webhook ID</label>
                <code>{webhook.id}</code>
              </div>

              <div className="detail-item">
                <label>Endpoint URL</label>
                <code>{webhook.url}</code>
              </div>

              <div className="detail-item">
                <label>Created At</label>
                <span>{new Date(webhook.created_at).toLocaleString()}</span>
              </div>

              <div className="detail-item">
                <label>Updated At</label>
                <span>{new Date(webhook.updated_at).toLocaleString()}</span>
              </div>

              {webhook.last_tested_at && (
                <div className="detail-item">
                  <label>Last Tested</label>
                  <span>{new Date(webhook.last_tested_at).toLocaleString()}</span>
                </div>
              )}

              <div className="detail-item">
                <label>Max Retries</label>
                <span>{webhook.max_retries}</span>
              </div>

              <div className="detail-item">
                <label>Retry Backoff Base</label>
                <span>{webhook.retry_backoff_base}s</span>
              </div>

              <div className="detail-item">
                <label>Consecutive Failures</label>
                <span className={webhook.consecutive_fails > 0 ? 'fail-text' : ''}>
                  {webhook.consecutive_fails}
                </span>
              </div>

              <div className="detail-item full-width">
                <label>Events</label>
                <div className="events">
                  {webhook.events.map((event) => (
                    <span key={event} className="event-badge">
                      {event}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            <div className="detail-actions">
              <Link to={`/webhooks/${webhook.id}/edit`} className="btn btn-secondary">
                Edit
              </Link>
              <button
                onClick={handleToggleActive}
                className={`btn ${webhook.is_active ? 'btn-warning' : 'btn-success'}`}
              >
                {webhook.is_active ? 'Disable' : 'Enable'}
              </button>
              <button onClick={handleDelete} className="btn btn-danger">
                Delete
              </button>
            </div>
          </div>
        )}

        {activeTab === 'deliveries' && (
          <div className="tab-content">
            <div className="deliveries-header">
              <h3>Delivery History</h3>
              <button onClick={loadDeliveries} className="btn btn-sm btn-secondary">
                Refresh
              </button>
            </div>

            {deliveries.length === 0 ? (
              <div className="empty-state">
                <p>No deliveries yet. Trigger an event or send a test to see deliveries.</p>
              </div>
            ) : (
              <table className="deliveries-table">
                <thead>
                  <tr>
                    <th>Event</th>
                    <th>Status</th>
                    <th>Attempts</th>
                    <th>Last Attempt</th>
                    <th>Error</th>
                    <th>Created</th>
                  </tr>
                </thead>
                <tbody>
                  {deliveries.map((delivery) => (
                    <tr key={delivery.id}>
                      <td>
                        <span className="event-badge">{delivery.event}</span>
                      </td>
                      <td>
                        <span className={`delivery-status status-${delivery.status}`}>
                          {DELIVERY_STATUSES[delivery.status as DeliveryStatus] ||
                            delivery.status}
                        </span>
                      </td>
                      <td>{delivery.attempts}</td>
                      <td>
                        {delivery.last_attempt_at
                          ? new Date(delivery.last_attempt_at).toLocaleString()
                          : '-'}
                      </td>
                      <td>
                        {delivery.last_error ? (
                          <span className="error-text" title={delivery.last_error}>
                            {delivery.last_error.substring(0, 50)}
                            {delivery.last_error.length > 50 ? '...' : ''}
                          </span>
                        ) : (
                          '-'
                        )}
                      </td>
                      <td>{new Date(delivery.created_at).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {activeTab === 'test' && (
          <div className="tab-content">
            <h3>Test Webhook</h3>
            <p className="test-description">
              Send a test delivery to verify your webhook endpoint is configured correctly.
            </p>

            <div className="form-field">
              <label htmlFor="testEvent">Event Type</label>
              <select
                id="testEvent"
                value={testEvent}
                onChange={(e) => setTestEvent(e.target.value)}
                disabled={testing}
              >
                <option value="webhook.test">webhook.test</option>
                <option value="agent.created">agent.created</option>
                <option value="agent.deleted">agent.deleted</option>
                <option value="token.issued">token.issued</option>
              </select>
            </div>

            <div className="form-field">
              <label htmlFor="testPayload">Custom Payload (JSON, optional)</label>
              <textarea
                id="testPayload"
                value={testPayload}
                onChange={(e) => setTestPayload(e.target.value)}
                placeholder='{"key": "value"}'
                rows={4}
                disabled={testing}
              />
              <span className="hint">
                Leave empty to use a default test payload.
              </span>
            </div>

            <button
              onClick={handleTest}
              className="btn btn-primary"
              disabled={testing}
            >
              {testing ? 'Sending...' : 'Send Test'}
            </button>

            {testResult && (
              <div
                className={`test-result ${testResult.success ? 'test-success' : 'test-failure'}`}
              >
                <h4>{testResult.success ? 'Success' : 'Failed'}</h4>
                {testResult.status_code > 0 && (
                  <p>HTTP Status: {testResult.status_code}</p>
                )}
                {testResult.error && <p className="error-text">{testResult.error}</p>}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
