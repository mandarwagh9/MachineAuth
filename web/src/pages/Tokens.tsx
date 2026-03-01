import { useState } from 'react'
import { 
  Key, 
  Search, 
  XCircle, 
  CheckCircle, 
  Copy, 
  RefreshCw,
  AlertTriangle,
  ChevronDown,
  ChevronUp
} from 'lucide-react'
import { TokenService, AgentService } from '@/services/api'
import type { TokenResponse } from '@/types'
import { toast } from 'sonner'

type Tab = 'generate' | 'introspect' | 'revoke'

function JsonView({ data }: { data: unknown }) {
  return (
    <pre className="bg-slate-900 text-green-400 rounded-lg p-4 text-xs overflow-auto max-h-64 font-mono">
      {JSON.stringify(data, null, 2)}
    </pre>
  )
}

function CopyButton({ text }: { text: string }) {
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
      className="p-1.5 rounded-md hover:bg-slate-100 text-slate-500 hover:text-slate-700 transition-colors"
      title="Copy"
    >
      {copied ? <CheckCircle className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
    </button>
  )
}

// ── Generate Tab ───────────────────────────────────────────────────────────────

function GenerateTab() {
  const [clientId, setClientId] = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [scope, setScope] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TokenResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleGenerate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!clientId.trim() || !clientSecret.trim()) {
      toast.error('Client ID and Client Secret are required')
      return
    }
    setLoading(true)
    setError(null)
    setResult(null)
    try {
      const data = await TokenService.request({
        grant_type: 'client_credentials',
        client_id: clientId.trim(),
        client_secret: clientSecret.trim(),
        scope: scope.trim() || '',
      })
      setResult(data)
      toast.success('Token generated successfully')
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error_description?: string; error?: string } } })
        ?.response?.data?.error_description ||
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Failed to generate token'
      setError(msg)
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  const handleFetchAgents = async () => {
    try {
      const data = await AgentService.list()
      if (data.agents?.length > 0) {
        toast.info(`${data.agents.length} agent(s) available. Enter credentials manually.`)
      }
    } catch {
      toast.error('Failed to fetch agents')
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-slate-900">Generate Access Token</h2>
        <p className="text-sm text-slate-500 mt-1">
          Request an OAuth 2.0 access token using the Client Credentials flow.
        </p>
      </div>

      <form onSubmit={handleGenerate} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Client ID <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            value={clientId}
            onChange={(e) => setClientId(e.target.value)}
            placeholder="e.g. agent_abc123"
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Client Secret <span className="text-red-500">*</span>
          </label>
          <input
            type="password"
            value={clientSecret}
            onChange={(e) => setClientSecret(e.target.value)}
            placeholder="Your agent's client secret"
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Scope <span className="text-slate-400 font-normal">(optional)</span>
          </label>
          <input
            type="text"
            value={scope}
            onChange={(e) => setScope(e.target.value)}
            placeholder="e.g. read write"
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div className="flex gap-3">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Key className="w-4 h-4" />}
            {loading ? 'Generating...' : 'Generate Token'}
          </button>
          <button
            type="button"
            onClick={handleFetchAgents}
            className="flex items-center gap-2 px-4 py-2 border border-slate-300 text-slate-700 rounded-lg text-sm font-medium hover:bg-slate-50 transition-colors"
          >
            <Search className="w-4 h-4" />
            Browse Agents
          </button>
        </div>
      </form>

      {error && (
        <div className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg">
          <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-medium text-red-800">Token generation failed</p>
            <p className="text-sm text-red-600 mt-0.5">{error}</p>
          </div>
        </div>
      )}

      {result && (
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <CheckCircle className="w-5 h-5 text-green-500" />
            <h3 className="font-medium text-slate-900">Token Generated</h3>
          </div>

          <div className="bg-white border border-slate-200 rounded-xl p-4 space-y-3">
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium text-slate-500 mb-1">Access Token</p>
                <p className="text-xs font-mono text-slate-800 break-all">{result.access_token}</p>
              </div>
              <CopyButton text={result.access_token} />
            </div>
            <div className="grid grid-cols-3 gap-4 pt-3 border-t border-slate-100">
              <div>
                <p className="text-xs text-slate-500">Type</p>
                <p className="text-sm font-medium text-slate-800 capitalize">{result.token_type}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500">Expires In</p>
                <p className="text-sm font-medium text-slate-800">{result.expires_in}s</p>
              </div>
              {result.scope && (
                <div>
                  <p className="text-xs text-slate-500">Scope</p>
                  <p className="text-sm font-medium text-slate-800">{result.scope}</p>
                </div>
              )}
            </div>
            {result.refresh_token && (
              <div className="flex items-start justify-between gap-2 pt-3 border-t border-slate-100">
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-medium text-slate-500 mb-1">Refresh Token</p>
                  <p className="text-xs font-mono text-slate-800 break-all">{result.refresh_token}</p>
                </div>
                <CopyButton text={result.refresh_token} />
              </div>
            )}
          </div>

          <div>
            <p className="text-xs font-medium text-slate-500 mb-2">Raw Response</p>
            <JsonView data={result} />
          </div>
        </div>
      )}
    </div>
  )
}

