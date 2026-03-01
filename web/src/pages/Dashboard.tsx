import { useEffect, useState } from 'react'
import { 
  Users, 
  Key, 
  Activity, 
  TrendingUp,
  Shield,
  Clock
} from 'lucide-react'
import { MetricsService, HealthService } from '@/services/api'
import type { Metrics, HealthCheck } from '@/types'

const StatCard = ({ 
  title, 
  value, 
  icon: Icon, 
  trend,
  color 
}: { 
  title: string
  value: string | number
  icon: React.ElementType
  trend?: string
  color: string 
}) => (
  <div className="bg-white rounded-xl border border-slate-200 p-6">
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm font-medium text-slate-500">{title}</p>
        <p className="mt-2 text-3xl font-bold text-slate-900">{value}</p>
        {trend && (
          <p className="mt-1 text-sm text-green-600 flex items-center gap-1">
            <TrendingUp className="w-3 h-3" />
            {trend}
          </p>
        )}
      </div>
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${color}`}>
        <Icon className="w-6 h-6 text-white" />
      </div>
    </div>
  </div>
)

export function Dashboard() {
  const [metrics, setMetrics] = useState<Metrics | null>(null)
  const [health, setHealth] = useState<HealthCheck | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [metricsData, healthData] = await Promise.all([
          MetricsService.get(),
          HealthService.ready()
        ])
        setMetrics(metricsData)
        setHealth(healthData)
      } catch (error) {
        console.error('Failed to fetch metrics:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
    const interval = setInterval(fetchData, 30000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            {[1,2,3,4].map(i => (
              <div key={i} className="h-32 bg-slate-200 rounded-xl" />
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>
        <p className="text-slate-500 mt-1">Overview of your MachineAuth instance</p>
      </div>

      {/* Status Banner */}
      <div className="mb-8 bg-green-50 border border-green-200 rounded-xl p-4 flex items-center gap-3">
        <div className="w-3 h-3 bg-green-500 rounded-full" />
        <span className="text-green-700 font-medium">System Online</span>
        <span className="text-green-600">•</span>
        <span className="text-green-600">{health?.agents_count || 0} agents registered</span>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatCard
          title="Total Agents"
          value={metrics?.total_agents || 0}
          icon={Users}
          color="bg-blue-500"
        />
        <StatCard
          title="Active Tokens"
          value={metrics?.active_tokens || 0}
          icon={Key}
          color="bg-green-500"
        />
        <StatCard
          title="Tokens Issued"
          value={metrics?.tokens_issued || 0}
          icon={Shield}
          color="bg-purple-500"
        />
        <StatCard
          title="Total Requests"
          value={metrics?.requests || 0}
          icon={Activity}
          color="bg-orange-500"
        />
      </div>

      {/* Additional Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <h3 className="font-semibold text-slate-900 mb-4">Token Activity</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Tokens Issued</span>
              <span className="font-medium text-slate-900">{metrics?.tokens_issued || 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Tokens Refreshed</span>
              <span className="font-medium text-slate-900">{metrics?.tokens_refreshed || 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Tokens Revoked</span>
              <span className="font-medium text-slate-900">{metrics?.tokens_revoked || 0}</span>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <h3 className="font-semibold text-slate-900 mb-4">System Health</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-slate-600 flex items-center gap-2">
                <Clock className="w-4 h-4" />
                Last Check
              </span>
              <span className="font-medium text-slate-900">
                {health?.timestamp ? new Date(health.timestamp).toLocaleTimeString() : 'N/A'}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-600 flex items-center gap-2">
                <Activity className="w-4 h-4" />
                Status
              </span>
              <span className="font-medium text-green-600 flex items-center gap-1">
                <div className="w-2 h-2 bg-green-500 rounded-full" />
                Healthy
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
