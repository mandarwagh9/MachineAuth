#!/bin/bash
# MachineAuth Full Deployment Script
# Deploys backend (auth.writesomething.fun) + admin UI (authadmin.writesomething.fun)
# Server: mandar@192.168.1.3 | Uses Cloudflare Tunnel for HTTPS
set -e

echo "============================================="
echo "  MachineAuth Full Deployment"
echo "  Backend:  auth.writesomething.fun"
echo "  Admin UI: authadmin.writesomething.fun"
echo "============================================="

DEPLOY_DIR="/opt/machineauth"
REPO_DIR="/home/mandar/MachineAuth"
GO_VERSION="1.23.6"
NODE_VERSION="20"
PG_DB="agentauth"
PG_USER="machineauth"
PG_PASS="machineauth_secret_2025"

# ─── 1. System Dependencies ─────────────────────────────────────────

echo ""
echo "[1/11] Checking system dependencies..."

# Install essential packages
sudo apt-get update -qq
sudo apt-get install -y -qq git curl wget nginx postgresql postgresql-contrib > /dev/null 2>&1

# ─── 2. Setup PostgreSQL ─────────────────────────────────────────────

echo "[2/11] Setting up PostgreSQL..."

# Ensure PostgreSQL is running
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Create database and user if they don't exist
sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='${PG_USER}'" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE USER ${PG_USER} WITH PASSWORD '${PG_PASS}';"
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='${PG_DB}'" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE DATABASE ${PG_DB} OWNER ${PG_USER};"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE ${PG_DB} TO ${PG_USER};"
sudo -u postgres psql -d "${PG_DB}" -c "GRANT ALL ON SCHEMA public TO ${PG_USER};"
echo "  PostgreSQL ready: ${PG_DB} database with ${PG_USER} user"

# ─── 3. Install Go ──────────────────────────────────────────────────

echo "[3/11] Checking Go installation..."

