import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { useState, useEffect, createContext, useContext } from 'react'
import { Toaster } from 'sonner'
import { Layout } from './components/layout'
import { LoginPage } from './pages/Login'
import { Dashboard } from './pages/Dashboard'
import { AgentsPage } from './pages/Agents'
import { CreateAgentPage } from './pages/CreateAgent'
import { MetricsPage } from './pages/Metrics'

interface AuthContextType {
  isAuthenticated: boolean
  login: (username: string, password: string) => boolean
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

const AUTH_USER = 'admin'
const AUTH_PASS = 'mandar123' // Change this to your desired password

function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    const stored = localStorage.getItem('machineauth_auth')
    if (stored === 'true') {
      setIsAuthenticated(true)
    }
  }, [])

  const login = (username: string, password: string): boolean => {
    if (username === AUTH_USER && password === AUTH_PASS) {
      setIsAuthenticated(true)
      localStorage.setItem('machineauth_auth', 'true')
      return true
    }
    return false
  }

  const logout = () => {
    setIsAuthenticated(false)
    localStorage.removeItem('machineauth_auth')
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
    await new Promise(resolve => setTimeout(resolve, 500)) // Fake delay for UX
    if (auth?.login(username, password)) {
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
        <Route path="agents" element={<AgentsPage />} />
        <Route path="agents/new" element={<CreateAgentPage />} />
        <Route path="agents/:id" element={<AgentsPage />} />
        <Route path="metrics" element={<MetricsPage />} />
        <Route path="tokens" element={<AgentsPage />} />
        <Route path="settings" element={<Dashboard />} />
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
