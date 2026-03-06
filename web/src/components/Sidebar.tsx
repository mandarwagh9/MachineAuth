import { Link, useLocation } from 'react-router-dom'
import { useState } from 'react'
import { 
  LayoutDashboard, 
  Building2,
  Users, 
  Settings, 
  Activity,
  Key,
  Shield,
  User,
  Webhook,
  ChevronDown,
  Plus
} from 'lucide-react'
import { useOrg } from '@/contexts/OrgContext'

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Organizations', href: '/organizations', icon: Building2 },
  { name: 'Agents', href: '/agents', icon: Users },
  { name: 'Tokens', href: '/tokens', icon: Key },
  { name: 'Webhooks', href: '/webhooks', icon: Webhook },
  { name: 'Metrics', href: '/metrics', icon: Activity },
  { name: 'Settings', href: '/settings', icon: Settings },
  { name: 'My Account', href: '/my-account', icon: User },
]

export function Sidebar() {
  const location = useLocation()
  const { currentOrg, organizations, switchOrg, loading } = useOrg()
  const [dropdownOpen, setDropdownOpen] = useState(false)

  return (
    <div className="flex flex-col h-full bg-slate-900 border-r border-slate-800 w-64">
      {/* Organization Selector */}
      <div className="px-3 py-3 border-b border-slate-800">
        <div className="relative">
          <button
            onClick={() => setDropdownOpen(!dropdownOpen)}
            className="w-full flex items-center gap-2 px-3 py-2 bg-slate-800 hover:bg-slate-700 rounded-lg text-left"
          >
            <Building2 className="w-4 h-4 text-blue-500" />
            <div className="flex-1 min-w-0">
              {loading ? (
                <span className="text-xs text-slate-400">Loading...</span>
              ) : currentOrg ? (
                <>
                  <p className="text-sm font-medium text-white truncate">{currentOrg.name}</p>
                  <p className="text-xs text-slate-400 truncate">{currentOrg.slug}</p>
                </>
              ) : (
                <span className="text-sm text-slate-400">Select Organization</span>
              )}
            </div>
            <ChevronDown className={`w-4 h-4 text-slate-400 transition-transform ${dropdownOpen ? 'rotate-180' : ''}`} />
          </button>
          
          {dropdownOpen && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-slate-800 border border-slate-700 rounded-lg shadow-lg overflow-hidden z-50">
              <div className="max-h-60 overflow-y-auto">
                {organizations.map((org) => (
                  <button
                    key={org.id}
                    onClick={() => {
                      switchOrg(org.id)
                      setDropdownOpen(false)
                    }}
                    className={`w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-slate-700 ${
                      currentOrg?.id === org.id ? 'bg-slate-700' : ''
                    }`}
                  >
                    <Building2 className="w-4 h-4 text-slate-400" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-white truncate">{org.name}</p>
                      <p className="text-xs text-slate-400 truncate">{org.slug}</p>
                    </div>
                  </button>
                ))}
              </div>
              <div className="border-t border-slate-700 px-3 py-2">
                <Link
                  to="/organizations/new"
                  onClick={() => setDropdownOpen(false)}
                  className="flex items-center gap-2 text-sm text-blue-400 hover:text-blue-300"
                >
                  <Plus className="w-4 h-4" />
                  New Organization
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2 px-6 py-4 border-b border-slate-800">
        <Shield className="w-8 h-8 text-blue-500" />
        <div>
          <h1 className="text-lg font-bold text-white">MachineAuth</h1>
          <p className="text-xs text-slate-400">OAuth for AI Agents</p>
        </div>
      </div>

      <nav className="flex-1 px-3 py-4 space-y-1">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href || 
            (item.href !== '/' && location.pathname.startsWith(item.href))
          return (
            <Link
              key={item.name}
              to={item.href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-blue-600 text-white'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white'
              }`}
            >
              <item.icon className="w-5 h-5" />
              {item.name}
            </Link>
          )
        })}
      </nav>

      <div className="px-4 py-4 border-t border-slate-800">
        <div className="flex items-center gap-2 text-xs text-slate-500">
          <div className="w-2 h-2 rounded-full bg-green-500" />
          <span>System Online</span>
        </div>
      </div>
    </div>
  )
}
