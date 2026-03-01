import { useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { 
  ArrowLeft,
  Users, 
  Copy,
  CheckCircle,
  XCircle,
  Clock,
  RefreshCw,
  Trash2,
  AlertTriangle,
  Key,
  Activity,
  Shield,
  Eye,
  EyeOff
} from 'lucide-react'
import { AgentService } from '@/services/api'
import type { Agent } from '@/types'
import { toast } from 'sonner'

function CopyButton({ text, label = 'Copy' }: { text: string; label?: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = () => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    toast.success('Copied to clipboard')
    setTimeout(() => setCopied(false), 2000)
  }
  return (
    <button
      onClick={handleCopy}
      className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-slate-600 border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
    >
      {copied ? <CheckCircle className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
      {copied ? 'Copied!' : label}
    </button>
  )
}

function InfoRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start py-3 border-b border-slate-100 last:border-0">
      <span className="w-40 text-sm text-slate-500 flex-shrink-0">{label}</span>
      <div className="flex-1 text-sm text-slate-900">{children}</div>
    </div>
  )
}

export function AgentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [agent, setAgent] = useState<Agent | null>(null)
  const [loading, setLoading] = useState(true)
  const [rotating, setRotating] = useState(false)
  const [deactivating, setDeactivating] = useState(false)
  const [newSecret, setNewSecret] = useState<string | null>(null)
  const [showSecret, setShowSecret] = useState(false)

  useEffect(() => {
    if (id) fetchAgent()
  }, [id])

  const fetchAgent = async () => {
    try {
      const data = await AgentService.get(id!)
      setAgent(data)
    } catch (error) {
      console.error('Failed to fetch agent:', error)
      toast.error('Agent not found')
      navigate('/agents')
    } finally {
      setLoading(false)
    }
  }

  const handleRotate = async () => {
    if (!confirm('Rotate credentials? The current client secret will be invalidated immediately.')) return
    setRotating(true)
    try {
      const data = await AgentService.rotate(id!)
      if (data.client_secret) {
        setNewSecret(data.client_secret)
        setShowSecret(true)
      }
      toast.success('Credentials rotated successfully')
      fetchAgent()
    } catch {
      toast.error('Failed to rotate credentials')
    } finally {
      setRotating(false)
    }
  }

  const handleDeactivate = async () => {
    if (!agent) return
    const action = agent.is_active ? 'deactivate' : 'reactivate'
    if (!confirm(`Are you sure you want to ${action} this agent?`)) return
    setDeactivating(true)
    try {
      await AgentService.deactivate(id!)
      toast.success(`Agent ${action}d successfully`)
      fetchAgent()
    } catch {
      toast.error(`Failed to ${action} agent`)
    } finally {
      setDeactivating(false)
    }
  }

  const formatDate = (dateString?: string) => {
    if (!dateString) return '—'
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit'
    })
  }

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="h-64 bg-slate-200 rounded-xl" />
          <div className="h-40 bg-slate-200 rounded-xl" />
        </div>
      </div>
    )
  }

  if (!agent) return null

  return (
    <div className="p-8 max-w-4xl">
      {/* Header */}
      <div className="mb-8">
        <Link
          to="/agents"
          className="inline-flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-700 mb-4"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Agents
        </Link>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 bg-blue-100 rounded-xl flex items-center justify-center">
              <Users className="w-7 h-7 text-blue-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-slate-900">{agent.name}</h1>
              {agent.description && (
                <p className="text-slate-500 mt-0.5">{agent.description}</p>
              )}
              <div className="flex items-center gap-2 mt-2">
                {agent.is_active ? (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-green-100 text-green-700 rounded-full text-xs font-medium">
                    <CheckCircle className="w-3 h-3" />
                    Active
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-red-100 text-red-700 rounded-full text-xs font-medium">
                    <XCircle className="w-3 h-3" />
                    Inactive
                  </span>
                )}
                {agent.expires_at && (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-amber-100 text-amber-700 rounded-full text-xs font-medium">
                    <Clock className="w-3 h-3" />
                    Expires {formatDate(agent.expires_at)}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex items-center gap-3">
            <button
              onClick={handleRotate}
              disabled={rotating}
              className="flex items-center gap-2 px-4 py-2 border border-slate-300 text-slate-700 rounded-lg text-sm font-medium hover:bg-slate-50 disabled:opacity-50 transition-colors"
            >
              {rotating ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Key className="w-4 h-4" />}
              Rotate Credentials
            </button>
            <button
              onClick={handleDeactivate}
              disabled={deactivating}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium disabled:opacity-50 transition-colors ${
                agent.is_active
                  ? 'bg-red-50 text-red-700 border border-red-200 hover:bg-red-100'
                  : 'bg-green-50 text-green-700 border border-green-200 hover:bg-green-100'
              }`}
            >
              {deactivating 
                ? <RefreshCw className="w-4 h-4 animate-spin" />
                : agent.is_active ? <XCircle className="w-4 h-4" /> : <CheckCircle className="w-4 h-4" />}
              {agent.is_active ? 'Deactivate' : 'Reactivate'}
            </button>
          </div>
        </div>
      </div>

      {/* New Secret Banner */}
      {newSecret && (
        <div className="mb-6 p-4 bg-amber-50 border border-amber-300 rounded-xl">
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-semibold text-amber-800">Save your new client secret</p>
              <p className="text-sm text-amber-700 mt-0.5">This is shown only once. Copy it now.</p>
              <div className="mt-3 flex items-center gap-2">
                <code className="flex-1 text-sm font-mono bg-white border border-amber-200 rounded-lg px-3 py-2 text-slate-800">
                  {showSecret ? newSecret : '•'.repeat(40)}
                </code>
                <button
                  onClick={() => setShowSecret(!showSecret)}
                  className="p-2 text-amber-600 hover:text-amber-800"
                >
                  {showSecret ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
                <CopyButton text={newSecret} label="Copy Secret" />
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Details */}
        <div className="lg:col-span-2 space-y-6">
          {/* Credentials */}
          <div className="bg-white border border-slate-200 rounded-xl p-6">
            <div className="flex items-center gap-2 mb-4">
              <Shield className="w-5 h-5 text-slate-700" />
              <h2 className="font-semibold text-slate-900">Credentials</h2>
            </div>
            <div className="space-y-3">
              <div>
                <p className="text-xs font-medium text-slate-500 mb-1">Agent ID</p>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-sm font-mono text-slate-600 bg-slate-50 border border-slate-200 rounded-lg px-3 py-2">{agent.id}</code>
                  <CopyButton text={agent.id} />
                </div>
              </div>
              <div>
                <p className="text-xs font-medium text-slate-500 mb-1">Client ID</p>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-sm font-mono text-slate-600 bg-slate-50 border border-slate-200 rounded-lg px-3 py-2">{agent.client_id}</code>
                  <CopyButton text={agent.client_id} />
                </div>
              </div>
            </div>
          </div>

          {/* Configuration */}
          <div className="bg-white border border-slate-200 rounded-xl p-6">
            <h2 className="font-semibold text-slate-900 mb-4">Configuration</h2>
            <div>
              <InfoRow label="Status">
                {agent.is_active ? (
                  <span className="inline-flex items-center gap-1.5 text-green-600">
                    <CheckCircle className="w-4 h-4" /> Active
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 text-red-600">
                    <XCircle className="w-4 h-4" /> Inactive
                  </span>
                )}
              </InfoRow>
              <InfoRow label="Scopes">
                <div className="flex flex-wrap gap-1">
                  {agent.scopes.length > 0 ? agent.scopes.map((s) => (
                    <span key={s} className="px-2 py-0.5 bg-blue-50 text-blue-700 rounded text-xs font-medium">{s}</span>
                  )) : (
                    <span className="text-slate-400">No scopes defined</span>
                  )}
                </div>
              </InfoRow>
              <InfoRow label="Created">{formatDate(agent.created_at)}</InfoRow>
              <InfoRow label="Updated">{formatDate(agent.updated_at)}</InfoRow>
              <InfoRow label="Expires">{formatDate(agent.expires_at)}</InfoRow>
            </div>
          </div>
        </div>

        {/* Stats sidebar */}
        <div className="space-y-4">
          <div className="bg-white border border-slate-200 rounded-xl p-5">
            <div className="flex items-center gap-2 mb-4">
              <Activity className="w-5 h-5 text-slate-700" />
              <h2 className="font-semibold text-slate-900">Usage</h2>
            </div>
            <div className="space-y-4">
              <div className="text-center p-3 bg-blue-50 rounded-lg">
                <p className="text-3xl font-bold text-blue-700">{agent.token_count ?? 0}</p>
                <p className="text-xs text-blue-600 mt-1">Tokens Issued</p>
              </div>
              <div className="text-center p-3 bg-purple-50 rounded-lg">
                <p className="text-3xl font-bold text-purple-700">{agent.refresh_count ?? 0}</p>
                <p className="text-xs text-purple-600 mt-1">Tokens Refreshed</p>
              </div>
            </div>
          </div>

          <div className="bg-white border border-slate-200 rounded-xl p-5">
            <h2 className="font-semibold text-slate-900 mb-3">Recent Activity</h2>
            <div className="space-y-2">
              <div>
                <p className="text-xs text-slate-500">Last Token Issued</p>
                <p className="text-sm text-slate-700 mt-0.5">{formatDate(agent.last_token_issued_at)}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500">Last Activity</p>
                <p className="text-sm text-slate-700 mt-0.5">{formatDate(agent.last_activity_at)}</p>
              </div>
            </div>
          </div>

          {/* Danger Zone */}
          <div className="bg-white border border-red-200 rounded-xl p-5">
            <div className="flex items-center gap-2 mb-3">
              <Trash2 className="w-4 h-4 text-red-500" />
              <h2 className="font-semibold text-red-700">Danger Zone</h2>
            </div>
            <p className="text-xs text-slate-500 mb-3">
              Permanently delete this agent and all its data. This cannot be undone.
            </p>
            <button
              onClick={async () => {
                if (!confirm(`Permanently delete "${agent.name}"? This cannot be undone.`)) return
                try {
                  // Use deactivate as "soft delete" – backend doesn't expose hard delete via admin yet
                  await AgentService.deactivate(id!)
                  toast.success('Agent deleted')
                  navigate('/agents')
                } catch {
                  toast.error('Failed to delete agent')
                }
              }}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors"
            >
              <Trash2 className="w-4 h-4" />
              Delete Agent
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
