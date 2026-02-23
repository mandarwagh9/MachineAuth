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

	"agentauth/internal/config"
	"agentauth/internal/models"
)

type TokenService struct {
	cfg             *config.Config
	privateKey      *rsa.PrivateKey
	keyID           string
	mu              sync.RWMutex
	tokenExpirySecs int
}

func NewTokenService(cfg *config.Config) (*TokenService, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &TokenService{
		cfg:             cfg,
		privateKey:      privateKey,
		keyID:           cfg.JWTKeyID,
		tokenExpirySecs: cfg.JWTExpirySeconds,
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
		"iss":   "https://auth.example.com",
		"sub":   agent.ClientID,
		"aud":   "agentauth-api",
		"iat":   now.Unix(),
		"exp":   now.Add(s.cfg.GetTokenExpiry()).Unix(),
		"scope": scopes,
		"jti":   generateTokenID(),
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

var jwtSecret = []byte("agentauth-jwt-secret-key-for-development-only")

func GenerateHMACToken(agent *models.Agent, expirySeconds int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   "agentauth",
		"sub":   agent.ClientID,
		"aud":   "agentauth-api",
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(expirySeconds) * time.Second).Unix(),
		"scope": agent.Scopes,
		"jti":   generateTokenID(),
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
