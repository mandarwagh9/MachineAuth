import { useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { 
  ArrowLeft, 
  Building2, 
  Users, 
  Key, 
  Settings,
  Plus,
  Trash2,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import { OrganizationService } from '@/services/api'
import type { Organization, Team, Agent, APIKey } from '@/types'
import { toast } from 'sonner'

type Tab = 'overview' | 'teams' | 'agents' | 'api-keys'

export function OrganizationDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [organization, setOrganization] = useState<Organization | null>(null)
  const [teams, setTeams] = useState<Team[]>([])
  const [agents, setAgents] = useState<Agent[]>([])
  const [apiKeys, setAPIKeys] = useState<APIKey[]>([])
  const [activeTab, setActiveTab] = useState<Tab>('overview')
  const [loading, setLoading] = useState(true)
  const [showCreateTeam, setShowCreateTeam] = useState(false)
  const [showCreateAPIKey, setShowCreateAPIKey] = useState(false)
  const [newTeamName, setNewTeamName] = useState('')
  const [newTeamDesc, setNewTeamDesc] = useState('')
  const [newAPIKeyName, setNewAPIKeyName] = useState('')

  useEffect(() => {
    if (id) {
      fetchOrganization()
    }
  }, [id])

  useEffect(() => {
    if (activeTab === 'teams') {
      fetchTeams()
    } else if (activeTab === 'agents') {
      fetchAgents()
    } else if (activeTab === 'api-keys') {
      fetchAPIKeys()
    }
  }, [activeTab, id])

  const fetchOrganization = async () => {
    try {
      const data = await OrganizationService.get(id!)
      setOrganization(data.organization || data)
    } catch (error) {
      console.error('Failed to fetch organization:', error)
      toast.error('Failed to load organization')
      navigate('/organizations')
    } finally {
      setLoading(false)
    }
  }

  const fetchTeams = async () => {
    try {
      const data = await OrganizationService.listTeams(id!)
      setTeams(data.teams || [])
    } catch (error) {
      console.error('Failed to fetch teams:', error)
      toast.error('Failed to load teams')
    }
  }

  const fetchAgents = async () => {
    try {
      const data = await OrganizationService.listAgents(id!)
      setAgents(data.agents || [])
    } catch (error) {
      console.error('Failed to fetch agents:', error)
      toast.error('Failed to load agents')
    }
  }

  const fetchAPIKeys = async () => {
    try {
      const data = await OrganizationService.listAPIKeys(id!)
      setAPIKeys(data.api_keys || [])
    } catch (error) {
      console.error('Failed to fetch API keys:', error)
      toast.error('Failed to load API keys')
    }
  }

  const handleCreateTeam = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await OrganizationService.createTeam(id!, { name: newTeamName, description: newTeamDesc })
      toast.success('Team created successfully')
      setShowCreateTeam(false)
      setNewTeamName('')
      setNewTeamDesc('')
      fetchTeams()
    } catch (error) {
      console.error('Failed to create team:', error)
      toast.error('Failed to create team')
    }
  }

  const handleCreateAPIKey = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const response = await OrganizationService.createAPIKey(id!, { name: newAPIKeyName })
      toast.success('API Key created - copy it now, it won\'t be shown again!')
      navigator.clipboard.writeText(response.key)
      setShowCreateAPIKey(false)
      setNewAPIKeyName('')
      fetchAPIKeys()
    } catch (error) {
      console.error('Failed to create API key:', error)
      toast.error('Failed to create API key')
    }
  }

  const handleDeleteAPIKey = async (keyId: string) => {
    if (!confirm('Are you sure you want to revoke this API key?')) return
    try {
      await OrganizationService.deleteAPIKey(id!, keyId)
      toast.success('API key revoked')
      fetchAPIKeys()
    } catch (error) {
      console.error('Failed to delete API key:', error)
      toast.error('Failed to revoke API key')
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-slate-200 rounded w-64" />
          <div className="h-64 bg-slate-200 rounded-xl" />
        </div>
      </div>
    )
  }

  if (!organization) {
    return null
  }

  const tabs = [
    { id: 'overview', label: 'Overview', icon: Settings },
    { id: 'teams', label: 'Teams', icon: Users },
    { id: 'agents', label: 'Agents', icon: Users },
    { id: 'api-keys', label: 'API Keys', icon: Key },
  ]

  return (
    <div className="p-8">
      <button
        onClick={() => navigate('/organizations')}
        className="flex items-center gap-2 text-slate-600 hover:text-slate-900 mb-6"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Organizations
      </button>

      {/* Header */}
      <div className="flex items-center gap-4 mb-8">
        <div className="w-16 h-16 bg-indigo-100 rounded-xl flex items-center justify-center">
          <Building2 className="w-8 h-8 text-indigo-600" />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-slate-900">{organization.name}</h1>
          <p className="text-slate-500 font-mono">{organization.slug}</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-slate-200 mb-6">
        <nav className="flex gap-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as Tab)}
              className={`flex items-center gap-2 py-3 border-b-2 font-medium text-sm transition-colors ${
                activeTab === tab.id
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-slate-500 hover:text-slate-700'
              }`}
            >
              <tab.icon className="w-4 h-4" />
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <h2 className="text-lg font-semibold mb-4">Organization Details</h2>
          <dl className="grid grid-cols-2 gap-4">
            <div>
              <dt className="text-sm text-slate-500">Name</dt>
              <dd className="text-slate-900 font-medium">{organization.name}</dd>
            </div>
            <div>
              <dt className="text-sm text-slate-500">Slug</dt>
              <dd className="text-slate-900 font-mono">{organization.slug}</dd>
            </div>
            <div>
              <dt className="text-sm text-slate-500">Owner Email</dt>
              <dd className="text-slate-900">{organization.owner_email || '-'}</dd>
            </div>
            <div>
              <dt className="text-sm text-slate-500">Created</dt>
              <dd className="text-slate-900">{formatDate(organization.created_at)}</dd>
            </div>
          </dl>
        </div>
      )}

      {activeTab === 'teams' && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button
              onClick={() => setShowCreateTeam(true)}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg"
            >
              <Plus className="w-4 h-4" />
              Add Team
            </button>
          </div>

          {showCreateTeam && (
            <div className="bg-white rounded-xl border border-slate-200 p-4">
              <form onSubmit={handleCreateTeam} className="flex gap-4 items-end">
                <div className="flex-1">
                  <input
                    type="text"
                    placeholder="Team name"
                    value={newTeamName}
                    onChange={(e) => setNewTeamName(e.target.value)}
                    required
                    className="w-full px-3 py-2 border border-slate-200 rounded-lg"
                  />
                </div>
                <div className="flex-1">
                  <input
                    type="text"
                    placeholder="Description (optional)"
                    value={newTeamDesc}
                    onChange={(e) => setNewTeamDesc(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-200 rounded-lg"
                  />
                </div>
                <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded-lg">
                  Create
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreateTeam(false)}
                  className="px-4 py-2 bg-slate-100 rounded-lg"
                >
                  Cancel
                </button>
              </form>
            </div>
          )}

          <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
            {teams.length === 0 ? (
              <div className="p-8 text-center text-slate-500">
                <Users className="w-12 h-12 mx-auto mb-3 text-slate-300" />
                <p>No teams yet</p>
              </div>
            ) : (
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Name</th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Description</th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Created</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {teams.map((team) => (
                    <tr key={team.id} className="hover:bg-slate-50">
                      <td className="px-6 py-4 font-medium text-slate-900">{team.name}</td>
                      <td className="px-6 py-4 text-slate-600">{team.description || '-'}</td>
                      <td className="px-6 py-4 text-slate-600 text-sm">{formatDate(team.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {activeTab === 'agents' && (
        <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
          {agents.length === 0 ? (
            <div className="p-8 text-center text-slate-500">
              <Users className="w-12 h-12 mx-auto mb-3 text-slate-300" />
              <p>No agents in this organization</p>
              <Link to={`/agents/new?org=${id}`} className="text-blue-600 hover:underline mt-2 inline-block">
                Create an agent
              </Link>
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-slate-50 border-b border-slate-200">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Name</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Client ID</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Scopes</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200">
                {agents.map((agent) => (
                  <tr key={agent.id} className="hover:bg-slate-50">
                    <td className="px-6 py-4 font-medium text-slate-900">{agent.name}</td>
                    <td className="px-6 py-4">
                      <code className="text-sm font-mono bg-slate-100 px-2 py-1 rounded">
                        {agent.client_id.slice(0, 12)}...
                      </code>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {agent.scopes.map((scope) => (
                          <span key={scope} className="px-2 py-0.5 bg-slate-100 text-slate-600 text-xs rounded">
                            {scope}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {agent.is_active ? (
                        <span className="text-green-600 flex items-center gap-1">
                          <CheckCircle className="w-4 h-4" /> Active
                        </span>
                      ) : (
                        <span className="text-red-600 flex items-center gap-1">
                          <XCircle className="w-4 h-4" /> Inactive
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-slate-600 text-sm">{formatDate(agent.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === 'api-keys' && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button
              onClick={() => setShowCreateAPIKey(true)}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg"
            >
              <Plus className="w-4 h-4" />
              Create API Key
            </button>
          </div>

          {showCreateAPIKey && (
            <div className="bg-white rounded-xl border border-slate-200 p-4">
              <form onSubmit={handleCreateAPIKey} className="flex gap-4 items-end">
                <div className="flex-1">
                  <input
                    type="text"
                    placeholder="Key name (e.g., production)"
                    value={newAPIKeyName}
                    onChange={(e) => setNewAPIKeyName(e.target.value)}
                    required
                    className="w-full px-3 py-2 border border-slate-200 rounded-lg"
                  />
                </div>
                <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded-lg">
                  Create
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreateAPIKey(false)}
                  className="px-4 py-2 bg-slate-100 rounded-lg"
                >
                  Cancel
                </button>
              </form>
            </div>
          )}

          <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
            {apiKeys.length === 0 ? (
              <div className="p-8 text-center text-slate-500">
                <Key className="w-12 h-12 mx-auto mb-3 text-slate-300" />
                <p>No API keys</p>
              </div>
            ) : (
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Name</th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Prefix</th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Status</th>
                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500">Created</th>
                    <th className="px-6 py-3 text-right text-xs font-semibold text-slate-500">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {apiKeys.map((key) => (
                    <tr key={key.id} className="hover:bg-slate-50">
                      <td className="px-6 py-4 font-medium text-slate-900">{key.name}</td>
                      <td className="px-6 py-4">
                        <code className="text-sm font-mono bg-slate-100 px-2 py-1 rounded">
                          {key.prefix}
                        </code>
                      </td>
                      <td className="px-6 py-4">
                        {key.is_active ? (
                          <span className="text-green-600 flex items-center gap-1">
                            <CheckCircle className="w-4 h-4" /> Active
                          </span>
                        ) : (
                          <span className="text-red-600 flex items-center gap-1">
                            <XCircle className="w-4 h-4" /> Revoked
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4 text-slate-600 text-sm">{formatDate(key.created_at)}</td>
                      <td className="px-6 py-4 text-right">
                        <button
                          onClick={() => handleDeleteAPIKey(key.id)}
                          className="p-2 text-red-600 hover:bg-red-50 rounded-lg"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
