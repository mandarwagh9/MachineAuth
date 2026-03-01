import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AgentList } from './pages/agent-list'
import { AgentDetail } from './pages/agent-detail'
import { CreateAgent } from './pages/create-agent'
import { TokenGenerator } from './pages/token-generator'
import { WebhookList } from './pages/webhook-list'
import { WebhookForm } from './pages/webhook-form'
import { WebhookDetail } from './pages/webhook-detail'
import { Layout } from './components/layout'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/agents" replace />} />
          <Route path="agents" element={<AgentList />} />
          <Route path="agents/new" element={<CreateAgent />} />
          <Route path="agents/:id" element={<AgentDetail />} />
          <Route path="tokens" element={<TokenGenerator />} />
          <Route path="webhooks" element={<WebhookList />} />
          <Route path="webhooks/new" element={<WebhookForm />} />
          <Route path="webhooks/:id" element={<WebhookDetail />} />
          <Route path="webhooks/:id/edit" element={<WebhookForm />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
