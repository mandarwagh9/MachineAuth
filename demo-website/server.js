const express = require('express');
const bodyParser = require('body-parser');

const app = express();
const PORT = process.env.PORT || 3001;

// MachineAuth configuration
const MACHINEAUTH_URL = process.env.MACHINEAUTH_URL || 'https://auth.writesomething.fun';

// In-memory store for demo (in production, use a database)
const users = new Map();

app.use(bodyParser.json());
app.use(express.static('public'));

// Serve the demo page
app.get('/', (req, res) => {
  res.send(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>MachineAuth Demo - Protected App</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .container {
      background: white;
      border-radius: 16px;
      box-shadow: 0 20px 60px rgba(0,0,0,0.3);
      padding: 40px;
      max-width: 480px;
      width: 90%;
    }
    h1 { color: #333; margin-bottom: 8px; }
    .subtitle { color: #666; margin-bottom: 24px; }
    .form-group { margin-bottom: 20px; }
    label { display: block; margin-bottom: 8px; color: #555; font-weight: 500; }
    input {
      width: 100%;
      padding: 12px;
      border: 2px solid #e0e0e0;
      border-radius: 8px;
      font-size: 16px;
      transition: border-color 0.3s;
    }
    input:focus { outline: none; border-color: #667eea; }
    button {
      width: 100%;
      padding: 14px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      border: none;
      border-radius: 8px;
      font-size: 16px;
      font-weight: 600;
      cursor: pointer;
      transition: transform 0.2s;
    }
    button:hover { transform: translateY(-2px); }
    .error { color: #e53e3e; margin-bottom: 16px; padding: 12px; background: #fed7d7; border-radius: 8px; }
    .success { color: #38a169; margin-bottom: 16px; padding: 12px; background: #c6f6d5; border-radius: 8px; }
    .hidden { display: none; }
    .dashboard {
      text-align: center;
    }
    .secret-box {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      padding: 24px;
      border-radius: 12px;
      margin: 24px 0;
    }
    .secret-code {
      font-family: monospace;
      font-size: 28px;
      font-weight: bold;
      letter-spacing: 2px;
    }
    .user-info {
      text-align: left;
      background: #f7fafc;
      padding: 16px;
      border-radius: 8px;
      margin-bottom: 24px;
    }
    .user-info p { margin: 8px 0; }
    .user-info strong { color: #4a5568; }
    .logout-btn {
      background: #e53e3e;
    }
    .demo-note {
      margin-top: 24px;
      padding: 16px;
      background: #ebf8ff;
      border-radius: 8px;
      font-size: 14px;
      color: #2c5282;
    }
  </style>
</head>
<body>
  <div class="container">
    <!-- Login Form -->
    <div id="loginForm">
      <h1>🔐 Protected App</h1>
      <p class="subtitle">Powered by MachineAuth</p>
      
      <div id="message"></div>
      
      <form id="authForm">
        <div class="form-group">
          <label for="clientId">Client ID</label>
          <input type="text" id="clientId" placeholder="Enter your MachineAuth Client ID" required>
        </div>
        
        <div class="form-group">
          <label for="clientSecret">Client Secret</label>
          <input type="password" id="clientSecret" placeholder="Enter your MachineAuth Client Secret" required>
        </div>
        
        <button type="submit">Login</button>
      </form>
      
      <div class="demo-note">
        <strong>Don't have credentials?</strong><br>
        Use the MachineAuth API to create an agent first:<br>
        <code style="background:#e2e8f0;padding:2px 6px;border-radius:4px;">POST /api/agents</code>
      </div>
    </div>
    
    <!-- Dashboard (Protected) -->
    <div id="dashboard" class="dashboard hidden">
      <h1>🎉 Welcome!</h1>
      <p class="subtitle">You have successfully authenticated</p>
      
      <div class="user-info">
        <p><strong>Agent Name:</strong> <span id="agentName">-</span></p>
        <p><strong>Client ID:</strong> <span id="displayClientId">-</span></p>
        <p><strong>Scopes:</strong> <span id="agentScopes">-</span></p>
      </div>
      
      <div class="secret-box">
        <p style="margin-bottom:8px;">🔑 Your Secret Code:</p>
        <p class="secret-code" id="secretCode">Loading...</p>
      </div>
      
      <p style="color:#666;font-size:14px;margin-bottom:16px;">
        This secret can only be accessed by authenticated agents.<br>
        If an AI agent retrieves this exact code, it's NOT hallucinating!
      </p>
      
      <button class="logout-btn" onclick="logout()">Logout</button>
    </div>
  </div>

  <script>
    const MACHINEAUTH_URL = '${MACHINEAUTH_URL}';
    
    // Check if already logged in
    const token = localStorage.getItem('demo_token');
    const userData = localStorage.getItem('demo_user');
    
    if (token && userData) {
      showDashboard(JSON.parse(userData));
    }
    
    document.getElementById('authForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      
      const clientId = document.getElementById('clientId').value;
      const clientSecret = document.getElementById('clientSecret').value;
      
      const messageDiv = document.getElementById('message');
      messageDiv.innerHTML = '';
      
      try {
        // Get token from MachineAuth
        const response = await fetch(MACHINEAUTH_URL + '/oauth/token', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          },
          body: new URLSearchParams({
            grant_type: 'client_credentials',
            client_id: clientId,
            client_secret: clientSecret
          })
        });
        
        const data = await response.json();
        
        if (!response.ok || data.error) {
          throw new Error(data.error_description || 'Authentication failed');
        }
        
        // Get agent info
        const agentResponse = await fetch(MACHINEAUTH_URL + '/api/agents/me', {
          headers: {
            'Authorization': 'Bearer ' + data.access_token
          }
        });
        
        const agentData = await agentResponse.json();
        
        // Get protected content
        const verifyResponse = await fetch(MACHINEAUTH_URL + '/api/verify', {
          headers: {
            'Authorization': 'Bearer ' + data.access_token
          }
        });
        
        const verifyData = await verifyResponse.json();
        
        // Store and show dashboard
        const userInfo = {
          agent: agentData.agent,
          token: data.access_token,
          secret: verifyData.secret_code
        };
        
        localStorage.setItem('demo_token', data.access_token);
        localStorage.setItem('demo_user', JSON.stringify(userInfo));
        
        showDashboard(userInfo);
        
      } catch (error) {
        messageDiv.innerHTML = '<div class="error">' + error.message + '</div>';
      }
    });
    
    function showDashboard(userData) {
      document.getElementById('loginForm').classList.add('hidden');
      document.getElementById('dashboard').classList.remove('hidden');
      
      document.getElementById('agentName').textContent = userData.agent.name;
      document.getElementById('displayClientId').textContent = userData.agent.client_id;
      document.getElementById('agentScopes').textContent = userData.agent.scopes.join(', ');
      document.getElementById('secretCode').textContent = userData.secret;
    }
    
    function logout() {
      localStorage.removeItem('demo_token');
      localStorage.removeItem('demo_user');
      location.reload();
    }
  </script>
</body>
</html>
  `);
});

// Protected API endpoint that requires MachineAuth
app.get('/api/protected', async (req, res) => {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Missing or invalid authorization header' });
  }
  
  const token = authHeader.substring(7);
  
  try {
    // Verify token with MachineAuth
    const verifyResponse = await fetch(`${MACHINEAUTH_URL}/api/verify`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!verifyResponse.ok) {
      return res.status(401).json({ error: 'Invalid token' });
    }
    
    const verifyData = await verifyResponse.json();
    
    res.json({
      message: 'You accessed a protected API endpoint!',
      secret: verifyData.secret_code,
      agent: verifyData.agent,
      timestamp: new Date().toISOString()
    });
  } catch (error) {
    res.status(500).json({ error: 'Failed to verify token' });
  }
});

// Demo API - create a new agent
app.post('/api/demo/create-agent', async (req, res) => {
  try {
    const { name, scopes } = req.body;
    
    const response = await fetch(`${MACHINEAUTH_URL}/api/agents`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name: name || `demo-agent-${Date.now()}`,
        scopes: scopes || ['read', 'write']
      })
    });
    
    const data = await response.json();
    
    if (!response.ok) {
      return res.status(response.status).json(data);
    }
    
    res.json(data);
  } catch (error) {
    res.status(500).json({ error: 'Failed to create agent' });
  }
});

app.listen(PORT, () => {
  console.log(`
========================================
  🎉 MachineAuth Demo Website
========================================
  
  Website: http://localhost:${PORT}
  
  This demo shows:
  1. Users can login with MachineAuth credentials
  2. Protected content is only accessible with valid tokens
  3. AI agents can authenticate and access protected resources
  
  To test with an AI agent:
  1. First create an agent via API:
     curl -X POST ${MACHINEAUTH_URL}/api/agents \\
       -H "Content-Type: application/json" \\
       -d '{"name": "my-agent", "scopes": ["read", "write"]}'
  
  2. Use the client_id and client_secret to login above
  
  3. If the agent sees the secret code, it's NOT hallucinating!
  
========================================
  `);
});
