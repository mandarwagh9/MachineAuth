import { useState, useEffect } from 'react'
import { AgentService, TokenService } from '@/services/api'
import type { Agent } from '@/types'
import './token-generator.css'

export function TokenGenerator() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedAgent, setSelectedAgent] = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [scope, setScope] = useState('')
  const [token, setToken] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [requesting, setRequesting] = useState(false)

  useEffect(() => {
    loadAgents()
  }, [])

  const loadAgents = async () => {
    try {
      const data = await AgentService.list()
      setAgents(data)
      if (data.length > 0) {
        setSelectedAgent(data[0].client_id)
      }
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleRequest = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setToken(null)

    if (!clientSecret) {
      setError('Client secret is required')
      return
    }

    try {
      setRequesting(true)
      const response = await TokenService.request({
        grant_type: 'client_credentials',
        client_id: selectedAgent,
        client_secret: clientSecret,
        scope: scope || undefined,
      })

      setToken(response.access_token)
    } catch (err: unknown) {
      const errorMessage =
        err && typeof err === 'object' && 'response' in err
          ? (err.response as { data?: { error_description?: string } })?.data?.error_description || 'Failed to get token'
          : 'Failed to get token'
      setError(errorMessage)
    } finally {
      setRequesting(false)
    }
  }

  const copyToClipboard = () => {
    if (token) {
      navigator.clipboard.writeText(token)
    }
  }

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  return (
    <div className="token-generator">
      <h2>Get Access Token</h2>
      <p className="subtitle">
        Request an OAuth 2.0 access token using the Client Credentials flow.
      </p>

      {agents.length === 0 ? (
        <div className="empty-state">
          <p>No agents available. Create an agent first to get tokens.</p>
        </div>
      ) : (
        <form onSubmit={handleRequest} className="token-form">
          {error && <div className="error-message">{error}</div>}

          <div className="form-field">
            <label htmlFor="clientId">Client ID</label>
            <select
              id="clientId"
              value={selectedAgent}
              onChange={(e) => setSelectedAgent(e.target.value)}
            >
              {agents.map((agent) => (
                <option key={agent.client_id} value={agent.client_id}>
                  {agent.name} ({agent.client_id})
                </option>
              ))}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="clientSecret">Client Secret</label>
            <input
              id="clientSecret"
              type="password"
              value={clientSecret}
              onChange={(e) => setClientSecret(e.target.value)}
              placeholder="Enter client secret"
            />
          </div>

          <div className="form-field">
            <label htmlFor="scope">Scope (optional)</label>
            <input
              id="scope"
              type="text"
              value={scope}
              onChange={(e) => setScope(e.target.value)}
              placeholder="e.g., read:data write:data"
            />
            <span className="hint">
              Request specific scopes (must be subset of agent's scopes)
            </span>
          </div>

          <button type="submit" className="btn btn-primary" disabled={requesting}>
            {requesting ? 'Getting Token...' : 'Get Token'}
          </button>
        </form>
      )}

      {token && (
        <div className="token-result">
          <div className="token-header">
            <h3>Access Token</h3>
            <button onClick={copyToClipboard} className="btn btn-sm">
              Copy
            </button>
          </div>
          <code className="token-value">{token}</code>
        </div>
      )}
    </div>
  )
}
