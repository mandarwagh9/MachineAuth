import { Link, useLocation } from 'react-router-dom'
import { 
  LayoutDashboard, 
  Building2,
  Users, 
  Settings, 
  Activity,
  Key,
  Shield,
  User
} from 'lucide-react'

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Organizations', href: '/organizations', icon: Building2 },
  { name: 'Agents', href: '/agents', icon: Users },
  { name: 'Tokens', href: '/tokens', icon: Key },
  { name: 'Metrics', href: '/metrics', icon: Activity },
  { name: 'Settings', href: '/settings', icon: Settings },
  { name: 'My Account', href: '/my-account', icon: User },
]

export function Sidebar() {
  const location = useLocation()

  return (
    <div className="flex flex-col h-full bg-slate-900 border-r border-slate-800 w-64">
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
