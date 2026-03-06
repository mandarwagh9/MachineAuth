import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { 
  Users, 
  Key, 
  Activity, 
  Shield,
  Clock,
  AlertCircle,
  RefreshCw,
  Plus,
  ArrowRight,
  Zap,
  CheckCircle2,
  XCircle
} from 'lucide-react'
import { MetricsService, HealthService, AgentService, OrganizationService } from '@/services/api'
import type { Metrics, HealthCheck, Agent } from '@/types'
import { useOrg } from '@/contexts/OrgContext'

const StatCard = ({ 
  title, 
  value, 
  subtitle,
  icon: Icon, 
  color,
  href
}: { 
  title: string
  value: string | number
  subtitle?: string
  icon: React.ElementType
  color: string
  href?: string 
}) => {
  const content = (
    <div className={`bg-white rounded-xl border border-slate-200 p-6 transition-all ${href ? 'hover:shadow-md hover:border-slate-300 cursor-pointer' : ''}`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-slate-500">{title}</p>
          <p className="mt-2 text-3xl font-bold text-slate-900">{value}</p>
          {subtitle && (
            <p className="mt-1 text-sm text-slate-400">{subtitle}</p>
          )}
        </div>
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${color}`}>
          <Icon className="w-6 h-6 text-white" />
        </div>
      </div>
    </div>
  )
  if (href) return <Link to={href}>{content}</Link>
  return content
}

export function Dashboard() {
  const [metrics, setMetrics] = useState<Metrics | null>(null)
  const [health, setHealth] = useState<HealthCheck | null>(null)
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())
  const { currentOrg } = useOrg()

  const fetchData = async () => {
    try {
      setError(null)
      
      let agentsData: { agents?: Agent[] } = { agents: [] }
      
      // If org is selected, fetch org-specific data
      if (currentOrg) {
        agentsData = await OrganizationService.listAgents(currentOrg.id).catch(() => ({ agents: [] }))
      } else {
        agentsData = await AgentService.list().catch(() => ({ agents: [] }))
      }
      
      const [metricsData, healthData] = await Promise.all([
        MetricsService.get().catch(() => null),
        HealthService.ready().catch(() => null),
      ])
      
      setMetrics(metricsData)
      setHealth(healthData)
      setAgents(agentsData?.agents || [])
      setLastRefresh(new Date())
      if (!metricsData && !healthData) {
        setError('Unable to connect to MachineAuth backend')
      }
    } catch (err) {
      setError('Failed to fetch dashboard data')
      console.error('Dashboard fetch error:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 30000)
    return () => clearInterval(interval)
  }, [currentOrg])

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="h-16 bg-slate-200 rounded-xl" />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {[1,2,3,4].map(i => (
              <div key={i} className="h-32 bg-slate-200 rounded-xl" />
            ))}
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="h-64 bg-slate-200 rounded-xl" />
            <div className="h-64 bg-slate-200 rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  const activeAgents = agents.filter(a => a.is_active).length
  const expiredAgents = agents.filter(a => a.expires_at && new Date(a.expires_at) < new Date()).length
  const recentAgents = agents
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5)

  return (
    <div className="p-8 max-w-7xl">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>
          <p className="text-slate-500 mt-1">Overview of your MachineAuth instance</p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-slate-400">
            Updated {lastRefresh.toLocaleTimeString()}
          </span>
          <button
            onClick={() => { setLoading(true); fetchData() }}
            className="p-2 bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors"
            title="Refresh"
          >
            <RefreshCw className="w-4 h-4 text-slate-600" />
          </button>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 rounded-xl p-4 flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-red-500 flex-shrink-0" />
          <div>
            <span className="text-red-700 font-medium">{error}</span>
            <p className="text-red-600 text-sm mt-0.5">Check that the backend is running at auth.writesomething.fun</p>
          </div>
          <button onClick={fetchData} className="ml-auto text-red-600 hover:text-red-800 text-sm font-medium">
            Retry
          </button>
        </div>
      )}

      {/* Status Banner */}
      {!error && (
        <div className="mb-6 bg-green-50 border border-green-200 rounded-xl p-4 flex items-center gap-3">
          <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse" />
          <span className="text-green-700 font-medium">System Online</span>
          <span className="text-green-600">·</span>
          <span className="text-green-600">{health?.agents_count || agents.length} agents registered</span>
          <span className="text-green-600">·</span>
          <span className="text-green-600">{activeAgents} active</span>
        </div>
      )}

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatCard
          title="Total Agents"
          value={metrics?.total_agents || agents.length}
          subtitle={`${activeAgents} active · ${agents.length - activeAgents} inactive`}
          icon={Users}
          color="bg-blue-500"
          href="/agents"
        />
        <StatCard
          title="Active Tokens"
          value={metrics?.active_tokens || 0}
          subtitle="Currently valid tokens"
          icon={Key}
          color="bg-emerald-500"
        />
        <StatCard
          title="Tokens Issued"
          value={metrics?.tokens_issued || 0}
          subtitle={`${metrics?.tokens_refreshed || 0} refreshed`}
          icon={Shield}
          color="bg-purple-500"
          href="/metrics"
        />
        <StatCard
          title="Tokens Revoked"
          value={metrics?.tokens_revoked || 0}
          subtitle="Explicitly invalidated"
          icon={Activity}
          color="bg-orange-500"
          href="/metrics"
        />
      </div>

      {/* Two Column Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Token Activity */}
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <div className="flex items-center justify-between mb-5">
            <h3 className="font-semibold text-slate-900">Token Activity</h3>
            <Link to="/metrics" className="text-sm text-blue-600 hover:text-blue-800 flex items-center gap-1">
              View all <ArrowRight className="w-3 h-3" />
            </Link>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <Zap className="w-4 h-4 text-green-500" />
                Tokens Issued
              </span>
              <span className="font-semibold text-slate-900 text-lg">{(metrics?.tokens_issued || 0).toLocaleString()}</span>
            </div>
            <div className="border-t border-slate-100" />
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <RefreshCw className="w-4 h-4 text-blue-500" />
                Tokens Refreshed
              </span>
              <span className="font-semibold text-slate-900 text-lg">{(metrics?.tokens_refreshed || 0).toLocaleString()}</span>
            </div>
            <div className="border-t border-slate-100" />
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <XCircle className="w-4 h-4 text-red-500" />
                Tokens Revoked
              </span>
              <span className="font-semibold text-slate-900 text-lg">{(metrics?.tokens_revoked || 0).toLocaleString()}</span>
            </div>
            <div className="border-t border-slate-100" />
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <Key className="w-4 h-4 text-emerald-500" />
                Active Tokens
              </span>
              <span className="font-semibold text-emerald-600 text-lg">{(metrics?.active_tokens || 0).toLocaleString()}</span>
            </div>
          </div>
        </div>

        {/* System Health */}
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <h3 className="font-semibold text-slate-900 mb-5">System Health</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <Activity className="w-4 h-4" />
                Backend Status
              </span>
              <span className="font-medium text-green-600 flex items-center gap-1.5">
                <div className="w-2 h-2 bg-green-500 rounded-full" />
                {health?.status === 'ok' ? 'Healthy' : 'Unknown'}
              </span>
            </div>
            <div className="border-t border-slate-100" />
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <Clock className="w-4 h-4" />
                Last Health Check
              </span>
              <span className="font-medium text-slate-900">
                {health?.timestamp ? new Date(health.timestamp).toLocaleTimeString() : 'N/A'}
              </span>
            </div>
            <div className="border-t border-slate-100" />
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <Users className="w-4 h-4" />
                Registered Agents
              </span>
              <span className="font-medium text-slate-900">{health?.agents_count || agents.length}</span>
            </div>
            <div className="border-t border-slate-100" />
            <div className="flex items-center justify-between py-2">
              <span className="text-slate-600 flex items-center gap-2">
                <AlertCircle className="w-4 h-4" />
                Expired Agents
              </span>
              <span className={`font-medium ${expiredAgents > 0 ? 'text-amber-600' : 'text-slate-900'}`}>{expiredAgents}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Agents */}
      <div className="bg-white rounded-xl border border-slate-200 p-6">
        <div className="flex items-center justify-between mb-5">
          <h3 className="font-semibold text-slate-900">Recent Agents</h3>
          <div className="flex items-center gap-3">
            <Link to="/agents" className="text-sm text-blue-600 hover:text-blue-800 flex items-center gap-1">
              View all <ArrowRight className="w-3 h-3" />
            </Link>
            <Link 
              to="/agents/create" 
              className="flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 transition-colors"
            >
              <Plus className="w-3.5 h-3.5" /> New Agent
            </Link>
          </div>
        </div>

        {recentAgents.length === 0 ? (
          <div className="text-center py-12">
            <Users className="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <p className="text-slate-500 mb-4">No agents created yet</p>
            <Link 
              to="/agents/create" 
              className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              <Plus className="w-4 h-4" /> Create Your First Agent
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-sm text-slate-500 border-b border-slate-100">
                  <th className="pb-3 font-medium">Name</th>
                  <th className="pb-3 font-medium">Status</th>
                  <th className="pb-3 font-medium">Tokens</th>
                  <th className="pb-3 font-medium">Last Active</th>
                  <th className="pb-3 font-medium">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {recentAgents.map(agent => (
                  <tr key={agent.id} className="hover:bg-slate-50">
                    <td className="py-3">
                      <Link to={`/agents/${agent.id}`} className="text-blue-600 hover:text-blue-800 font-medium">
                        {agent.name}
                      </Link>
                    </td>
                    <td className="py-3">
                      {agent.is_active ? (
                        <span className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full bg-green-100 text-green-700">
                          <CheckCircle2 className="w-3 h-3" /> Active
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full bg-red-100 text-red-700">
                          <XCircle className="w-3 h-3" /> Inactive
                        </span>
                      )}
                    </td>
                    <td className="py-3 text-slate-600">{agent.token_count || 0}</td>
                    <td className="py-3 text-slate-500 text-sm">
                      {agent.last_activity_at 
                        ? new Date(agent.last_activity_at).toLocaleDateString() 
                        : '—'}
                    </td>
                    <td className="py-3 text-slate-500 text-sm">
                      {new Date(agent.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
