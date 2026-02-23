import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AgentService } from '@/services/api'
import './create-agent.css'

export function CreateAgent() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [scopes, setScopes] = useState('')
  const [expiresIn, setExpiresIn] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [createdAgent, setCreatedAgent] = useState<{
    client_id: string
    client_secret: string
  } | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      setError('Name is required')
      return
    }

    try {
      setLoading(true)
      setError(null)

      const scopesArray = scopes
        .split(',')
        .map((s) => s.trim())
        .filter((s) => s)

      const expiresInNum = expiresIn ? parseInt(expiresIn, 10) : undefined

      const response = await AgentService.create({
        name: name.trim(),
        scopes: scopesArray.length > 0 ? scopesArray : undefined,
        expires_in: expiresInNum,
      })

      setCreatedAgent({
        client_id: response.client_id,
        client_secret: response.client_secret,
      })
    } catch (err) {
      setError('Failed to create agent')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  if (createdAgent) {
    return (
      <div className="create-agent">
        <div className="success-card">
          <h2>Agent Created Successfully</h2>
          <p className="warning">
            Copy your client secret now. You won't be able to see it again!
          </p>

          <div className="credential-field">
            <label>Client ID</label>
            <code>{createdAgent.client_id}</code>
          </div>

          <div className="credential-field">
            <label>Client Secret</label>
            <code className="secret">{createdAgent.client_secret}</code>
          </div>

          <div className="actions">
            <button
              onClick={() => navigate('/agents')}
              className="btn btn-primary"
            >
              View Agents
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="create-agent">
      <div className="form-card">
        <h2>Create New AI Agent</h2>

        <form onSubmit={handleSubmit}>
          {error && <div className="error-message">{error}</div>}

          <div className="form-field">
            <label htmlFor="name">Agent Name</label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-ai-agent"
              disabled={loading}
            />
          </div>

          <div className="form-field">
            <label htmlFor="scopes">Scopes (comma-separated)</label>
            <input
              id="scopes"
              type="text"
              value={scopes}
              onChange={(e) => setScopes(e.target.value)}
              placeholder="read:data, write:data"
              disabled={loading}
            />
            <span className="hint">
              Optional. Define what permissions this agent has.
            </span>
          </div>

          <div className="form-field">
            <label htmlFor="expiresIn">Expires In (seconds)</label>
            <input
              id="expiresIn"
              type="number"
              value={expiresIn}
              onChange={(e) => setExpiresIn(e.target.value)}
              placeholder="86400"
              disabled={loading}
            />
            <span className="hint">
              Optional. Leave empty for no expiration.
            </span>
          </div>

          <div className="form-actions">
            <button
              type="button"
              onClick={() => navigate('/agents')}
              className="btn btn-secondary"
              disabled={loading}
            >
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? 'Creating...' : 'Create Agent'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
