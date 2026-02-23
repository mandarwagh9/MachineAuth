#!/bin/bash
# AgentAuth Quick Install Script for Ubuntu
# Run: wget -O- https://raw.githubusercontent.com/... | bash

set -e

echo "=== Installing AgentAuth ==="

# Install Go (correct URL)
echo "[1/4] Installing Go..."
wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz -O /tmp/go.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo "Go installed: $(/usr/local/go/bin/go version)"

# Create directory
echo "[2/4] Creating app directory..."
mkdir -p /opt/agentauth
cd /opt/agentauth

# For now, let's just use a simpler approach - 
# create a minimal Go file that doesn't need CGO

# Download main.go
cat > main.go << 'MAINEOF'
package main

import (
    "context"
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
    "github.com/joho/godotenv"
    "golang.org/x/crypto/bcrypt"
)

type Config struct {
    Port             int
    Env              string
    DatabaseFile     string
    JWTKeyID         string
    JWTExpirySeconds int
    AllowedOrigins   string
}

func LoadConfig() *Config {
    godotenv.Load()
    return &Config{
        Port:             8081,
        Env:             "production",
        DatabaseFile:     "/opt/agentauth/agentauth.json",
        JWTKeyID:        "key-1",
        JWTExpirySeconds: 3600,
        AllowedOrigins:   "https://writesomething.fun,https://auth.writesomething.fun",
    }
}

type Agent struct {
    ID                string     `json:"id"`
    Name              string     `json:"name"`
    ClientID          string     `json:"client_id"`
    ClientSecretHash  string     `json:"client_secret_hash"`
    Scopes            []string   `json:"scopes"`
    IsActive          bool       `json:"is_active"`
    CreatedAt         time.Time  `json:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at"`
}

type DB struct {
    Agents []Agent `json:"agents"`
}

var db DB
var dbFile string
var jwtKey *rsa.PrivateKey

func main() {
    cfg := LoadConfig()
    dbFile = cfg.DatabaseFile
    
    loadDB()
    
    var err error
    jwtKey, err = generateKey()
    if err != nil {
        log.Fatal(err)
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/health", health)
    mux.HandleFunc("/oauth/token", tokenHandler)
    mux.HandleFunc("/.well-known/jwks.json", jwksHandler)
    mux.HandleFunc("/api/agents", agentsHandler)
    
    server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: mux}
    
    go func() {
        log.Printf("Server starting on port %d", cfg.Port)
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
        hash, _ := bcrypt.GenerateFromPassword([]byte(clientSecret), 10)
        
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
        "iss": "https://auth.writesomething.fun",
        "sub": clientID,
        "iat": time.Now().Unix(),
        "exp": time.Now().Add(3600*time.Second).Unix(),
        "scope": agent.Scopes,
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = "key-1"
    tokenString, _ := token.SignedString(jwtKey)
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "access_token": tokenString,
        "token_type":   "Bearer",
        "expires_in":   3600,
    })
}

func jwksHandler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{
        "keys": []map[string]string{
            {"kty": "RSA", "kid": "key-1", "use": "sig", "alg": "RS256"},
        },
    })
}

func generateKey() (*rsa.PrivateKey, error) {
    return rsa.GenerateKey(rand.Reader, 2048)
}
MAINEOF

echo "[3/4] Building..."
/usr/local/go/bin/go mod init agentauth
/usr/local/go/bin/go get github.com/golang-jwt/jwt/v5
/usr/local/go/bin/go get github.com/google/uuid
/usr/local/go/bin/go get github.com/joho/godotenv
/usr/local/go/bin/go get golang.org/x/crypto
/usr/local/go/bin/go build -o agentauth .

echo "[4/4] Setting up service..."
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

echo "Done! Testing..."
sleep 2
curl -s http://localhost:8081/health
