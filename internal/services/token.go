package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"machineauth/internal/config"
	"machineauth/internal/db"
	"machineauth/internal/models"
)

type TokenService struct {
	cfg             *config.Config
	privateKey      *rsa.PrivateKey
	keyID           string
	mu              sync.RWMutex
	tokenExpirySecs int
	db              *db.DB
}

func NewTokenService(cfg *config.Config, database *db.DB) (*TokenService, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &TokenService{
		cfg:             cfg,
		privateKey:      privateKey,
		keyID:           cfg.JWTKeyID,
		tokenExpirySecs: cfg.JWTExpirySeconds,
		db:              database,
	}, nil
}

func (s *TokenService) GenerateToken(agent *models.Agent, requestedScope string) (*models.TokenResponse, error) {
	scopes := agent.Scopes
	if requestedScope != "" {
		requestedScopes := parseScope(requestedScope)
		scopes = filterScopes(scopes, requestedScopes)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":      "https://auth.example.com",
		"sub":      agent.ClientID,
		"agent_id": agent.ID.String(),
		"org_id":   agent.OrganizationID,
		"team_id":  agent.TeamID.String(),
		"aud":      "machineauth-api",
		"iat":      now.Unix(),
		"exp":      now.Add(s.cfg.GetTokenExpiry()).Unix(),
		"scope":    scopes,
		"jti":      generateTokenID(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID

	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &models.TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   s.tokenExpirySecs,
		Scope:       joinScopes(scopes),
		IssuedAt:    now.Unix(),
	}, nil
}

func (s *TokenService) JWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	jwks := models.JWKS{
		Keys: []models.JWK{
			{
				Kty: "RSA",
				Kid: s.keyID,
				Use: "sig",
				Alg: "RS256",
				N:   base64.RawURLEncoding.EncodeToString(s.privateKey.PublicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.privateKey.PublicKey.E)).Bytes()),
			},
		},
	}

	json.NewEncoder(w).Encode(jwks)
}

func (s *TokenService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &s.privateKey.PublicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *TokenService) GetAgentIDFromClaims(claims jwt.MapClaims) (string, bool) {
	agentID, ok := claims["agent_id"].(string)
	return agentID, ok
}

func (s *TokenService) GetClientIDFromClaims(claims jwt.MapClaims) (string, bool) {
	clientID, ok := claims["sub"].(string)
	return clientID, ok
}

func (s *TokenService) GetPublicKey() *rsa.PublicKey {
	return &s.privateKey.PublicKey
}

