import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { useState, useEffect, createContext, useContext } from 'react'
import { Toaster } from 'sonner'
import { WebhookList } from './pages/webhook-list'
import { WebhookForm } from './pages/webhook-form'
import { WebhookDetail } from './pages/webhook-detail'
import { Layout } from './components/layout'
import { LoginPage } from './pages/Login'
import { Dashboard } from './pages/Dashboard'
import { AgentsPage } from './pages/Agents'
import { CreateAgentPage } from './pages/CreateAgent'
import { MetricsPage } from './pages/Metrics'
import { MyAccountPage } from './pages/MyAccount'
import { TokensPage } from './pages/Tokens'
import { AgentDetailPage } from './pages/AgentDetail'
import { OrganizationsPage } from './pages/Organizations'
import { CreateOrganizationPage } from './pages/CreateOrganization'
import { OrganizationDetailPage } from './pages/OrganizationDetail'

interface AuthContextType {
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<boolean>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    const stored = localStorage.getItem('machineauth_auth')
    if (stored === 'true') {
      setIsAuthenticated(true)
    }
  }, [])

  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })
      const data = await res.json()
      if (data.success && data.access_token) {
        setIsAuthenticated(true)
        localStorage.setItem('machineauth_auth', 'true')
        localStorage.setItem('machineauth_token', data.access_token)
        return true
      }
      return false
    } catch {
      return false
    }
  }

  const logout = () => {
    setIsAuthenticated(false)
    localStorage.removeItem('machineauth_auth')
    localStorage.removeItem('machineauth_token')
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const auth = useContext(AuthContext)
  if (!auth?.isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function Login() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const auth = useContext(AuthContext)

  const handleLogin = async (username: string, password: string) => {
    setLoading(true)
    const success = await auth?.login(username, password)
    if (success) {
      navigate('/')
    } else {
      alert('Invalid credentials')
    }
    setLoading(false)
  }

  return <LoginPage onLogin={handleLogin} isLoading={loading} />
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<ProtectedRoute><Layout /></ProtectedRoute>}>
        <Route index element={<Dashboard />} />
        <Route path="organizations" element={<OrganizationsPage />} />
        <Route path="organizations/new" element={<CreateOrganizationPage />} />
        <Route path="organizations/:id" element={<OrganizationDetailPage />} />
        <Route path="agents" element={<AgentsPage />} />
        <Route path="agents/new" element={<CreateAgentPage />} />
        <Route path="agents/:id" element={<AgentDetailPage />} />
        <Route path="metrics" element={<MetricsPage />} />
        <Route path="tokens" element={<TokensPage />} />
        <Route path="settings" element={<Dashboard />} />
        <Route path="my-account" element={<MyAccountPage />} />
        <Route path="webhooks" element={<WebhookList />} />
        <Route path="webhooks/new" element={<WebhookForm />} />
        <Route path="webhooks/:id" element={<WebhookDetail />} />
        <Route path="webhooks/:id/edit" element={<WebhookForm />} />
      </Route>
    </Routes>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Toaster position="top-right" />
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
