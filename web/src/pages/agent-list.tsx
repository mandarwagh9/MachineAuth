import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { AgentService } from '@/services/api'
import type { Agent } from '@/types'
import './agent-list.css'

export function AgentList() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadAgents()
  }, [])

  const loadAgents = async () => {
    try {
      setLoading(true)
      const data = await AgentService.list()
      setAgents(data)
      setError(null)
    } catch (err) {
      setError('Failed to load agents')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this agent?')) return

    try {
      await AgentService.delete(id)
      setAgents(agents.filter((a) => a.id !== id))
    } catch (err) {
      alert('Failed to delete agent')
      console.error(err)
    }
  }

  if (loading) {
    return <div className="loading">Loading agents...</div>
  }

  if (error) {
    return <div className="error">{error}</div>
  }

  return (
    <div className="agent-list">
      <div className="page-header">
        <h2>AI Agents</h2>
        <Link to="/agents/new" className="btn btn-primary">
          Create Agent
        </Link>
      </div>

      {agents.length === 0 ? (
        <div className="empty-state">
          <p>No agents yet. Create your first AI agent to get started.</p>
          <Link to="/agents/new" className="btn btn-primary">
            Create Agent
          </Link>
        </div>
      ) : (
        <table className="agents-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Client ID</th>
              <th>Scopes</th>
              <th>Status</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {agents.map((agent) => (
              <tr key={agent.id}>
                <td>
                  <Link to={`/agents/${agent.id}`} className="agent-name">
                    {agent.name}
                  </Link>
                </td>
                <td>
                  <code className="client-id">{agent.client_id}</code>
                </td>
                <td>
                  <div className="scopes">
                    {agent.scopes.map((scope) => (
                      <span key={scope} className="scope-badge">
                        {scope}
                      </span>
                    ))}
                  </div>
                </td>
                <td>
                  <span className={`status ${agent.is_active ? 'active' : 'inactive'}`}>
                    {agent.is_active ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td>{new Date(agent.created_at).toLocaleDateString()}</td>
                <td>
                  <div className="actions">
                    <Link to={`/agents/${agent.id}`} className="btn btn-sm">
                      View
                    </Link>
                    <button
                      onClick={() => handleDelete(agent.id)}
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