GO_INSTALLED=$(go version 2>/dev/null | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")
GO_REQUIRED="1.23"

if [ "$(printf '%s\n' "$GO_REQUIRED" "$GO_INSTALLED" | sort -V | head -n1)" != "$GO_REQUIRED" ]; then
    echo "  Installing Go ${GO_VERSION}..."
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    # Add to PATH if not already there
    if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:/usr/local/go/bin
    echo "  Go $(go version) installed"
else
    echo "  Go ${GO_INSTALLED} already installed (>= ${GO_REQUIRED})"
fi

# ─── 4. Install Node.js ─────────────────────────────────────────────

echo "[4/11] Checking Node.js installation..."

if ! command -v node &> /dev/null || [ "$(node -v | grep -oP 'v\K[0-9]+')" -lt "$NODE_VERSION" ]; then
    echo "  Installing Node.js ${NODE_VERSION}..."
    curl -fsSL "https://deb.nodesource.com/setup_${NODE_VERSION}.x" | sudo -E bash - > /dev/null 2>&1
    sudo apt-get install -y -qq nodejs > /dev/null 2>&1
    echo "  Node.js $(node -v) installed"
else
    echo "  Node.js $(node -v) already installed"
fi

# ─── 5. Clone/Pull Latest Code ──────────────────────────────────────

echo "[5/11] Getting latest code from GitHub..."

if [ -d "$REPO_DIR/.git" ]; then
    cd "$REPO_DIR"
    git fetch origin
    git reset --hard origin/master
    echo "  Updated existing repo"
else
    git clone https://github.com/mandarwagh9/MachineAuth.git "$REPO_DIR"
    cd "$REPO_DIR"
    echo "  Cloned fresh repo"
fi

# ─── 6. Build Go Backend ────────────────────────────────────────────

echo "[6/11] Building Go backend..."

cd "$REPO_DIR"
export PATH=$PATH:/usr/local/go/bin

# Ensure deploy directory exists before build
sudo mkdir -p "${DEPLOY_DIR}"
sudo mkdir -p "${DEPLOY_DIR}/keys"
sudo mkdir -p "${DEPLOY_DIR}/web"
sudo chown -R mandar:mandar "${DEPLOY_DIR}"

go mod download
CGO_ENABLED=0 go build -o "${DEPLOY_DIR}/machineauth" ./cmd/server
echo "  Backend binary built: ${DEPLOY_DIR}/machineauth"

# ─── 7. Build React Frontend ────────────────────────────────────────

echo "[7/11] Building React frontend..."

cd "$REPO_DIR/web"
npm ci --silent 2>/dev/null || npm install --silent
npm run build
echo "  Frontend built to: ${REPO_DIR}/web/dist/"

# ─── 8. Set up deploy directory ─────────────────────────────────────

echo "[8/11] Setting up deployment directory..."

sudo mkdir -p "${DEPLOY_DIR}"
sudo mkdir -p "${DEPLOY_DIR}/keys"
sudo mkdir -p "${DEPLOY_DIR}/web"
sudo chown -R mandar:mandar "${DEPLOY_DIR}"

# Copy frontend build
cp -r "$REPO_DIR/web/dist" "${DEPLOY_DIR}/web/dist"

# Create backend .env
cat > "${DEPLOY_DIR}/.env" << ENVEOF
PORT=8080
ENV=production
DATABASE_URL=postgres://${PG_USER}:${PG_PASS}@localhost:5432/${PG_DB}?sslmode=disable
JWT_SIGNING_ALGORITHM=RS256
JWT_KEY_ID=key-1
JWT_KEY_PATH=/opt/machineauth/keys
JWT_ACCESS_TOKEN_EXPIRY=3600
JWT_ISSUER=https://auth.writesomething.fun
ALLOWED_ORIGINS=https://authadmin.writesomething.fun,https://auth.writesomething.fun
REQUIRE_HTTPS=true
ADMIN_EMAIL=admin
ADMIN_PASSWORD=admin
WEBHOOK_WORKER_COUNT=3
WEBHOOK_MAX_RETRIES=10
WEBHOOK_TIMEOUT_SECS=10
ENVEOF

echo "  Deployment directory ready"

# ─── 9. Configure nginx for Admin UI ────────────────────────────────

echo "[9/11] Configuring nginx for admin UI..."

sudo tee /etc/nginx/sites-available/machineauth-admin > /dev/null << 'NGINXEOF'
server {
    listen 3000;
    server_name authadmin.writesomething.fun localhost;

    root /opt/machineauth/web/dist;
    index index.html;

    # SPA fallback - serve index.html for all non-file routes
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API calls to the Go backend
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Proxy OAuth endpoints
    location /oauth/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Proxy well-known endpoints
    location /.well-known/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Proxy health endpoint
    location /health {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Proxy metrics endpoint
    location /metrics {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Gzip compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript;
    gzip_min_length 1000;

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
NGINXEOF

# Enable the site
sudo ln -sf /etc/nginx/sites-available/machineauth-admin /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
sudo nginx -t
echo "  Nginx configured on port 3000"

# ─── 10. Create systemd services ───────────────────────────────────

echo "[10/11] Creating systemd services..."

# Backend service
sudo tee /etc/systemd/system/machineauth.service > /dev/null << 'SVCEOF'
[Unit]
Description=MachineAuth - OAuth 2.0 Backend API
After=network.target

[Service]
Type=simple
User=mandar
WorkingDirectory=/opt/machineauth
ExecStart=/opt/machineauth/machineauth
EnvironmentFile=/opt/machineauth/.env
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SVCEOF

echo "  systemd services created"

# ─── 11. Update Cloudflare Tunnel ───────────────────────────────────

echo "[11/11] Updating Cloudflare tunnel config..."

CLOUDFLARED_CONFIG="/home/mandar/.cloudflared/config.yml"

if [ -f "$CLOUDFLARED_CONFIG" ]; then
    # Backup existing config
    cp "$CLOUDFLARED_CONFIG" "${CLOUDFLARED_CONFIG}.bak"
fi

cat > "$CLOUDFLARED_CONFIG" << 'CFEOF'
tunnel: 82377467-2594-41e3-a760-84c64a788819
credentials-file: /home/mandar/.cloudflared/82377467-2594-41e3-a760-84c64a788819.json

ingress:
  - hostname: auth.writesomething.fun
    service: http://localhost:8080
  - hostname: authadmin.writesomething.fun
    service: http://localhost:3000
  - hostname: wiki.writesomething.fun
    service: http://localhost:8090
  - service: http_status:404
CFEOF

echo "  Cloudflare tunnel config updated"

# ─── Start Everything ───────────────────────────────────────────────

echo ""
echo "============================================="
echo "  Starting all services..."
echo "============================================="

# Stop old services if running
sudo systemctl stop agentauth 2>/dev/null || true
sudo systemctl disable agentauth 2>/dev/null || true

# Reload systemd
sudo systemctl daemon-reload

# Start backend
sudo systemctl enable machineauth
sudo systemctl restart machineauth
echo "  ✓ Backend started (port 8080)"

# Start nginx (serves admin UI on port 3000)
sudo systemctl enable nginx
sudo systemctl restart nginx
echo "  ✓ Nginx/Admin UI started (port 3000)"

# Restart cloudflared
if systemctl is-active --quiet cloudflared 2>/dev/null; then
    sudo systemctl restart cloudflared
    echo "  ✓ Cloudflare tunnel restarted"
elif command -v cloudflared &> /dev/null; then
    # If running as a manual process, restart it
    pkill cloudflared 2>/dev/null || true
    sleep 1
    nohup cloudflared tunnel run > /tmp/cloudflared.log 2>&1 &
    echo "  ✓ Cloudflare tunnel started (manual mode)"
else
    echo "  ⚠ cloudflared not found — install it or start tunnel manually"
fi

# Wait for services to start
sleep 3

# ─── Verification ───────────────────────────────────────────────────

echo ""
echo "============================================="
echo "  Verification"
echo "============================================="

# Check backend health
echo -n "  Backend health: "
curl -sf http://localhost:8080/health 2>/dev/null && echo "" || echo "FAILED"

# Check nginx
echo -n "  Admin UI (nginx): "
curl -sf http://localhost:3000/ > /dev/null 2>&1 && echo "OK" || echo "FAILED"

# Check backend via admin proxy
echo -n "  Admin -> Backend proxy: "
curl -sf http://localhost:3000/health 2>/dev/null && echo "" || echo "FAILED"

# Service status
echo ""
echo "  Service Status:"
echo "  - machineauth: $(systemctl is-active machineauth)"
echo "  - nginx:       $(systemctl is-active nginx)"
echo "  - cloudflared: $(systemctl is-active cloudflared 2>/dev/null || echo 'unknown')"

echo ""
echo "============================================="
echo "  Deployment Complete!"
echo ""
echo "  Backend API:  https://auth.writesomething.fun"
echo "  Admin UI:     https://authadmin.writesomething.fun"
echo "  Admin Login:  admin / admin"
echo ""
echo "  ⚠ Change ADMIN_PASSWORD in ${DEPLOY_DIR}/.env"
echo "============================================="
