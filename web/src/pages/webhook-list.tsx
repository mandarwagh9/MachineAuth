import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { 
  Webhook, 
  Plus, 
  Search,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Eye,
  Trash2,
  ToggleLeft,
  ToggleRight,
  Clock
} from 'lucide-react'
import { WebhookService } from '@/services/webhook'
import type { WebhookConfig } from '@/types/webhook'
import { toast } from 'sonner'

export function WebhookList() {
  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    loadWebhooks()
  }, [])

  const loadWebhooks = async () => {
    try {
      const data = await WebhookService.list()
      setWebhooks(data)
    } catch {
      toast.error('Failed to load webhooks')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete webhook "${name}"? This cannot be undone.`)) return
    try {
      await WebhookService.delete(id)
      setWebhooks(webhooks.filter((w) => w.id !== id))
      toast.success('Webhook deleted')
    } catch {
      toast.error('Failed to delete webhook')
    }
  }

  const handleToggle = async (webhook: WebhookConfig) => {
    try {
      const updated = await WebhookService.update(webhook.id, { is_active: !webhook.is_active })
      setWebhooks(webhooks.map((w) => (w.id === webhook.id ? updated : w)))
      toast.success(`Webhook ${updated.is_active ? 'enabled' : 'disabled'}`)
    } catch {
      toast.error('Failed to update webhook')
    }
  }

  const filteredWebhooks = webhooks.filter(w =>
    w.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    w.url.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const formatDate = (d: string) =>
    new Date(d).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="h-16 bg-slate-200 rounded-xl" />
          {[1, 2, 3].map(i => <div key={i} className="h-20 bg-slate-200 rounded-xl" />)}
        </div>
      </div>
    )
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Webhooks</h1>
          <p className="text-slate-500 mt-1">Receive real-time event notifications</p>
        </div>
        <Link
          to="/webhooks/new"
          className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
        >
          <Plus className="w-4 h-4" />
          Create Webhook
        </Link>
      </div>

      {/* Search */}
      <div className="flex items-center gap-3 mb-6">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
          <input
            type="text"
            placeholder="Search webhooks..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
        <span className="text-sm text-slate-500">{filteredWebhooks.length} webhook{filteredWebhooks.length !== 1 ? 's' : ''}</span>
      </div>

      {/* Table */}
      <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
        <table className="w-full">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Webhook</th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Events</th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Status</th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Failures</th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase">Created</th>
              <th className="px-6 py-3 text-right text-xs font-semibold text-slate-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {filteredWebhooks.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-16 text-center">
                  <Webhook className="w-12 h-12 mx-auto mb-3 text-slate-300" />
                  <p className="text-slate-500 font-medium">No webhooks configured</p>
                  <p className="text-sm text-slate-400 mt-1">
                    {searchQuery ? 'Try a different search.' : 'Create one to start receiving event notifications.'}
                  </p>
                  {!searchQuery && (
                    <Link to="/webhooks/new" className="mt-4 inline-block text-blue-600 hover:underline text-sm">
                      Create your first webhook →
                    </Link>
                  )}
                </td>
              </tr>
            ) : (
              filteredWebhooks.map((webhook) => (
                <tr key={webhook.id} className="hover:bg-slate-50">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="w-9 h-9 bg-indigo-100 rounded-lg flex items-center justify-center flex-shrink-0">
                        <Webhook className="w-4 h-4 text-indigo-600" />
                      </div>
                      <div>
                        <Link to={`/webhooks/${webhook.id}`} className="font-medium text-slate-900 hover:text-blue-600">
                          {webhook.name}
                        </Link>
                        <p className="text-xs text-slate-400 font-mono truncate max-w-xs">{webhook.url}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-1">
                      {webhook.events.slice(0, 2).map((event) => (
                        <span key={event} className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-xs font-medium">
                          {event}
                        </span>
                      ))}
                      {webhook.events.length > 2 && (
                        <span className="px-2 py-0.5 bg-slate-100 text-slate-500 rounded text-xs">
                          +{webhook.events.length - 2}
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    {webhook.is_active ? (
                      <span className="inline-flex items-center gap-1.5 text-green-600 text-sm">
                        <CheckCircle className="w-4 h-4" /> Active
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 text-slate-400 text-sm">
                        <XCircle className="w-4 h-4" /> Disabled
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4">
                    {webhook.consecutive_fails > 0 ? (
                      <span className="inline-flex items-center gap-1.5 text-red-600 text-sm font-medium">
                        <AlertTriangle className="w-4 h-4" />
                        {webhook.consecutive_fails}
                      </span>
                    ) : (
                      <span className="text-sm text-slate-400">—</span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm text-slate-600">
                    <div className="flex items-center gap-1">
                      <Clock className="w-4 h-4" />
                      {formatDate(webhook.created_at)}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <Link
                        to={`/webhooks/${webhook.id}`}
                        className="p-2 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                        title="View details"
                      >
                        <Eye className="w-4 h-4" />
                      </Link>
                      <button
                        onClick={() => handleToggle(webhook)}
                        className="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-lg transition-colors"
                        title={webhook.is_active ? 'Disable' : 'Enable'}
                      >
                        {webhook.is_active
                          ? <ToggleRight className="w-4 h-4 text-green-500" />
                          : <ToggleLeft className="w-4 h-4" />}
                      </button>
                      <button
                        onClick={() => handleDelete(webhook.id, webhook.name)}
                        className="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
