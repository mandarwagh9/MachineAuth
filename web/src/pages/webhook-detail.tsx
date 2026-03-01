import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ArrowLeft, Webhook, CheckCircle, XCircle, AlertTriangle,
  Edit, Trash2, RefreshCw, Play, ToggleLeft, ToggleRight, Clock
} from 'lucide-react'
import { WebhookService } from '@/services/webhook'
import type { WebhookConfig, WebhookDelivery, TestWebhookResponse } from '@/types/webhook'
import { toast } from 'sonner'

type TabType = 'details' | 'deliveries' | 'test'

const STATUS_STYLES: Record<string, string> = {
  delivered: 'bg-green-100 text-green-700',
  pending: 'bg-blue-100 text-blue-700',
  retrying: 'bg-yellow-100 text-yellow-700',
  failed: 'bg-red-100 text-red-700',
  dead: 'bg-slate-100 text-slate-600',
}

function InfoRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start py-3 border-b border-slate-100 last:border-0">
      <span className="w-44 text-sm text-slate-500 flex-shrink-0">{label}</span>
      <div className="flex-1 text-sm text-slate-900">{children}</div>
    </div>
  )
}

export function WebhookDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [webhook, setWebhook] = useState<WebhookConfig | null>(null)
  const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<TabType>('details')
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<TestWebhookResponse | null>(null)
  const [testEvent, setTestEvent] = useState('webhook.test')
  const [testPayload, setTestPayload] = useState('')

  useEffect(() => {
    if (id) loadWebhook()
  }, [id])

  useEffect(() => {
    if (id && activeTab === 'deliveries') loadDeliveries()
  }, [id, activeTab])

  const loadWebhook = async () => {
    try {
      const data = await WebhookService.get(id!)
      setWebhook(data)
    } catch {
      toast.error('Failed to load webhook')
      navigate('/webhooks')
    } finally {
      setLoading(false)
    }
  }

  const loadDeliveries = async () => {
    try {
      const data = await WebhookService.getDeliveries(id!)
      setDeliveries(data || [])
    } catch {
      toast.error('Failed to load deliveries')
    }
  }

  const handleDelete = async () => {
    if (!confirm('Delete this webhook? This cannot be undone.')) return
    try {
      await WebhookService.delete(id!)
      toast.success('Webhook deleted')
      navigate('/webhooks')
    } catch {
      toast.error('Failed to delete webhook')
    }
  }

  const handleToggle = async () => {
    if (!webhook) return
    try {
      const updated = await WebhookService.update(id!, { is_active: !webhook.is_active })
      setWebhook(updated)
      toast.success(`Webhook ${updated.is_active ? 'enabled' : 'disabled'}`)
    } catch {
      toast.error('Failed to update webhook')
    }
  }

  const handleTest = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      const result = await WebhookService.test(id!, {
        event: testEvent,
        payload: testPayload || undefined,
      })
      setTestResult(result)
      loadWebhook()
    } catch {
      setTestResult({ success: false, status_code: 0, error: 'Failed to reach endpoint' })
    } finally {
      setTesting(false)
    }
  }

  const fmt = (d?: string) => d ? new Date(d).toLocaleString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
  }) : '—'

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="h-64 bg-slate-200 rounded-xl" />
        </div>
      </div>
    )
  }

  if (!webhook) return null

  return (
    <div className="p-8 max-w-5xl">
      {/* Header */}
      <div className="mb-8">
        <Link to="/webhooks" className="inline-flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-700 mb-4">
          <ArrowLeft className="w-4 h-4" />
          Back to Webhooks
        </Link>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-indigo-100 rounded-xl flex items-center justify-center">
              <Webhook className="w-6 h-6 text-indigo-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-slate-900">{webhook.name}</h1>
              <p className="text-slate-500 text-sm font-mono mt-0.5">{webhook.url}</p>
              <div className="flex items-center gap-2 mt-2">
                {webhook.is_active ? (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-green-100 text-green-700 rounded-full text-xs font-medium">
                    <CheckCircle className="w-3 h-3" /> Active
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 text-slate-600 rounded-full text-xs font-medium">
                    <XCircle className="w-3 h-3" /> Disabled
                  </span>
                )}
                {webhook.consecutive_fails > 0 && (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-red-100 text-red-700 rounded-full text-xs font-medium">
                    <AlertTriangle className="w-3 h-3" /> {webhook.consecutive_fails} failures
                  </span>
                )}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Link
              to={`/webhooks/${webhook.id}/edit`}
              className="flex items-center gap-2 px-3 py-2 border border-slate-300 text-slate-700 rounded-lg text-sm font-medium hover:bg-slate-50 transition-colors"
            >
              <Edit className="w-4 h-4" /> Edit
            </Link>
            <button
              onClick={handleToggle}
              className="flex items-center gap-2 px-3 py-2 border border-slate-300 text-slate-700 rounded-lg text-sm font-medium hover:bg-slate-50 transition-colors"
            >
              {webhook.is_active
                ? <><ToggleRight className="w-4 h-4 text-green-500" /> Disable</>
                : <><ToggleLeft className="w-4 h-4" /> Enable</>}
            </button>
            <button
              onClick={handleDelete}
              className="flex items-center gap-2 px-3 py-2 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm font-medium hover:bg-red-100 transition-colors"
            >
              <Trash2 className="w-4 h-4" /> Delete
            </button>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-slate-100 rounded-lg p-1 w-fit">
        {(['details', 'deliveries', 'test'] as TabType[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors capitalize ${
              activeTab === tab ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab: Details */}
      {activeTab === 'details' && (
        <div className="bg-white border border-slate-200 rounded-xl p-6">
          <InfoRow label="Webhook ID"><code className="text-xs font-mono text-slate-600">{webhook.id}</code></InfoRow>
          <InfoRow label="Endpoint URL"><a href={webhook.url} target="_blank" rel="noreferrer" className="text-blue-600 hover:underline font-mono text-xs">{webhook.url}</a></InfoRow>
          <InfoRow label="Status">
            {webhook.is_active
              ? <span className="text-green-600 flex items-center gap-1"><CheckCircle className="w-4 h-4" /> Active</span>
              : <span className="text-slate-400 flex items-center gap-1"><XCircle className="w-4 h-4" /> Disabled</span>}
          </InfoRow>
          <InfoRow label="Events">
            <div className="flex flex-wrap gap-1">
              {webhook.events.map((e) => (
                <span key={e} className="px-2 py-0.5 bg-blue-50 text-blue-700 rounded text-xs font-medium">{e}</span>
              ))}
            </div>
          </InfoRow>
          <InfoRow label="Max Retries">{webhook.max_retries}</InfoRow>
          <InfoRow label="Backoff Base">{webhook.retry_backoff_base}s</InfoRow>
          <InfoRow label="Consecutive Fails">
            <span className={webhook.consecutive_fails > 0 ? 'text-red-600 font-semibold' : ''}>
              {webhook.consecutive_fails}
            </span>
          </InfoRow>
          <InfoRow label="Last Tested"><span className="flex items-center gap-1"><Clock className="w-4 h-4 text-slate-400" />{fmt(webhook.last_tested_at)}</span></InfoRow>
          <InfoRow label="Created"><span className="flex items-center gap-1"><Clock className="w-4 h-4 text-slate-400" />{fmt(webhook.created_at)}</span></InfoRow>
          <InfoRow label="Updated"><span className="flex items-center gap-1"><Clock className="w-4 h-4 text-slate-400" />{fmt(webhook.updated_at)}</span></InfoRow>
        </div>
      )}

      {/* Tab: Deliveries */}
      {activeTab === 'deliveries' && (
        <div className="bg-white border border-slate-200 rounded-xl overflow-hidden">
          <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200">
            <h2 className="font-semibold text-slate-900">Delivery History</h2>
            <button
              onClick={loadDeliveries}
              className="flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-700"
            >
              <RefreshCw className="w-4 h-4" /> Refresh
            </button>
          </div>
          <table className="w-full">
            <thead className="bg-slate-50 border-b border-slate-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Event</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Attempts</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Last Attempt</th>
                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Error</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {deliveries.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-slate-500">
                    No deliveries yet. Trigger an event or send a test to see results.
                  </td>
                </tr>
              ) : (
                deliveries.map((d) => (
                  <tr key={d.id} className="hover:bg-slate-50">
                    <td className="px-6 py-3">
                      <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-xs font-medium">{d.event}</span>
                    </td>
                    <td className="px-6 py-3">
                      <span className={`px-2 py-0.5 rounded text-xs font-medium capitalize ${STATUS_STYLES[d.status] || 'bg-slate-100 text-slate-600'}`}>
                        {d.status}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-sm text-slate-600">{d.attempts}</td>
                    <td className="px-6 py-3 text-sm text-slate-600">{d.last_attempt_at ? fmt(d.last_attempt_at) : '—'}</td>
                    <td className="px-6 py-3 text-xs text-red-600 max-w-xs truncate" title={d.last_error}>
                      {d.last_error || '—'}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Tab: Test */}
      {activeTab === 'test' && (
        <div className="bg-white border border-slate-200 rounded-xl p-6 max-w-xl space-y-5">
          <div>
            <h2 className="font-semibold text-slate-900">Test Webhook</h2>
            <p className="text-sm text-slate-500 mt-1">Send a test request to verify your endpoint is configured correctly.</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Event Type</label>
            <select
              value={testEvent}
              onChange={(e) => setTestEvent(e.target.value)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="webhook.test">webhook.test</option>
              <option value="agent.created">agent.created</option>
              <option value="agent.deleted">agent.deleted</option>
              <option value="token.issued">token.issued</option>
              <option value="agent.rotated">agent.rotated</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Custom Payload <span className="text-slate-400 font-normal">(optional JSON)</span>
            </label>
            <textarea
              value={testPayload}
              onChange={(e) => setTestPayload(e.target.value)}
              placeholder={'{"key": "value"}'}
              rows={3}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
            />
          </div>

          <button
            onClick={handleTest}
            disabled={testing}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {testing ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
            {testing ? 'Sending...' : 'Send Test'}
          </button>

          {testResult && (
            <div className={`p-4 rounded-xl border ${testResult.success ? 'bg-green-50 border-green-200' : 'bg-red-50 border-red-200'}`}>
              <div className="flex items-center gap-2 mb-1">
                {testResult.success
                  ? <CheckCircle className="w-5 h-5 text-green-500" />
                  : <XCircle className="w-5 h-5 text-red-500" />}
                <span className={`font-semibold text-sm ${testResult.success ? 'text-green-800' : 'text-red-800'}`}>
                  {testResult.success ? 'Delivery succeeded' : 'Delivery failed'}
                </span>
                {testResult.status_code > 0 && (
                  <span className="text-xs text-slate-500">· HTTP {testResult.status_code}</span>
                )}
              </div>
              {testResult.error && <p className="text-sm text-red-700 mt-1">{testResult.error}</p>}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

