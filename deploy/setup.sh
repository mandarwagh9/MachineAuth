#!/bin/bash
# AgentAuth Setup Script - Run on Ubuntu Server
# Copy this script to your server and run: bash setup.sh

set -e

echo "=== AgentAuth Setup ==="

# 1. Create directory
echo "[1/6] Creating directory..."
mkdir -p /opt/agentauth

# 2. If binary not present, download it
if [ ! -f /tmp/agentauth ]; then
    echo "[2/6] Binary not found at /tmp/agentauth"
    echo "Please upload the binary first, then run this script again"
    echo "To upload: pscp -scp agentauth-linux mandar@192.168.1.3:/tmp/agentauth"
    exit 1
fi

# 3. Copy binary
echo "[3/6] Installing binary..."
cp /tmp/agentauth /opt/agentauth/agentauth
chmod +x /opt/agentauth/agentauth

# 4. Create env file
echo "[4/6] Creating environment config..."
cat > /opt/agentauth/.env << 'EOF'
PORT=8081
ENV=production
DATABASE_URL=/opt/agentauth/agentauth.db
JWT_SIGNING_ALGORITHM=RS256
JWT_KEY_ID=key-1
JWT_ACCESS_TOKEN_EXPIRY=3600
ALLOWED_ORIGINS=https://writesomething.fun,https://auth.writesomething.fun
REQUIRE_HTTPS=true
EOF

# 5. Create systemd service
echo "[5/6] Creating systemd service..."
sudo tee /etc/systemd/system/agentauth.service > /dev/null << 'EOF'
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

# 6. Start service
echo "[6/6] Starting service..."
sudo systemctl daemon-reload
sudo systemctl enable agentauth
sudo systemctl restart agentauth

sleep 2

# Verify
echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Checking service status..."
sudo systemctl status agentauth --no-pager || true

echo ""
echo "Testing health endpoint..."
curl -s http://localhost:8081/health || echo "Health check failed"

echo ""
echo "Done! Next: Configure Cloudflare Tunnel"