// ── Introspect Tab ─────────────────────────────────────────────────────────────

function IntrospectTab() {
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<{ active: boolean; exp?: number; iat?: number; client_id?: string; scope?: string; token_type?: string; [key: string]: unknown } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [showRaw, setShowRaw] = useState(false)

  const handleIntrospect = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token.trim()) {
      toast.error('Token is required')
      return
    }
    setLoading(true)
    setError(null)
    setResult(null)
    try {
      const data = await TokenService.introspect(token.trim()) as { active: boolean; exp?: number; iat?: number; client_id?: string; scope?: string; token_type?: string; [key: string]: unknown }
      setResult(data)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error_description?: string } } })
        ?.response?.data?.error_description || 'Failed to introspect token'
      setError(msg)
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-slate-900">Introspect Token</h2>
        <p className="text-sm text-slate-500 mt-1">
          Check whether an access token is active and view its claims.
        </p>
      </div>

      <form onSubmit={handleIntrospect} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Access Token <span className="text-red-500">*</span>
          </label>
          <textarea
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="Paste your JWT access token here..."
            rows={4}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Search className="w-4 h-4" />}
          {loading ? 'Checking...' : 'Introspect Token'}
        </button>
      </form>

      {error && (
        <div className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg">
          <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {result && (
        <div className="space-y-4">
          <div className={`flex items-center gap-3 p-4 rounded-xl border ${
            result.active 
              ? 'bg-green-50 border-green-200' 
              : 'bg-red-50 border-red-200'
          }`}>
            {result.active 
              ? <CheckCircle className="w-6 h-6 text-green-500 flex-shrink-0" /> 
              : <XCircle className="w-6 h-6 text-red-500 flex-shrink-0" />}
            <div>
              <p className={`font-semibold ${result.active ? 'text-green-800' : 'text-red-800'}`}>
                {result.active ? 'Token is Active' : 'Token is Inactive or Revoked'}
              </p>
              {result.active && result.exp && (
                <p className="text-sm text-green-700 mt-0.5">
                  Expires: {new Date((result.exp as number) * 1000).toLocaleString()}
                </p>
              )}
            </div>
          </div>

          {result.active && (
            <div className="bg-white border border-slate-200 rounded-xl p-4 grid grid-cols-2 gap-4">
              {result.client_id && (
                <div>
                  <p className="text-xs text-slate-500">Client ID</p>
                  <p className="text-sm font-medium text-slate-800 font-mono">{result.client_id as string}</p>
                </div>
              )}
              {result.scope && (
                <div>
                  <p className="text-xs text-slate-500">Scope</p>
                  <p className="text-sm font-medium text-slate-800">{result.scope as string}</p>
                </div>
              )}
              {result.token_type && (
                <div>
                  <p className="text-xs text-slate-500">Token Type</p>
                  <p className="text-sm font-medium text-slate-800 capitalize">{result.token_type as string}</p>
                </div>
              )}
              {result.iat && (
                <div>
                  <p className="text-xs text-slate-500">Issued At</p>
                  <p className="text-sm font-medium text-slate-800">
                    {new Date((result.iat as number) * 1000).toLocaleString()}
                  </p>
                </div>
              )}
            </div>
          )}

          <button
            onClick={() => setShowRaw(!showRaw)}
            className="flex items-center gap-1 text-xs text-slate-500 hover:text-slate-700"
          >
            {showRaw ? <ChevronUp className="w-3 h-3" /> : <ChevronDown className="w-3 h-3" />}
            {showRaw ? 'Hide' : 'Show'} raw response
          </button>
          {showRaw && <JsonView data={result} />}
        </div>
      )}
    </div>
  )
}

