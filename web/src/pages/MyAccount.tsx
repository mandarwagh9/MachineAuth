import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { 
  User, 
  Key, 
  Trash2, 
  RefreshCw,
  Activity,
  Clock,
  Copy,
  CheckCircle,
  XCircle,
  AlertTriangle
} from 'lucide-react'
import { AgentSelfService } from '@/services/api'
import type { AgentUsage } from '@/types'
import { toast } from 'sonner'

export function MyAccountPage() {
  const [usage, setUsage] = useState<AgentUsage | null>(null)
  const [loading, setLoading] = useState(true)
  const [token, setToken] = useState<string | null>(null)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const storedToken = localStorage.getItem('agent_token')
    if (!storedToken) {
      toast.error('Please login as an agent to access this page')
      navigate('/login')
      return
    }
    setToken(storedToken)
    fetchUsage(storedToken)
  }, [navigate])

  const fetchUsage = async (authToken: string) => {
    try {
      const data = await AgentSelfService.getUsage(authToken)
      setUsage(data)
    } catch (error) {
      console.error('Failed to fetch usage:', error)
      toast.error('Failed to load account data')
    } finally {
      setLoading(false)
    }
  }

  const handleRotateCredentials = async () => {
    if (!token) return
    
    try {
      const result = await AgentSelfService.rotateCredentials(token)
      toast.success('Credentials rotated successfully')
      toast.info(`New client_secret: ${result.client_secret}`, { duration: 10000 })
      fetchUsage(token)
    } catch (error) {
      console.error('Failed to rotate credentials:', error)
      toast.error('Failed to rotate credentials')
    }
  }

  const handleDeactivate = async () => {
    if (!token) return
    
    try {
      await AgentSelfService.deactivate(token)
      toast.success('Account deactivated')
      fetchUsage(token)
    } catch (error) {
      console.error('Failed to deactivate:', error)
      toast.error('Failed to deactivate account')
    }
  }

  const handleReactivate = async () => {
    if (!token) return
    
    try {
      await AgentSelfService.reactivate(token)
      toast.success('Account reactivated')
      fetchUsage(token)
    } catch (error) {
      console.error('Failed to reactivate:', error)
      toast.error('Failed to reactivate account')
    }
  }

  const handleDelete = async () => {
    if (!token) return
    
    setDeleting(true)
    try {
      await AgentSelfService.delete(token)
      localStorage.removeItem('agent_token')
      toast.success('Account deleted')
      navigate('/login')
    } catch (error) {
      console.error('Failed to delete:', error)
      toast.error('Failed to delete account')
    } finally {
      setDeleting(false)
      setShowDeleteConfirm(false)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard')
  }

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never'
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-slate-200 rounded w-48" />
          <div className="h-64 bg-slate-200 rounded-xl" />
        </div>
      </div>
    )
  }

  if (!usage) {
    return (
      <div className="p-8">
        <p className="text-slate-500">Failed to load account data</p>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-slate-900">My Account</h1>
        <p className="text-slate-500 mt-1">Manage your agent account and view usage</p>
      </div>

      {/* Status Banner */}
      {!usage.agent.is_active && (
        <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg flex items-center gap-3">
          <AlertTriangle className="w-5 h-5 text-amber-600" />
          <p className="text-amber-800">Your account is currently deactivated. You cannot obtain new tokens.</p>
        </div>
      )}

      {/* Agent Info Card */}
      <div className="bg-white rounded-xl border border-slate-200 overflow-hidden mb-6">
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
              <User className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <h2 className="font-semibold text-slate-900">{usage.agent.name}</h2>
              <p className="text-sm text-slate-500">Agent Account</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {usage.agent.is_active ? (
              <span className="inline-flex items-center gap-1 px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm font-medium">
                <CheckCircle className="w-4 h-4" />
                Active
              </span>
            ) : (
              <span className="inline-flex items-center gap-1 px-3 py-1 bg-red-100 text-red-700 rounded-full text-sm font-medium">
                <XCircle className="w-4 h-4" />
                Inactive
              </span>
            )}
          </div>
        </div>
        
        <div className="p-6 space-y-4">
          <div>
            <label className="text-sm font-medium text-slate-500">Client ID</label>
            <div className="flex items-center gap-2 mt-1">
              <code className="flex-1 text-sm text-slate-600 font-mono bg-slate-100 px-3 py-2 rounded-lg">
                {usage.agent.client_id}
              </code>
              <button
                onClick={() => copyToClipboard(usage.agent.client_id)}
                className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg"
              >
                <Copy className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div>
            <label className="text-sm font-medium text-slate-500">Scopes</label>
            <div className="flex flex-wrap gap-2 mt-2">
              {usage.agent.scopes.map((scope) => (
                <span
                  key={scope}
                  className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-slate-100 text-slate-600"
                >
                  {scope}
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Usage Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Key className="w-5 h-5 text-blue-600" />
            <span className="text-sm font-medium text-slate-500">Tokens Issued</span>
          </div>
          <p className="text-3xl font-bold text-slate-900">{usage.token_count}</p>
        </div>

        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <div className="flex items-center gap-3 mb-2">
            <RefreshCw className="w-5 h-5 text-purple-600" />
            <span className="text-sm font-medium text-slate-500">Token Refreshes</span>
          </div>
          <p className="text-3xl font-bold text-slate-900">{usage.refresh_count}</p>
        </div>

        <div className="bg-white rounded-xl border border-slate-200 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Activity className="w-5 h-5 text-green-600" />
            <span className="text-sm font-medium text-slate-500">Last Activity</span>
          </div>
          <p className="text-lg font-semibold text-slate-900">{formatDate(usage.last_activity_at)}</p>
        </div>
      </div>

      {/* Activity History */}
      <div className="bg-white rounded-xl border border-slate-200 p-6 mb-6">
        <h3 className="font-semibold text-slate-900 mb-4">Activity</h3>
        <div className="space-y-3">
          <div className="flex items-center justify-between py-2 border-b border-slate-100">
            <div className="flex items-center gap-3">
              <Clock className="w-4 h-4 text-slate-400" />
              <span className="text-sm text-slate-600">Account Created</span>
            </div>
            <span className="text-sm text-slate-900">{formatDate(usage.agent.created_at)}</span>
          </div>
          <div className="flex items-center justify-between py-2 border-b border-slate-100">
            <div className="flex items-center gap-3">
              <Clock className="w-4 h-4 text-slate-400" />
              <span className="text-sm text-slate-600">Last Token Issued</span>
            </div>
            <span className="text-sm text-slate-900">{formatDate(usage.last_token_issued_at)}</span>
          </div>
          <div className="flex items-center justify-between py-2">
            <div className="flex items-center gap-3">
              <Clock className="w-4 h-4 text-slate-400" />
              <span className="text-sm text-slate-600">Last Activity</span>
            </div>
            <span className="text-sm text-slate-900">{formatDate(usage.last_activity_at)}</span>
          </div>
        </div>
      </div>

      {/* Rotation History */}
      {usage.rotation_history.length > 0 && (
        <div className="bg-white rounded-xl border border-slate-200 p-6 mb-6">
          <h3 className="font-semibold text-slate-900 mb-4">Credential Rotations</h3>
          <div className="space-y-2">
            {usage.rotation_history.map((rotation, index) => (
              <div key={index} className="flex items-center justify-between py-2 border-b border-slate-100 last:border-0">
                <div className="flex items-center gap-3">
                  <Key className="w-4 h-4 text-slate-400" />
                  <span className="text-sm text-slate-600">Rotation #{usage.rotation_history.length - index}</span>
                </div>
                <span className="text-sm text-slate-900">{formatDate(rotation.rotated_at)}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="bg-white rounded-xl border border-slate-200 p-6">
        <h3 className="font-semibold text-slate-900 mb-4">Account Actions</h3>
        <div className="space-y-3">
          <button
            onClick={handleRotateCredentials}
            disabled={!usage.agent.is_active}
            className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed text-white font-medium rounded-lg transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
            Rotate Credentials
          </button>

          {usage.agent.is_active ? (
            <button
              onClick={handleDeactivate}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-amber-100 hover:bg-amber-200 text-amber-700 font-medium rounded-lg transition-colors"
            >
              <XCircle className="w-4 h-4" />
              Deactivate Account
            </button>
          ) : (
            <button
              onClick={handleReactivate}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-green-100 hover:bg-green-200 text-green-700 font-medium rounded-lg transition-colors"
            >
              <CheckCircle className="w-4 h-4" />
              Reactivate Account
            </button>
          )}

          <button
            onClick={() => setShowDeleteConfirm(true)}
            className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-red-50 hover:bg-red-100 text-red-600 font-medium rounded-lg transition-colors"
          >
            <Trash2 className="w-4 h-4" />
            Delete Account
          </button>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-slate-900 mb-2">Delete Account?</h3>
            <p className="text-slate-600 mb-6">
              This action cannot be undone. Your agent account will be permanently deleted and you will no longer be able to obtain tokens.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="flex-1 px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                disabled={deleting}
                className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-400 text-white font-medium rounded-lg transition-colors"
              >
                {deleting ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
