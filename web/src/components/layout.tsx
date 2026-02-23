import { NavLink, Outlet } from 'react-router-dom'
import './layout.css'

export function Layout() {
  return (
    <div className="layout">
      <header className="header">
        <div className="header-content">
          <h1 className="logo">AgentAuth</h1>
          <nav className="nav">
            <NavLink
              to="/agents"
              className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}
            >
              Agents
            </NavLink>
            <NavLink
              to="/tokens"
              className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}
            >
              Tokens
            </NavLink>
          </nav>
        </div>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
