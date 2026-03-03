package services

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"machineauth/internal/config"
	"machineauth/internal/db"
	"machineauth/internal/models"
)

// AdminService manages admin users and their JWT sessions.
type AdminService struct {
	db         *db.DB
	cfg        *config.Config
	privateKey *rsa.PrivateKey
	keyID      string
}

// NewAdminService creates an AdminService. It re-uses the same RSA key as the
// token service for signing admin JWTs (different issuer + audience).
func NewAdminService(cfg *config.Config, database *db.DB, privateKey *rsa.PrivateKey, keyID string) *AdminService {
	return &AdminService{
		db:         database,
		cfg:        cfg,
		privateKey: privateKey,
		keyID:      keyID,
	}
}

const (
	adminTokenExpiry   = 8 * time.Hour
	adminTokenIssuer   = "machineauth-admin"
	adminTokenAudience = "machineauth-admin-api"
)

// EnsureDefaultAdmin creates the default admin user from config if no admin
// users exist yet. This allows backwards compatibility with the old
// plaintext admin login.
func (s *AdminService) EnsureDefaultAdmin() error {
	users, err := s.db.ListAdminUsers()
	if err != nil {
		return fmt.Errorf("failed to list admin users: %w", err)
	}
	if len(users) > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	now := time.Now()
	user := db.AdminUser{
		ID:           uuid.New().String(),
		Email:        s.cfg.AdminEmail,
		PasswordHash: string(hash),
		Role:         "owner",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.db.CreateAdminUser(user)
}

// Authenticate validates email/password and returns an admin JWT.
func (s *AdminService) Authenticate(email, password string) (*models.AdminTokenResponse, error) {
	user, err := s.db.GetAdminUserByEmail(email)
	if err != nil {
		// Fallback: check legacy config-based credentials.
		if email == s.cfg.AdminEmail && password == s.cfg.AdminPassword {
			return s.issueToken(s.cfg.AdminEmail, "owner", "legacy")
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return s.issueToken(user.Email, user.Role, user.ID)
}

func (s *AdminService) issueToken(email, role, userID string) (*models.AdminTokenResponse, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":      adminTokenIssuer,
		"sub":      email,
		"aud":      adminTokenAudience,
		"admin_id": userID,
		"role":     role,
		"iat":      now.Unix(),
		"exp":      now.Add(adminTokenExpiry).Unix(),
		"jti":      uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID

	tokenStr, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign admin token: %w", err)
	}

	return &models.AdminTokenResponse{
		Success:     true,
		AccessToken: tokenStr,
		ExpiresIn:   int(adminTokenExpiry.Seconds()),
		Role:        role,
	}, nil
}

// ValidateAdminToken parses and validates an admin JWT. Returns the claims.
func (s *AdminService) ValidateAdminToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return &s.privateKey.PublicKey, nil
	},
		jwt.WithIssuer(adminTokenIssuer),
		jwt.WithAudience(adminTokenAudience),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid admin token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid admin token claims")
	}

	return claims, nil
}
