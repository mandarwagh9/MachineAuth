import { useEffect, useState, useCallback } from 'react'
import { 
  Activity, 
  TrendingUp, 
  TrendingDown,
  RefreshCw,
  Clock,
  Shield,
  Key,
  Users,
  AlertCircle,
  Zap,
  XCircle,
  CheckCircle2,
  BarChart3
} from 'lucide-react'
import { MetricsService, HealthService } from '@/services/api'
import type { Metrics, HealthCheck } from '@/types'

interface MetricsHistory {
  timestamp: Date
  metrics: Metrics
}

const MetricCard = ({ 
  title, 
  value, 
  prevValue,
  icon: Icon, 
  color,
  suffix
}: { 
  title: string
  value: number
  prevValue?: number
  icon: React.ElementType
  color: string
  suffix?: string
}) => {
  const change = prevValue !== undefined && prevValue > 0 
    ? Math.round(((value - prevValue) / prevValue) * 100) 
    : undefined

  return (
    <div className="bg-white rounded-xl border border-slate-200 p-6">
      <div className="flex items-center justify-between mb-4">
        <span className="text-sm font-medium text-slate-500">{title}</span>
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${color}`}>
          <Icon className="w-5 h-5 text-white" />
        </div>
      </div>
      <div className="flex items-end justify-between">
        <div>
          <span className="text-3xl font-bold text-slate-900">{value.toLocaleString()}</span>
          {suffix && <span className="text-sm text-slate-400 ml-1">{suffix}</span>}
        </div>
        {change !== undefined && change !== 0 && (
          <span className={`flex items-center text-sm font-medium ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
            {change >= 0 ? <TrendingUp className="w-4 h-4 mr-1" /> : <TrendingDown className="w-4 h-4 mr-1" />}
            {change > 0 ? '+' : ''}{change}%
          </span>
        )}
      </div>
    </div>
  )
}

const MiniBar = ({ label, value, max, color }: { label: string; value: number; max: number; color: string }) => (
  <div>
    <div className="flex items-center justify-between mb-1.5">
      <span className="text-sm text-slate-600">{label}</span>
      <span className="text-sm font-semibold text-slate-900">{value.toLocaleString()}</span>
    </div>
    <div className="w-full bg-slate-100 rounded-full h-2.5">
      <div 
        className={`${color} h-2.5 rounded-full transition-all duration-500`}
        style={{ width: `${max > 0 ? Math.min((value / max) * 100, 100) : 0}%` }}
      />
    </div>
  </div>
)

