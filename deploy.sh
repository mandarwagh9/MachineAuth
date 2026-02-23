#!/bin/bash

# AgentAuth Deployment Script
# Run as: chmod +x deploy.sh && ./deploy.sh

set -e

echo "=== AgentAuth Deployment Script ==="

# Check if running as root or has sudo
if [ "$EUID" -ne 0 ] && ! sudo -v 2>/dev/null; then 
    echo "Please run as root or with sudo"
    exit 1
fi

# Create application directory
echo "Creating application directory..."
sudo mkdir -p /opt/agentauth
sudo chown $USER:$USER /opt/agentauth

# Copy binary
echo "Copying binary..."
cp bin/agentauth-linux /opt/agentauth/agentauth
chmod +x /opt/agentauth/agentauth

# Copy environment file
echo "Setting up environment..."
cp deploy/.env.production /opt/agentauth/.env

# Install SQLite if not present
echo "Checking SQLite..."
if ! command -v sqlite3 &> /dev/null; then
    echo "Installing SQLite..."
    sudo apt update && sudo apt install -y sqlite3
fi

# Create systemd service
echo "Creating systemd service..."
sudo tee /etc/systemd/system/agentauth.service > /dev/null <<'EOF'
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

# Reload systemd and start service
echo "Starting service..."
sudo systemctl daemon-reload
sudo systemctl enable agentauth
sudo systemctl restart agentauth

# Wait a moment and check status
sleep 2
echo ""
echo "=== Service Status ==="
sudo systemctl status agentauth --no-pager || true

echo ""
echo "=== Testing Health Endpoint ==="
curl -s http://localhost:8081/health || echo "Health check failed"

echo ""
echo "=== Deployment Complete ==="
echo "AgentAuth running at: http://localhost:8081"
echo ""
echo "Next steps:"
echo "1. Configure Cloudflare Tunnel for auth.writesomething.fun"
echo "2. Test agent creation: curl -X POST http://localhost:8081/api/agents -H 'Content-Type: application/json' -d '{\"name\":\"test\",\"scopes\":[\"read\"]}'"
