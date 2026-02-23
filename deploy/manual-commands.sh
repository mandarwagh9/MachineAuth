#!/bin/bash
# Manual Deployment Commands for AgentAuth
# Run these commands on your Ubuntu server

# 1. Create directory
mkdir -p /opt/agentauth

# 2. Copy files from local machine:
#    - Copy bin/agentauth-linux to /opt/agentauth/agentauth
#    - Copy deploy/.env.production to /opt/agentauth/.env

# 3. Make executable
chmod +x /opt/agentauth/agentauth

# 4. Install SQLite (if not installed)
sudo apt update && sudo apt install -y sqlite3

# 5. Create systemd service
sudo tee /etc/systemd/system/agentauth.service << 'EOF'
[Unit]
Description=AgentAuth - AI Agent Authentication Service
After=network.target

[Service]
Type=simple
User=mandar
WorkingDirectory=/opt/agentauth
ExecStart=/opt/agentauth/agentauth
EnvironmentFile=/opt/agentauth/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 6. Enable and start
sudo systemctl daemon-reload
sudo systemctl enable agentauth
sudo systemctl start agentauth

# 7. Check status
sudo systemctl status agentauth

# 8. Test
curl http://localhost:8081/health