// ── Revoke Tab ─────────────────────────────────────────────────────────────────

function RevokeTab() {
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleRevoke = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token.trim()) {
      toast.error('Token is required')
      return
    }
    if (!confirm('Are you sure you want to revoke this token? This action cannot be undone.')) return

    setLoading(true)
    setError(null)
    setSuccess(false)
    try {
      await TokenService.revoke(token.trim())
      setSuccess(true)
      setToken('')
      toast.success('Token revoked successfully')
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error_description?: string } } })
        ?.response?.data?.error_description || 'Failed to revoke token'
      setError(msg)
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-slate-900">Revoke Token</h2>
        <p className="text-sm text-slate-500 mt-1">
          Immediately invalidate an access or refresh token. This action cannot be undone.
        </p>
      </div>

      <div className="flex items-start gap-3 p-4 bg-amber-50 border border-amber-200 rounded-lg">
        <AlertTriangle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
        <p className="text-sm text-amber-800">
          Revoking a token will immediately prevent its further use. Any service relying on this token will lose access.
        </p>
      </div>

      <form onSubmit={handleRevoke} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Token <span className="text-red-500">*</span>
          </label>
          <textarea
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="Paste the access token or refresh token to revoke..."
            rows={4}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent resize-none"
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? <RefreshCw className="w-4 h-4 animate-spin" /> : <XCircle className="w-4 h-4" />}
          {loading ? 'Revoking...' : 'Revoke Token'}
        </button>
      </form>

      {error && (
        <div className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg">
          <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {success && (
        <div className="flex items-center gap-3 p-4 bg-green-50 border border-green-200 rounded-lg">
          <CheckCircle className="w-5 h-5 text-green-500" />
          <p className="text-sm font-medium text-green-800">Token revoked successfully.</p>
        </div>
      )}
    </div>
  )
}

// ── Main Page ──────────────────────────────────────────────────────────────────

const tabs: { id: Tab; label: string; description: string }[] = [
  { id: 'generate', label: 'Generate', description: 'Request a new access token' },
  { id: 'introspect', label: 'Introspect', description: 'Validate and inspect a token' },
  { id: 'revoke', label: 'Revoke', description: 'Invalidate a token' },
]

export function TokensPage() {
  const [activeTab, setActiveTab] = useState<Tab>('generate')

  return (
    <div className="p-8 max-w-3xl">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-slate-900">Token Tools</h1>
        <p className="text-slate-500 mt-1">Generate, inspect, and revoke OAuth 2.0 tokens</p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 mb-8 bg-slate-100 rounded-lg p-1 w-fit">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-white text-slate-900 shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="bg-white border border-slate-200 rounded-xl p-6">
        {activeTab === 'generate' && <GenerateTab />}
        {activeTab === 'introspect' && <IntrospectTab />}
        {activeTab === 'revoke' && <RevokeTab />}
      </div>
    </div>
  )
}