export function MetricsPage() {
  const [metrics, setMetrics] = useState<Metrics | null>(null)
  const [prevMetrics, setPrevMetrics] = useState<Metrics | null>(null)
  const [health, setHealth] = useState<HealthCheck | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date())
  const [history, setHistory] = useState<MetricsHistory[]>([])
  const [autoRefresh, setAutoRefresh] = useState(true)

  const fetchMetrics = useCallback(async () => {
    try {
      setError(null)
      const [data, healthData] = await Promise.all([
        MetricsService.get().catch(() => null),
        HealthService.ready().catch(() => null)
      ])
      
      if (!data) {
        setError('Unable to connect to MachineAuth backend')
        return
      }

      setPrevMetrics(metrics)
      setMetrics(data)
      setHealth(healthData)
      setLastUpdated(new Date())
      
      setHistory(prev => {
        const next = [...prev, { timestamp: new Date(), metrics: data }]
        return next.slice(-30) // keep last 30 readings
      })
    } catch (err) {
      setError('Failed to fetch metrics')
      console.error('Metrics fetch error:', err)
    } finally {
      setLoading(false)
    }
  }, [metrics])

  useEffect(() => {
    fetchMetrics()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(fetchMetrics, 10000)
    return () => clearInterval(interval)
  }, [autoRefresh, fetchMetrics])

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {[1,2,3,4].map(i => (
              <div key={i} className="h-32 bg-slate-200 rounded-xl" />
            ))}
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="h-72 bg-slate-200 rounded-xl" />
            <div className="h-72 bg-slate-200 rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  const totalTokenOps = (metrics?.tokens_issued || 0) + (metrics?.tokens_refreshed || 0) + (metrics?.tokens_revoked || 0)

  return (
    <div className="p-8 max-w-7xl">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Metrics</h1>
          <p className="text-slate-500 mt-1">Real-time server statistics</p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg border transition-colors ${
              autoRefresh 
                ? 'bg-green-50 border-green-200 text-green-700' 
                : 'bg-slate-50 border-slate-200 text-slate-600'
            }`}
          >
            <div className={`w-2 h-2 rounded-full ${autoRefresh ? 'bg-green-500 animate-pulse' : 'bg-slate-400'}`} />
            {autoRefresh ? 'Live' : 'Paused'}
          </button>
          <span className="text-xs text-slate-400 flex items-center gap-1">
            <Clock className="w-3 h-3" />
            {lastUpdated.toLocaleTimeString()}
          </span>
          <button
            onClick={fetchMetrics}
            className="p-2 bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors"
            title="Refresh now"
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
            <p className="text-red-600 text-sm mt-0.5">The /api/stats endpoint may be unreachable</p>
          </div>
          <button onClick={fetchMetrics} className="ml-auto text-red-600 hover:text-red-800 text-sm font-medium">
            Retry
          </button>
        </div>
      )}

      {/* Main Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <MetricCard
          title="Tokens Issued"
          value={metrics?.tokens_issued || 0}
          prevValue={prevMetrics?.tokens_issued}
          icon={Zap}
          color="bg-blue-500"
        />
        <MetricCard
          title="Tokens Refreshed"
          value={metrics?.tokens_refreshed || 0}
          prevValue={prevMetrics?.tokens_refreshed}
          icon={RefreshCw}
          color="bg-purple-500"
        />
        <MetricCard
          title="Active Tokens"
          value={metrics?.active_tokens || 0}
          prevValue={prevMetrics?.active_tokens}
          icon={Key}
          color="bg-emerald-500"
        />
        <MetricCard
          title="Tokens Revoked"
          value={metrics?.tokens_revoked || 0}
          prevValue={prevMetrics?.tokens_revoked}
          icon={Shield}
          color="bg-red-500"
        />
      </div>

      {/* Detailed Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Token Breakdown */}
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <div className="flex items-center gap-2 mb-6">
            <BarChart3 className="w-5 h-5 text-slate-400" />
            <h3 className="font-semibold text-slate-900">Token Operations Breakdown</h3>
          </div>
          <div className="space-y-5">
            <MiniBar 
              label="Issued" 
              value={metrics?.tokens_issued || 0} 
              max={totalTokenOps || 1} 
              color="bg-blue-500" 
            />
            <MiniBar 
              label="Refreshed" 
              value={metrics?.tokens_refreshed || 0} 
              max={totalTokenOps || 1} 
              color="bg-purple-500" 
            />
            <MiniBar 
              label="Revoked" 
              value={metrics?.tokens_revoked || 0} 
              max={totalTokenOps || 1} 
              color="bg-red-500" 
            />
            <MiniBar 
              label="Active" 
              value={metrics?.active_tokens || 0} 
              max={metrics?.tokens_issued || 1} 
              color="bg-emerald-500" 
            />
          </div>
          <div className="mt-6 pt-4 border-t border-slate-100">
            <div className="flex items-center justify-between text-sm">
              <span className="text-slate-500">Total Token Operations</span>
              <span className="font-bold text-slate-900">{totalTokenOps.toLocaleString()}</span>
            </div>
          </div>
        </div>

        {/* System Overview */}
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <div className="flex items-center gap-2 mb-6">
            <Activity className="w-5 h-5 text-slate-400" />
            <h3 className="font-semibold text-slate-900">System Overview</h3>
          </div>

          <div className="grid grid-cols-2 gap-4 mb-6">
            <div className="bg-slate-50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-slate-500 text-sm mb-2">
                <Users className="w-4 h-4" />
                Total Agents
              </div>
              <p className="text-2xl font-bold text-slate-900">{metrics?.total_agents || 0}</p>
            </div>
            <div className="bg-slate-50 rounded-lg p-4">
              <div className="flex items-center gap-2 text-slate-500 text-sm mb-2">
                <Key className="w-4 h-4" />
                Active Tokens
              </div>
              <p className="text-2xl font-bold text-emerald-600">{metrics?.active_tokens || 0}</p>
            </div>
          </div>

          {/* Health status */}
          <div className="space-y-3">
            <div className="flex items-center justify-between py-2 border-t border-slate-100">
              <span className="text-sm text-slate-600">Backend</span>
              {health?.status === 'ok' ? (
                <span className="flex items-center gap-1.5 text-sm font-medium text-green-600">
                  <CheckCircle2 className="w-4 h-4" /> Healthy
                </span>
              ) : (
                <span className="flex items-center gap-1.5 text-sm font-medium text-red-600">
                  <XCircle className="w-4 h-4" /> Unreachable
                </span>
              )}
            </div>
            <div className="flex items-center justify-between py-2 border-t border-slate-100">
              <span className="text-sm text-slate-600">Auto-refresh</span>
              <span className="text-sm font-medium text-slate-900">{autoRefresh ? 'Every 10s' : 'Paused'}</span>
            </div>
            <div className="flex items-center justify-between py-2 border-t border-slate-100">
              <span className="text-sm text-slate-600">Readings collected</span>
              <span className="text-sm font-medium text-slate-900">{history.length}</span>
            </div>
          </div>
        </div>
      </div>

      {/* History Timeline */}
      {history.length > 1 && (
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <div className="flex items-center gap-2 mb-4">
            <Clock className="w-5 h-5 text-slate-400" />
            <h3 className="font-semibold text-slate-900">Recent History</h3>
            <span className="text-xs text-slate-400 ml-auto">Last {history.length} readings</span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-slate-500 border-b border-slate-100">
                  <th className="pb-2 font-medium">Time</th>
                  <th className="pb-2 font-medium text-right">Issued</th>
                  <th className="pb-2 font-medium text-right">Refreshed</th>
                  <th className="pb-2 font-medium text-right">Revoked</th>
                  <th className="pb-2 font-medium text-right">Active</th>
                  <th className="pb-2 font-medium text-right">Agents</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50">
                {[...history].reverse().slice(0, 10).map((h, i) => (
                  <tr key={i} className={i === 0 ? 'bg-blue-50/50' : ''}>
                    <td className="py-2 text-slate-600">{h.timestamp.toLocaleTimeString()}</td>
                    <td className="py-2 text-right font-medium">{h.metrics.tokens_issued}</td>
                    <td className="py-2 text-right font-medium">{h.metrics.tokens_refreshed}</td>
                    <td className="py-2 text-right font-medium">{h.metrics.tokens_revoked}</td>
                    <td className="py-2 text-right font-medium text-emerald-600">{h.metrics.active_tokens}</td>
                    <td className="py-2 text-right font-medium">{h.metrics.total_agents}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
