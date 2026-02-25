import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, X, Copy, CheckCircle } from 'lucide-react'
import { Link } from 'react-router-dom'
import { AgentService } from '@/services/api'
import { toast } from 'sonner'

const COMMON_SCOPES = ['read', 'write', 'admin', 'api:access', 'data:read', 'data:write']

export function CreateAgentPage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [scopes, setScopes] = useState<string[]>(['read'])
  const [customScope, setCustomScope] = useState('')
  const [created, setCreated] = useState<{ client_id: string; client_secret: string } | null>(null)

  const addScope = (scope: string) => {
    if (scope && !scopes.includes(scope)) {
      setScopes([...scopes, scope])
    }
    setCustomScope('')
  }

  const removeScope = (scope: string) => {
    setScopes(scopes.filter(s => s !== scope))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!name.trim()) {
      toast.error('Agent name is required')
      return
    }

    if (scopes.length === 0) {
      toast.error('At least one scope is required')
      return
    }

    setLoading(true)
    try {
      const response = await AgentService.create({
        name: name.trim(),
        description: description.trim() || undefined,
        scopes
      })
      setCreated({
        client_id: response.client_id,
        client_secret: response.client_secret
      })
      toast.success('Agent created successfully')
    } catch (error) {
      console.error('Failed to create agent:', error)
      toast.error('Failed to create agent')
    } finally {
      setLoading(false)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard')
  }

  if (created) {
    return (
      <div className="p-8 max-w-2xl mx-auto">
        <div className="bg-green-50 border border-green-200 rounded-xl p-6 mb-6">
          <div className="flex items-center gap-2 text-green-700 font-medium mb-2">
            <CheckCircle className="w-5 h-5" />
            Agent Created Successfully!
          </div>
          <p className="text-green-600 text-sm">
            Save these credentials now. The client_secret will not be shown again!
          </p>
        </div>

        <div className="bg-white rounded-xl border border-slate-200 p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Client ID
            </label>
            <div className="flex items-center gap-2">
              <code className="flex-1 px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg font-mono text-sm">
                {created.client_id}
              </code>
              <button
                onClick={() => copyToClipboard(created.client_id)}
                className="p-2.5 bg-slate-100 hover:bg-slate-200 rounded-lg"
              >
                <Copy className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Client Secret
            </label>
            <div className="flex items-center gap-2">
              <code className="flex-1 px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg font-mono text-sm">
                {created.client_secret}
              </code>
              <button
                onClick={() => copyToClipboard(created.client_secret)}
                className="p-2.5 bg-slate-100 hover:bg-slate-200 rounded-lg"
              >
                <Copy className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4 mt-6">
          <Link
            to="/agents"
            className="flex-1 px-4 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium rounded-lg text-center"
          >
            View Agents
          </Link>
          <button
            onClick={() => {
              setCreated(null)
              setName('')
              setDescription('')
              setScopes(['read'])
            }}
            className="flex-1 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg"
          >
            Create Another
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <Link
        to="/agents"
        className="inline-flex items-center gap-2 text-slate-600 hover:text-slate-900 mb-6"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Agents
      </Link>

      <div className="mb-8">
        <h1 className="text-2xl font-bold text-slate-900">Create Agent</h1>
        <p className="text-slate-500 mt-1">Add a new AI agent to your system</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Agent Name *
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., production-agent, staging-bot"
            className="w-full px-4 py-2.5 bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Description
          </label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Optional description for this agent"
            rows={3}
            className="w-full px-4 py-2.5 bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-2">
            Scopes *
          </label>
          
          {/* Common Scopes */}
          <div className="flex flex-wrap gap-2 mb-3">
            {COMMON_SCOPES.filter(s => !scopes.includes(s)).map((scope) => (
              <button
                key={scope}
                type="button"
                onClick={() => addScope(scope)}
                className="inline-flex items-center gap-1 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 text-sm rounded-lg transition-colors"
              >
                <Plus className="w-3 h-3" />
                {scope}
              </button>
            ))}
          </div>

          {/* Selected Scopes */}
          <div className="flex flex-wrap gap-2 mb-3">
            {scopes.map((scope) => (
              <span
                key={scope}
                className="inline-flex items-center gap-1 px-3 py-1.5 bg-blue-100 text-blue-700 text-sm rounded-lg"
              >
                {scope}
                <button
                  type="button"
                  onClick={() => removeScope(scope)}
                  className="hover:text-blue-900"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            ))}
          </div>

          {/* Custom Scope Input */}
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={customScope}
              onChange={(e) => setCustomScope(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  addScope(customScope.trim())
                }
              }}
              placeholder="Add custom scope"
              className="flex-1 px-4 py-2.5 bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <button
              type="button"
              onClick={() => addScope(customScope.trim())}
              className="px-4 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium rounded-lg"
            >
              Add
            </button>
          </div>
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Creating...' : 'Create Agent'}
        </button>
      </form>
    </div>
  )
}
