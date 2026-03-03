package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"machineauth/internal/config"
	"machineauth/internal/db"
	"machineauth/internal/models"
)

// RSA key and token configuration constants.
const (
	rsaKeyBits         = 2048
	jwtAudience        = "machineauth-api"
	refreshTokenExpiry = 7 * 24 * time.Hour
	revokedTokenExpiry = 24 * time.Hour
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
	privateKey, err := loadOrGenerateKey(cfg.JWTKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load or generate RSA key: %w", err)
	}

	return &TokenService{
		cfg:             cfg,
		privateKey:      privateKey,
		keyID:           cfg.JWTKeyID,
		tokenExpirySecs: cfg.JWTExpirySeconds,
		db:              database,
	}, nil
}

func loadOrGenerateKey(keyPath string) (*rsa.PrivateKey, error) {
	privateKeyPath := filepath.Join(keyPath, "jwt-private.pem")
	publicKeyPath := filepath.Join(keyPath, "jwt-public.pem")

	if _, err := os.Stat(privateKeyPath); err == nil {
		return loadPrivateKey(privateKeyPath)
	}

	if err := os.MkdirAll(keyPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	if err := savePrivateKey(privateKeyPath, key); err != nil {
		return nil, fmt.Errorf("failed to save private key: %w", err)
	}

	if err := savePublicKey(publicKeyPath, &key.PublicKey); err != nil {
		return nil, fmt.Errorf("failed to save public key: %w", err)
	}

	return key, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key format")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func savePrivateKey(path string, key *rsa.PrivateKey) error {
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return os.WriteFile(path, pem.EncodeToMemory(block), 0600)
}

func savePublicKey(path string, key *rsa.PublicKey) error {
	data, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}
	block := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: data}
	return os.WriteFile(path, pem.EncodeToMemory(block), 0644)
}

func (s *TokenService) GenerateToken(agent *models.Agent, requestedScope string) (*models.TokenResponse, error) {
	scopes := agent.Scopes
	if requestedScope != "" {
		requestedScopes := parseScope(requestedScope)
		scopes = filterScopes(scopes, requestedScopes)
	}

	now := time.Now()
	var teamIDStr string
	if agent.TeamID != nil {
		teamIDStr = agent.TeamID.String()
	}
	claims := jwt.MapClaims{
		"iss":      s.cfg.JWTIssuer,
		"sub":      agent.ClientID,
		"agent_id": agent.ID.String(),
		"org_id":   agent.OrganizationID,
		"team_id":  teamIDStr,
		"aud":      jwtAudience,
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

	TokensIssued.Inc()

	return &models.TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   s.tokenExpirySecs,
		Scope:       strings.Join(scopes, " "),
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
	// OAuth scopes may be space-separated (RFC 6749) or comma-separated
	// (common client convenience). Handle both.
	return strings.FieldsFunc(scope, func(r rune) bool { return r == ' ' || r == ',' })
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

func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func GenerateHMACToken(agent *models.Agent, expirySeconds int, hmacSecret []byte) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":      "machineauth",
		"sub":      agent.ClientID,
		"agent_id": agent.ID.String(),
		"aud":      jwtAudience,
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(expirySeconds) * time.Second).Unix(),
		"scope":    agent.Scopes,
		"jti":      generateTokenID(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(hmacSecret)
}

func ValidateHMACToken(tokenString string, hmacSecret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSecret, nil
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

	expiry := time.Now().Add(refreshTokenExpiry)
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
	TokensRevoked.Inc()
	return s.db.IncrementTokensRevoked()
}

func (s *TokenService) RecordTokenRefresh() error {
	TokensRefreshed.Inc()
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
			scopeStr = strings.Join(toStringSlice(scope), " ")
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
	expiry := time.Now().Add(revokedTokenExpiry)
	rt := db.RevokedToken{
		JTI:     jti,
		Expires: expiry,
	}
	if err := s.db.AddRevokedToken(rt); err != nil {
		return err
	}
	TokensRevoked.Inc()
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
