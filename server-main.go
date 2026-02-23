package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Agent struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ClientID         string    `json:"client_id"`
	ClientSecretHash string    `json:"client_secret_hash"`
	Scopes           []string  `json:"scopes"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type DB struct {
	Agents []Agent `json:"agents"`
}

var db DB
var dbFile = "/opt/machineauth/machineauth.json"
var jwtKey *rsa.PrivateKey

func main() {
	loadDB()

	var err error
	jwtKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"MachineAuth","status":"running"}`))
	})
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/oauth/token", tokenHandler)
	mux.HandleFunc("/.well-known/jwks.json", jwksHandler)
	mux.HandleFunc("/api/agents", agentsHandler)
	mux.HandleFunc("/api/secret", secretHandler)

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
			return
		}
		log.Fatal(err)
	}
	if len(data) > 0 {
		json.Unmarshal(data, &db)
	}
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
	} else {
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
		"iss":   "https://auth.writesomething.fun",
		"sub":   clientID,
		"iat":   float64(time.Now().Unix()),
		"exp":   float64(time.Now().Add(3600 * time.Second).Unix()),
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

func secretHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", 401)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization header", 401)
		return
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return &jwtKey.PublicKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid or expired token", 401)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Secret Page</title></head>
<body>
<h1>🔐 MachineAuth Protected Page</h1>
<p>Congratulations! Your agent successfully authenticated using JWT Bearer token.</p>
<hr>
<p><strong>Message:</strong> This is secret content only accessible with a valid token.</p>
<p><strong>Timestamp:</strong> ` + time.Now().Format(time.RFC3339) + `</p>
</body>
</html>`))
}
