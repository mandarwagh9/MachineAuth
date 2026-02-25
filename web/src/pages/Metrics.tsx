import { useEffect, useState } from 'react'
import { 
  Activity, 
  TrendingUp, 
  TrendingDown,
  RefreshCw,
  Clock,
  Shield,
  Key,
  Users
} from 'lucide-react'
import { MetricsService } from '@/services/api'
import type { Metrics } from '@/types'

const MetricCard = ({ 
  title, 
  value, 
  change,
  icon: Icon, 
  color 
}: { 
  title: string
  value: number
  change?: number
  icon: React.ElementType
  color: string 
}) => (
  <div className="bg-white rounded-xl border border-slate-200 p-6">
    <div className="flex items-center justify-between mb-4">
      <span className="text-sm font-medium text-slate-500">{title}</span>
      <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${color}`}>
        <Icon className="w-5 h-5 text-white" />
      </div>
    </div>
    <div className="flex items-end justify-between">
      <span className="text-3xl font-bold text-slate-900">{value.toLocaleString()}</span>
      {change !== undefined && (
        <span className={`flex items-center text-sm font-medium ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
          {change >= 0 ? <TrendingUp className="w-4 h-4 mr-1" /> : <TrendingDown className="w-4 h-4 mr-1" />}
          {Math.abs(change)}%
        </span>
      )}
    </div>
  </div>
)

export function MetricsPage() {
  const [metrics, setMetrics] = useState<Metrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date())

  const fetchMetrics = async () => {
    try {
      const data = await MetricsService.get()
      setMetrics(data)
      setLastUpdated(new Date())
    } catch (error) {
      console.error('Failed to fetch metrics:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchMetrics()
    const interval = setInterval(fetchMetrics, 10000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
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
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Metrics</h1>
          <p className="text-slate-500 mt-1">Real-time server statistics</p>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-slate-500 flex items-center gap-1">
            <Clock className="w-4 h-4" />
            Updated {lastUpdated.toLocaleTimeString()}
          </span>
          <button
            onClick={fetchMetrics}
            className="p-2 bg-slate-100 hover:bg-slate-200 rounded-lg"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Main Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <MetricCard
          title="Total Requests"
          value={metrics?.requests || 0}
          icon={Activity}
          color="bg-blue-500"
        />
        <MetricCard
          title="Tokens Issued"
          value={metrics?.tokens_issued || 0}
          icon={Key}
          color="bg-green-500"
        />
        <MetricCard
          title="Tokens Refreshed"
          value={metrics?.tokens_refreshed || 0}
          icon={RefreshCw}
          color="bg-purple-500"
        />
        <MetricCard
          title="Tokens Revoked"
          value={metrics?.tokens_revoked || 0}
          icon={Shield}
          color="bg-red-500"
        />
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Token Statistics */}
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <h3 className="font-semibold text-slate-900 mb-4">Token Statistics</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-slate-600">Active Tokens</span>
              <span className="font-semibold text-slate-900">{metrics?.active_tokens || 0}</span>
            </div>
            <div className="w-full bg-slate-100 rounded-full h-2">
              <div 
                className="bg-blue-500 h-2 rounded-full" 
                style={{ width: `${Math.min(((metrics?.active_tokens || 0) / 100) * 100, 100)}%` }}
              />
            </div>
            
            <div className="pt-4 border-t border-slate-100">
              <div className="flex items-center justify-between mb-2">
                <span className="text-slate-600">Total Agents</span>
                <span className="font-semibold text-slate-900">{metrics?.total_agents || 0}</span>
              </div>
              <div className="w-full bg-slate-100 rounded-full h-2">
                <div 
                  className="bg-green-500 h-2 rounded-full" 
                  style={{ width: `${Math.min(((metrics?.total_agents || 0) / 50) * 100, 100)}%` }}
                />
              </div>
            </div>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <h3 className="font-semibold text-slate-900 mb-4">Quick Stats</h3>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-slate-50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-slate-500 text-sm mb-1">
                <Users className="w-4 h-4" />
                Agents
              </div>
              <p className="text-2xl font-bold text-slate-900">{metrics?.total_agents || 0}</p>
            </div>
            <div className="bg-slate-50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-slate-500 text-sm mb-1">
                <Key className="w-4 h-4" />
                Active Tokens
              </div>
              <p className="text-2xl font-bold text-slate-900">{metrics?.active_tokens || 0}</p>
            </div>
            <div className="bg-slate-50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-slate-500 text-sm mb-1">
                <Activity className="w-4 h-4" />
                Requests
              </div>
              <p className="text-2xl font-bold text-slate-900">{(metrics?.requests || 0).toLocaleString()}</p>
            </div>
            <div className="bg-slate-50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-slate-500 text-sm mb-1">
                <Shield className="w-4 h-4" />
                Revoked
              </div>
              <p className="text-2xl font-bold text-slate-900">{metrics?.tokens_revoked || 0}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
