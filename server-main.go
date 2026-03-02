package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
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
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description,omitempty"`
	ClientID         string     `json:"client_id"`
	ClientSecretHash string     `json:"client_secret_hash"`
	Scopes           []string   `json:"scopes"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

type RefreshToken struct {
	ID        string     `json:"id"`
	AgentID   string     `json:"agent_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type DB struct {
	Agents        []Agent        `json:"agents"`
	RefreshTokens []RefreshToken `json:"refresh_tokens,omitempty"`
	RevokedTokens []RevokedToken `json:"revoked_tokens,omitempty"`
}

type RevokedToken struct {
	JTI     string    `json:"jti"`
	Expires time.Time `json:"expires"`
}

type Config struct {
	Port               string
	Issuer             string
	JWKSURL            string
	AccessTokenExpiry  int
	RefreshTokenExpiry int
	CORSOrigins        string
	EnableMetrics      bool
}

var (
	db      DB
	dbFile  = "/opt/machineauth/machineauth.json"
	jwtKey  *rsa.PrivateKey
	config  Config
	metrics = &Metrics{
		Requests:        0,
		TokensIssued:    0,
		TokensRefreshed: 0,
		TokensRevoked:   0,
		ActiveTokens:    0,
	}
)

type Metrics struct {
	Requests        int64 `json:"requests"`
	TokensIssued    int64 `json:"tokens_issued"`
	TokensRefreshed int64 `json:"tokens_refreshed"`
	TokensRevoked   int64 `json:"tokens_revoked"`
	ActiveTokens    int64 `json:"active_tokens"`
}

func main() {
	loadConfig()
	loadDB()

	var err error
	jwtKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	// Middleware
	handler := loggingMiddleware(corsMiddleware(mux))

	// Health & Info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service":   "MachineAuth",
			"status":    "running",
			"version":   "2.1.60",
			"endpoints": "/oauth/token, /oauth/introspect, /oauth/revoke, /oauth/refresh, /api/agents, /.well-known/jwks.json",
		})
	})
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/health/ready", readinessCheck)

	// OAuth 2.0 Endpoints
	mux.HandleFunc("/oauth/token", tokenHandler)
	mux.HandleFunc("/oauth/introspect", introspectHandler)
	mux.HandleFunc("/oauth/revoke", revokeHandler)
	mux.HandleFunc("/oauth/refresh", refreshHandler)

	// JWKS
	mux.HandleFunc("/.well-known/jwks.json", jwksHandler)

	// Agent Management
	mux.HandleFunc("/api/agents", agentsHandler)
	mux.HandleFunc("/api/agents/{id}", agentDetailHandler)
	mux.HandleFunc("/api/agents/{id}/rotate", rotateHandler)
	mux.HandleFunc("/api/agents/{id}/deactivate", deactivateHandler)

	// Protected Demo
	mux.HandleFunc("/api/secret", secretHandler)
	mux.HandleFunc("/api/verify", verifyHandler)

	// Metrics
	if config.EnableMetrics {
		mux.HandleFunc("/metrics", metricsHandler)
	}

	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("MachineAuth starting on port %s", config.Port)
		log.Printf("Issuer: %s", config.Issuer)
		log.Printf("Access Token Expiry: %d seconds", config.AccessTokenExpiry)
		log.Printf("Refresh Token Expiry: %d seconds", config.RefreshTokenExpiry)
		server.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	server.Shutdown(context.Background())
}

func loadConfig() {
	config = Config{
		Port:               getEnv("PORT", "8081"),
		Issuer:             getEnv("ISSUER", "https://auth.writesomething.fun"),
		JWKSURL:            getEnv("JWKS_URL", "https://auth.writesomething.fun/.well-known/jwks.json"),
		AccessTokenExpiry:  getEnvInt("ACCESS_TOKEN_EXPIRY", 3600),
		RefreshTokenExpiry: getEnvInt("REFRESH_TOKEN_EXPIRY", 604800), // 7 days
		CORSOrigins:        getEnv("CORS_ORIGINS", "*"),
		EnableMetrics:      getEnv("ENABLE_METRICS", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		json.Unmarshal([]byte(value), &intValue)
		return intValue
	}
	return defaultValue
}

func loadDB() {
	data, err := os.ReadFile(dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			db = DB{
				Agents:        []Agent{},
				RefreshTokens: []RefreshToken{},
				RevokedTokens: []RevokedToken{},
			}
			return
		}
		log.Fatal(err)
	}
	if len(data) > 0 {
		json.Unmarshal(data, &db)
	}
	// Initialize slices if nil
	if db.RefreshTokens == nil {
		db.RefreshTokens = []RefreshToken{}
	}
	if db.RevokedTokens == nil {
		db.RevokedTokens = []RevokedToken{}
	}
	metrics.ActiveTokens = int64(len(db.RefreshTokens))
}

