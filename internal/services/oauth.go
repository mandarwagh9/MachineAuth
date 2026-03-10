package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"machineauth/internal/config"
	"machineauth/internal/db"
)

const (
	authorizationCodeExpiry = 10 * time.Minute
)

type OAuthService struct {
	cfg          *config.Config
	db           db.Database
	tokenService *TokenService
}

func NewOAuthService(cfg *config.Config, database db.Database, tokenService *TokenService) *OAuthService {
	return &OAuthService{
		cfg:          cfg,
		db:           database,
		tokenService: tokenService,
	}
}

func generateSecureCode(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func (s *OAuthService) CreateAuthorizationCode(
	clientID, userID, orgID, redirectURI, scope, codeChallenge, codeChallengeMethod string,
) (*db.AuthorizationCode, error) {
	code := generateSecureCode(32)

	authCode := db.AuthorizationCode{
		ID:                  uuid.New(),
		ClientID:            clientID,
		UserID:              userID,
		OrganizationID:      orgID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Code:                code,
		ExpiresAt:           time.Now().Add(authorizationCodeExpiry),
		Used:                false,
		CreatedAt:           time.Now(),
	}

	if err := s.db.CreateAuthorizationCode(authCode); err != nil {
		return nil, fmt.Errorf("failed to create authorization code: %w", err)
	}

	return &authCode, nil
}

func (s *OAuthService) ValidateAuthorizationCode(code, redirectURI string) (*db.AuthorizationCode, error) {
	storedCode, err := s.db.GetAuthorizationCode(code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}

	if storedCode.Used {
		return nil, fmt.Errorf("authorization code already used")
	}

	if time.Now().After(storedCode.ExpiresAt) {
		return nil, fmt.Errorf("authorization code expired")
	}

	if storedCode.RedirectURI != redirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}

	return storedCode, nil
}

func (s *OAuthService) ConsumeAuthorizationCode(id uuid.UUID) error {
	return s.db.UseAuthorizationCode(id.String())
}

func (s *OAuthService) ValidateCodeVerifier(codeVerifier, codeChallenge, codeChallengeMethod string) error {
	if codeChallenge == "" {
		return nil
	}

	if codeChallengeMethod != "S256" {
		return fmt.Errorf("only S256 code_challenge_method is supported")
	}

	verifierHash := sha256.Sum256([]byte(codeVerifier))
	encodedVerifier := base64.RawURLEncoding.EncodeToString(verifierHash[:])

	if encodedVerifier != codeChallenge {
		return fmt.Errorf("invalid code_verifier")
	}

	return nil
}

type IDTokenClaims struct {
	Subject        string `json:"sub"`
	Audience       string `json:"aud"`
	IssuedAt       int64  `json:"iat"`
	ExpiresAt      int64  `json:"exp"`
	AuthTime       int64  `json:"auth_time"`
	Name           string `json:"name,omitempty"`
	Email          string `json:"email,omitempty"`
	OrganizationID string `json:"org_id,omitempty"`
}

func (s *OAuthService) GenerateIDToken(userID, email, orgID, clientID string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":    userID,
		"aud":    clientID,
		"iat":    now.Unix(),
		"exp":    now.Add(time.Duration(s.cfg.JWTExpirySeconds) * time.Second).Unix(),
		"iss":    s.cfg.BaseURL,
		"org_id": orgID,
		"name":   email,
		"email":  email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.tokenService.KeyID()

	tokenString, err := token.SignedString(s.tokenService.PrivateKey())
	if err != nil {
		return "", fmt.Errorf("failed to sign ID token: %w", err)
	}

	return tokenString, nil
}
