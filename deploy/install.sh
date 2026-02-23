#!/bin/bash
# AgentAuth Quick Install - Run this on your Ubuntu server

set -e

echo "=== Installing AgentAuth ==="

# Install Go using apt
echo "[1/5] Installing Go..."
sudo apt update
sudo apt install -y golang-go

echo "Go version: $(go version)"

# Create directory
echo "[2/5] Creating app directory..."
mkdir -p /opt/agentauth
cd /opt/agentauth

# Create main.go
echo "[3/5] Creating application..."
cat > main.go << 'MAINEOF'
package main

import (
    "context"
    "crypto/rand"
    "crypto/rsa"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

type Agent struct {
    ID                string    `json:"id"`
    Name              string    `json:"name"`
    ClientID          string    `json:"client_id"`
    ClientSecretHash  string    `json:"client_secret_hash"`
    Scopes            []string  `json:"scopes"`
    IsActive          bool      `json:"is_active"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}

type DB struct {
    Agents []Agent `json:"agents"`
}

var db DB
var dbFile = "/opt/agentauth/agentauth.json"
var jwtKey *rsa.PrivateKey

func main() {
    loadDB()
    
    var err error
    jwtKey, err = rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        log.Fatal(err)
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/health", health)
    mux.HandleFunc("/oauth/token", tokenHandler)
    mux.HandleFunc("/.well-known/jwks.json", jwksHandler)
    mux.HandleFunc("/api/agents", agentsHandler)
    
    server := &http.Server{Addr: ":8081", Handler: mux}
    
    go func() {
        log.Printf("Server starting on port 8081")
        server.ListenAndServe()
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down...")
    server.Shutdown(context.Background())
}

func loadDB() {
    data, err := os.ReadFile(dbFile)
    if err != nil {
        if os.IsNotExist(err) {
            db = DB{Agents: []Agent{}}
            saveDB()
            return
        }
        log.Fatal(err)
    }
    json.Unmarshal(data, &db)
}

func saveDB() {
    d, _ := json.MarshalIndent(db, "", "  ")
    os.WriteFile(dbFile, d, 0644)
}

func health(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func agentsHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        var req struct {
            Name   string   `json:"name"`
            Scopes []string `json:"scopes"`
        }
        json.NewDecoder(r.Body).Decode(&req)
        
        clientID := uuid.New().String()
        clientSecret := uuid.New().String()
        hash, _ := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
        
        agent := Agent{
            ID:               uuid.New().String(),
            Name:             req.Name,
            ClientID:         clientID,
            ClientSecretHash: string(hash),
            Scopes:           req.Scopes,
            IsActive:         true,
            CreatedAt:        time.Now(),
            UpdatedAt:        time.Now(),
        }
        
        db.Agents = append(db.Agents, agent)
        saveDB()
        
        json.NewEncoder(w).Encode(map[string]interface{}{
            "agent":         agent,
            "client_id":     clientID,
            "client_secret": clientSecret,
        })
    } else if r.Method == "GET" {
        json.NewEncoder(w).Encode(map[string]interface{}{"agents": db.Agents})
    }
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    clientID := r.Form.Get("client_id")
    clientSecret := r.Form.Get("client_secret")
    
    var agent *Agent
    for i := range db.Agents {
        if db.Agents[i].ClientID == clientID {
            agent = &db.Agents[i]
            break
        }
    }
    
    if agent == nil || !agent.IsActive {
        http.Error(w, `{"error":"invalid_client"}`, 401)
        return
    }
    
    if bcrypt.CompareHashAndPassword([]byte(agent.ClientSecretHash), []byte(clientSecret)) != nil {
        http.Error(w, `{"error":"invalid_client"}`, 401)
        return
    }
    
    claims := jwt.MapClaims{
        "iss":  "https://auth.writesomething.fun",
        "sub":  clientID,
        "iat":  float64(time.Now().Unix()),
        "exp":  float64(time.Now().Add(3600*time.Second).Unix()),
        "scope": agent.Scopes,
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = "key-1"
    tokenString, _ := token.SignedString(jwtKey)
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "access_token": tokenString,
        "token_type":  "Bearer",
        "expires_in":  3600,
    })
}

func jwksHandler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{
        "keys": []map[string]string{
            {"kty": "RSA", "kid": "key-1", "use": "sig", "alg": "RS256"},
        },
    })
}
MAINEOF

echo "[4/5] Building..."
go mod init agentauth
go mod tidy
go build -o agentauth .

echo "[5/5] Setting up service..."
sudo tee /etc/systemd/system/agentauth.service > /dev/null << 'EOF'
[Unit]
Description=AgentAuth
After=network.target

[Service]
Type=simple
User=mandar
WorkingDirectory=/opt/agentauth
ExecStart=/opt/agentauth/agentauth
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable agentauth
sudo systemctl start agentauth

echo ""
echo "=== Done! ==="
echo "Testing..."
sleep 2
curl -s http://localhost:8081/health
echo ""
echo "Create agent:"
echo 'curl -X POST http://localhost:8081/api/agents -H "Content-Type: application/json" -d '\''{"name":"test","scopes":["read"]}\'''