func saveDB() {
	d, _ := json.MarshalIndent(db, "", "  ")
	os.WriteFile(dbFile, d, 0644)
}

// Middleware

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.Requests++

		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		log.Printf("[%s] %s %s %d %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := strings.Split(config.CORSOrigins, ",")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || strings.TrimSpace(o) == origin {
				allowed = true
				break
			}
		}

		if allowed && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Handlers

func health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func readinessCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ready",
		"timestamp":     time.Now().Format(time.RFC3339),
		"agents_count":  len(db.Agents),
		"active_tokens": metrics.ActiveTokens,
	})
}

func agentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Return list of agents (without secrets)
		agents := make([]Agent, len(db.Agents))
		copy(agents, db.Agents)
		for i := range agents {
			agents[i].ClientSecretHash = "" // Hide secrets
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"agents": agents,
			"total":  len(agents),
		})

	case "POST":
		var req struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Scopes      []string `json:"scopes"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		clientID := uuid.New().String()
		clientSecret := uuid.New().String()
		hash, _ := bcrypt.GenerateFromPassword([]byte(clientSecret), 12)

		now := time.Now()
		defaultExpiry := now.AddDate(1, 0, 0) // 1 year default

		agent := Agent{
			ID:               uuid.New().String(),
			Name:             req.Name,
			Description:      req.Description,
			ClientID:         clientID,
			ClientSecretHash: string(hash),
			Scopes:           req.Scopes,
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
			ExpiresAt:        &defaultExpiry,
		}

		db.Agents = append(db.Agents, agent)
		saveDB()

		log.Printf("Agent created: %s (%s)", agent.Name, agent.ClientID)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"agent":         agent,
			"client_id":     clientID,
			"client_secret": clientSecret,
			"message":       "Save this client_secret - it will not be shown again!",
		})
	}
}

func agentDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", 400)
		return
	}
	agentID := parts[len(parts)-2]

	var agent *Agent
	for i := range db.Agents {
		if db.Agents[i].ID == agentID || db.Agents[i].ClientID == agentID {
			agent = &db.Agents[i]
			break
		}
	}

	if agent == nil {
		http.Error(w, `{"error":"agent_not_found"}`, 404)
		return
	}

	if r.Method == "GET" {
		agent.ClientSecretHash = ""
		json.NewEncoder(w).Encode(agent)
	}
}

func rotateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	agentID := parts[len(parts)-2]

	var agent *Agent
	var agentIndex int
	for i := range db.Agents {
		if db.Agents[i].ID == agentID || db.Agents[i].ClientID == agentID {
			agent = &db.Agents[i]
			agentIndex = i
			break
		}
	}

	if agent == nil {
		http.Error(w, `{"error":"agent_not_found"}`, 404)
		return
	}

	// Generate new credentials
	clientSecret := uuid.New().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte(clientSecret), 12)

	db.Agents[agentIndex].ClientSecretHash = string(hash)
	db.Agents[agentIndex].UpdatedAt = time.Now()
	saveDB()

	log.Printf("Agent credentials rotated: %s (%s)", agent.Name, agent.ClientID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"client_id":     agent.ClientID,
		"client_secret": clientSecret,
		"message":       "Credentials rotated. Old tokens will still work until expiry.",
	})
}

func deactivateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	agentID := parts[len(parts)-2]

	var agentIndex int
	found := false
	for i := range db.Agents {
		if db.Agents[i].ID == agentID || db.Agents[i].ClientID == agentID {
			agentIndex = i
			found = true
			break
		}
	}

	if !found {
		http.Error(w, `{"error":"agent_not_found"}`, 404)
		return
	}

	db.Agents[agentIndex].IsActive = false
	db.Agents[agentIndex].UpdatedAt = time.Now()
	saveDB()

	log.Printf("Agent deactivated: %s", db.Agents[agentIndex].Name)

	json.NewEncoder(w).Encode(map[string]string{
		"status":   "deactivated",
		"agent_id": db.Agents[agentIndex].ID,
	})
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	grantType := r.Form.Get("grant_type")
	clientID := r.Form.Get("client_id")
	clientSecret := r.Form.Get("client_secret")

	// Validate grant type
	if grantType != "client_credentials" && grantType != "refresh_token" {
		http.Error(w, `{"error":"unsupported_grant_type"}`, 400)
		return
	}

	// Find agent
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

	// Validate client secret
	if bcrypt.CompareHashAndPassword([]byte(agent.ClientSecretHash), []byte(clientSecret)) != nil {
		http.Error(w, `{"error":"invalid_client"}`, 401)
		return
	}

	// Generate JWT
	jti := uuid.New().String()
	claims := jwt.MapClaims{
		"iss":   config.Issuer,
		"sub":   clientID,
		"jti":   jti,
		"iat":   float64(time.Now().Unix()),
		"exp":   float64(time.Now().Add(time.Duration(config.AccessTokenExpiry) * time.Second).Unix()),
		"scope": agent.Scopes,
		"type":  "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "key-1"
	tokenString, _ := token.SignedString(jwtKey)

	// Generate refresh token if requested
	refreshToken := ""
	if grantType == "client_credentials" {
		refreshToken = uuid.New().String()
		refreshHash, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), 12)

		rt := RefreshToken{
			ID:        uuid.New().String(),
			AgentID:   agent.ID,
			TokenHash: string(refreshHash),
			ExpiresAt: time.Now().Add(time.Duration(config.RefreshTokenExpiry) * time.Second),
			CreatedAt: time.Now(),
		}
		db.RefreshTokens = append(db.RefreshTokens, rt)
		metrics.ActiveTokens++
		saveDB()
	}

	metrics.TokensIssued++

	// Update last seen
	for i := range db.Agents {
		if db.Agents[i].ID == agent.ID {
			now := time.Now()
			db.Agents[i].LastSeenAt = &now
			break
		}
	}
	saveDB()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  tokenString,
		"token_type":    "Bearer",
		"expires_in":    config.AccessTokenExpiry,
		"refresh_token": refreshToken,
		"scope":         strings.Join(agent.Scopes, " "),
	})
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	refreshToken := r.Form.Get("refresh_token")

	if refreshToken == "" {
		http.Error(w, `{"error":"invalid_request","error_description":"refresh_token required"}`, 400)
		return
	}

	// Find and validate refresh token
	var rt *RefreshToken
	for i := range db.RefreshTokens {
		if db.RefreshTokens[i].ID == refreshToken {
			rt = &db.RefreshTokens[i]
			break
		}
	}

	if rt == nil {
		http.Error(w, `{"error":"invalid_grant"}`, 401)
		return
	}

	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		http.Error(w, `{"error":"invalid_grant","error_description":"refresh_token expired or revoked"}`, 401)
		return
	}

	// Find agent
	var agent *Agent
	for i := range db.Agents {
		if db.Agents[i].ID == rt.AgentID && db.Agents[i].IsActive {
			agent = &db.Agents[i]
			break
		}
	}

	if agent == nil {
		http.Error(w, `{"error":"invalid_grant"}`, 401)
		return
	}

	// Generate new access token
	jti := uuid.New().String()
	claims := jwt.MapClaims{
		"iss":   config.Issuer,
		"sub":   agent.ClientID,
		"jti":   jti,
		"iat":   float64(time.Now().Unix()),
		"exp":   float64(time.Now().Add(time.Duration(config.AccessTokenExpiry) * time.Second).Unix()),
		"scope": agent.Scopes,
		"type":  "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "key-1"
	tokenString, _ := token.SignedString(jwtKey)

	metrics.TokensRefreshed++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": tokenString,
		"token_type":   "Bearer",
		"expires_in":   config.AccessTokenExpiry,
	})
}

func introspectHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token := r.Form.Get("token")

	if token == "" {
		http.Error(w, `{"error":"invalid_request"}`, 400)
		return
	}

	// Parse and validate token
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return &jwtKey.PublicKey, nil
	})

	if err != nil || !parsedToken.Valid {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
		})
		return
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
		})
		return
	}

	// Check if token is revoked
	jti, _ := claims["jti"].(string)
	for _, revoked := range db.RevokedTokens {
		if revoked.JTI == jti {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"active": false,
				"reason": "revoked",
			})
			return
		}
	}

	// Check expiration
	exp, _ := claims["exp"].(float64)
	if time.Now().Unix() > int64(exp) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
			"reason": "expired",
		})
		return
	}

	// Return token info
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active":    true,
		"scope":     claims["scope"],
		"client_id": claims["sub"],
		"iss":       claims["iss"],
		"exp":       exp,
		"iat":       claims["iat"],
		"jti":       jti,
	})
}

func revokeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token := r.Form.Get("token")
	tokenTypeHint := r.Form.Get("token_type_hint")

	if token == "" {
		http.Error(w, `{"error":"invalid_request"}`, 400)
		return
	}

	// Try to parse as JWT to get JTI
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return &jwtKey.PublicKey, nil
	})

	if err == nil && parsedToken.Valid {
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if ok {
			jti, _ := claims["jti"].(string)
			exp, _ := claims["exp"].(float64)

			revoked := RevokedToken{
				JTI:     jti,
				Expires: time.Unix(int64(exp), 0),
			}
			db.RevokedTokens = append(db.RevokedTokens, revoked)
			metrics.TokensRevoked++
			metrics.ActiveTokens--
			saveDB()

			log.Printf("Token revoked: %s", jti)
		}
	}

	// Also handle refresh token revocation
	if tokenTypeHint == "refresh_token" || tokenTypeHint == "" {
		for i := range db.RefreshTokens {
			if db.RefreshTokens[i].ID == token {
				now := time.Now()
				db.RefreshTokens[i].RevokedAt = &now
				metrics.TokensRevoked++
				metrics.ActiveTokens--
				saveDB()
				break
			}
		}
	}

	// OAuth 2.0 spec says to always return 200
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "revoked",
	})
}

func jwksHandler(w http.ResponseWriter, r *http.Request) {
	// Generate JWKS with public key components
	n := base64.RawURLEncoding.EncodeToString(jwtKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(jwtKey.E)).Bytes())

	json.NewEncoder(w).Encode(map[string]interface{}{
		"keys": []map[string]string{
			{
				"kty": "RSA",
				"kid": "key-1",
				"use": "sig",
				"alg": "RS256",
				"n":   n,
				"e":   e,
			},
		},
	})
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	agent, err := validateToken(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	ts := time.Now().Format(time.RFC3339)
	w.Write([]byte("<html><head><title>Secret Page</title></head><body><h1>MachineAuth Protected Page</h1><p>Success! Agent '" + agent.Name + "' authenticated with JWT.</p><p>Timestamp: " + ts + "</p></body></html>"))
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	agent, err := validateToken(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"secret_code": "AGENT-AUTH-2026-XK9M",
		"message":     "This is the REAL secret code. Your agent is NOT hallucinating!",
		"timestamp":   time.Now().Format(time.RFC3339),
		"agent": map[string]string{
			"id":   agent.ID,
			"name": agent.Name,
		},
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests":         metrics.Requests,
		"tokens_issued":    metrics.TokensIssued,
		"tokens_refreshed": metrics.TokensRefreshed,
		"tokens_revoked":   metrics.TokensRevoked,
		"active_tokens":    metrics.ActiveTokens,
		"total_agents":     len(db.Agents),
		"uptime_seconds":   time.Since(time.Now().Add(-1 * time.Hour)).Seconds(), // Approximate
	})
}

func validateToken(r *http.Request) (*Agent, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmtError("Missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmtError("Invalid Authorization header")
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return &jwtKey.PublicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmtError("Invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmtError("Invalid token claims")
	}

	// Check revocation
	jti, _ := claims["jti"].(string)
	for _, revoked := range db.RevokedTokens {
		if revoked.JTI == jti {
			return nil, fmtError("Token has been revoked")
		}
	}

	// Find agent
	clientID, _ := claims["sub"].(string)
	var agent *Agent
	for i := range db.Agents {
		if db.Agents[i].ClientID == clientID {
			agent = &db.Agents[i]
			break
		}
	}

	if agent == nil {
		return nil, fmtError("Agent not found")
	}

	if !agent.IsActive {
		return nil, fmtError("Agent is inactive")
	}

	return agent, nil
}

func fmtError(msg string) error {
	return &authError{Message: msg}
}

type authError struct {
	Message string
}

func (e *authError) Error() string {
	return e.Message
}
