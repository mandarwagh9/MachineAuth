import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { AgentService } from '@/services/api'
import type { Agent } from '@/types'
import './agent-detail.css'

export function AgentDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [agent, setAgent] = useState<Agent | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [rotating, setRotating] = useState(false)
  const [newSecret, setNewSecret] = useState<string | null>(null)

  useEffect(() => {
    if (id) {
      loadAgent()
    }
  }, [id])

  const loadAgent = async () => {
    try {
      setLoading(true)
      const data = await AgentService.get(id!)
      setAgent(data)
      setError(null)
    } catch (err) {
      setError('Failed to load agent')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleRotate = async () => {
    if (!confirm('This will invalidate the current client secret. Continue?')) {
      return
    }

    try {
      setRotating(true)
      const result = await AgentService.rotate(id!)
      setNewSecret(result.client_secret)
    } catch (err) {
      alert('Failed to rotate credentials')
      console.error(err)
    } finally {
      setRotating(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this agent? This cannot be undone.')) {
      return
    }

    try {
      await AgentService.delete(id!)
      navigate('/agents')
    } catch (err) {
      alert('Failed to delete agent')
      console.error(err)
    }
  }

  if (loading) {
    return <div className="loading">Loading agent...</div>
  }

  if (error || !agent) {
    return <div className="error">{error || 'Agent not found'}</div>
  }

  return (
    <div className="agent-detail">
      <div className="page-header">
        <button onClick={() => navigate('/agents')} className="back-link">
          ← Back to Agents
        </button>
      </div>

      {newSecret && (
        <div className="secret-alert">
          <h3>New Client Secret Generated</h3>
          <p>Copy this secret now. You won't see it again!</p>
          <code>{newSecret}</code>
          <button onClick={() => setNewSecret(null)} className="btn btn-sm">
            Dismiss
          </button>
        </div>
      )}

      <div className="detail-card">
        <div className="detail-header">
          <h2>{agent.name}</h2>
          <span className={`status ${agent.is_active ? 'active' : 'inactive'}`}>
            {agent.is_active ? 'Active' : 'Inactive'}
          </span>
        </div>

        <div className="detail-grid">
          <div className="detail-item">
            <label>Agent ID</label>
            <code>{agent.id}</code>
          </div>

          <div className="detail-item">
            <label>Client ID</label>
            <code>{agent.client_id}</code>
          </div>

          <div className="detail-item">
            <label>Created At</label>
            <span>{new Date(agent.created_at).toLocaleString()}</span>
          </div>

          <div className="detail-item">
            <label>Updated At</label>
            <span>{new Date(agent.updated_at).toLocaleString()}</span>
          </div>

          {agent.expires_at && (
            <div className="detail-item">
              <label>Expires At</label>
              <span>{new Date(agent.expires_at).toLocaleString()}</span>
            </div>
          )}

          <div className="detail-item full-width">
            <label>Scopes</label>
            <div className="scopes">
              {agent.scopes.length > 0 ? (
                agent.scopes.map((scope) => (
                  <span key={scope} className="scope-badge">
                    {scope}
                  </span>
                ))
              ) : (
                <span className="no-scopes">No scopes assigned</span>
              )}
            </div>
          </div>
        </div>

        <div className="detail-actions">
          <button
            onClick={handleRotate}
            className="btn btn-secondary"
            disabled={rotating}
          >
            {rotating ? 'Rotating...' : 'Rotate Credentials'}
          </button>
          <button onClick={handleDelete} className="btn btn-danger">
            Delete Agent
          </button>
        </div>
      </div>
    </div>
  )
}