func parseScope(scope string) []string {
	if scope == "" {
		return nil
	}
	var scopes []string
	for _, s := range splitScope(scope) {
		if s := trimSpace(s); s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}

func filterScopes(allowed []string, requested []string) []string {
	allowedSet := make(map[string]bool)
	for _, s := range allowed {
		allowedSet[s] = true
	}

	var result []string
	for _, s := range requested {
		if allowedSet[s] {
			result = append(result, s)
		}
	}
	return result
}

func joinScopes(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	result := scopes[0]
	for i := 1; i < len(scopes); i++ {
		result += " " + scopes[i]
	}
	return result
}

func splitScope(scope string) []string {
	var result []string
	current := ""
	for _, c := range scope {
		if c == ' ' || c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

var jwtSecret = []byte("machineauth-jwt-secret-key-for-development-only")

func GenerateHMACToken(agent *models.Agent, expirySeconds int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":      "machineauth",
		"sub":      agent.ClientID,
		"agent_id": agent.ID.String(),
		"aud":      "machineauth-api",
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(expirySeconds) * time.Second).Unix(),
		"scope":    agent.Scopes,
		"jti":      generateTokenID(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateHMACToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *TokenService) SignWithKey(token *jwt.Token) (string, error) {
	return token.SignedString(s.privateKey)
}

func (s *TokenService) GetSigningMethod() jwt.SigningMethod {
	return jwt.SigningMethodRS256
}

func (s *TokenService) GetKeyID() string {
	return s.keyID
}

func (s *TokenService) GetExpiry() time.Duration {
	return s.cfg.GetTokenExpiry()
}

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func HashSHA256(data []byte) string {
	hash := crypto.SHA256.New()
	hash.Write(data)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func (s *TokenService) GenerateRefreshToken(agentID string) (string, error) {
	refreshTokenID := uuid.New().String()
	refreshHash, err := bcrypt.GenerateFromPassword([]byte(refreshTokenID), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	expiry := time.Now().Add(7 * 24 * time.Hour)
	rt := db.RefreshToken{
		ID:        refreshTokenID,
		AgentID:   agentID,
		TokenHash: string(refreshHash),
		ExpiresAt: expiry,
		CreatedAt: time.Now(),
	}

	if err := s.db.CreateRefreshToken(rt); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshTokenID, nil
}

func (s *TokenService) ValidateRefreshToken(tokenString string) (*models.Agent, error) {
	rt, err := s.db.GetRefreshToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found")
	}

	if rt.RevokedAt != nil {
		return nil, fmt.Errorf("refresh token revoked")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	agent, err := s.db.GetAgentByID(rt.AgentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	if !agent.IsActive {
		return nil, fmt.Errorf("agent is inactive")
	}

	return &models.Agent{
		ID:        uuid.MustParse(agent.ID),
		Name:      agent.Name,
		ClientID:  agent.ClientID,
		Scopes:    agent.Scopes,
		IsActive:  agent.IsActive,
		CreatedAt: agent.CreatedAt,
		UpdatedAt: agent.UpdatedAt,
	}, nil
}

func (s *TokenService) RevokeRefreshToken(tokenString string) error {
	err := s.db.RevokeRefreshToken(tokenString)
	if err != nil {
		return err
	}
	return s.db.IncrementTokensRefreshed()
}

func (s *TokenService) RecordTokenRefresh() error {
	return s.db.IncrementTokensRefreshed()
}

func (s *TokenService) IntrospectToken(tokenString string) (*models.IntrospectResponse, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &s.privateKey.PublicKey, nil
	})

	if err != nil || !token.Valid {
		return &models.IntrospectResponse{Active: false}, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &models.IntrospectResponse{Active: false}, nil
	}

	jti, _ := claims["jti"].(string)
	if jti != "" && s.db.IsTokenRevoked(jti) {
		return &models.IntrospectResponse{Active: false, Revoked: true, Reason: "revoked"}, nil
	}

	exp, _ := claims["exp"].(float64)
	if time.Now().Unix() > int64(exp) {
		return &models.IntrospectResponse{Active: false, Reason: "expired"}, nil
	}

	scope, _ := claims["scope"].([]interface{})
	var scopeStr string
	if len(scope) > 0 {
		scopeStr = joinScopes(toStringSlice(scope))
	}

	sub, _ := claims["sub"].(string)
	iat, _ := claims["iat"].(float64)

	return &models.IntrospectResponse{
		Active:    true,
		Scope:     scopeStr,
		ClientID:  sub,
		Exp:       int64(exp),
		Iat:       int64(iat),
		TokenType: "Bearer",
	}, nil
}

func (s *TokenService) RevokeAccessToken(jti string) error {
	expiry := time.Now().Add(24 * time.Hour)
	rt := db.RevokedToken{
		JTI:     jti,
		Expires: expiry,
	}
	if err := s.db.AddRevokedToken(rt); err != nil {
		return err
	}
	return s.db.IncrementTokensRevoked()
}

func (s *TokenService) GetMetrics() (int64, int64) {
	metrics := s.db.GetMetrics()
	return metrics.TokensRefreshed, metrics.TokensRevoked
}

func toStringSlice(ifaces []interface{}) []string {
	result := make([]string, len(ifaces))
	for i, v := range ifaces {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}
