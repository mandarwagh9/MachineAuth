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
	db         db.Database
	cfg        *config.Config
	privateKey *rsa.PrivateKey
	keyID      string
}

// NewAdminService creates an AdminService. It re-uses the same RSA key as the
// token service for signing admin JWTs (different issuer + audience).
func NewAdminService(cfg *config.Config, database db.Database, privateKey *rsa.PrivateKey, keyID string) *AdminService {
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
			return s.issueToken(s.cfg.AdminEmail, "owner", "legacy", "")
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Find the user's org membership to embed org_id in the JWT.
	orgID := ""
	memberships, _ := s.db.ListOrgMembersByUser(user.ID)
	if len(memberships) > 0 {
		// Use the first org (users can switch orgs later via a separate endpoint).
		orgID = memberships[0].OrganizationID
	}

	return s.issueToken(user.Email, user.Role, user.ID, orgID)
}

func (s *AdminService) issueToken(email, role, userID, orgID string) (*models.AdminTokenResponse, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":      adminTokenIssuer,
		"sub":      email,
		"aud":      adminTokenAudience,
		"admin_id": userID,
		"role":     role,
		"org_id":   orgID,
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

// Signup creates a new admin user + organization + org_member atomically.
// This is the self-service registration flow.
func (s *AdminService) Signup(req models.SignupRequest) (*models.AdminTokenResponse, error) {
	if req.Email == "" || req.Password == "" || req.OrgName == "" || req.OrgSlug == "" {
		return nil, fmt.Errorf("email, password, org_name, and org_slug are required")
	}

	// Check if email already taken.
	if existing, _ := s.db.GetAdminUserByEmail(req.Email); existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Check if slug already taken.
	if existing, _ := s.db.GetOrganizationBySlug(req.OrgSlug); existing != nil {
		return nil, fmt.Errorf("organization slug already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	userID := uuid.New().String()
	orgID := uuid.New().String()

	// Create admin user.
	user := db.AdminUser{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.db.CreateAdminUser(user); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create organization.
	org := db.Organization{
		ID:             orgID,
		Name:           req.OrgName,
		Slug:           req.OrgSlug,
		OwnerEmail:     req.Email,
		JWTIssuer:      "machineauth-" + req.OrgSlug,
		JWTExpirySecs:  3600,
		AllowedOrigins: "",
		Plan:           "free",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.db.CreateOrganization(org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Create org membership (owner role).
	member := db.OrgMember{
		ID:             uuid.New().String(),
		UserID:         userID,
		OrganizationID: orgID,
		Role:           "owner",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.db.CreateOrgMember(member); err != nil {
		return nil, fmt.Errorf("failed to create org membership: %w", err)
	}

	return s.issueToken(req.Email, "admin", userID, orgID)
}

// IssueTokenForOrg issues a new admin JWT scoped to a specific organization.
// Used when an admin switches between orgs.
func (s *AdminService) IssueTokenForOrg(userID, orgID string) (*models.AdminTokenResponse, error) {
	user, err := s.db.GetAdminUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verify the user is a member of this org.
	member, err := s.db.GetOrgMember(userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("not a member of this organization")
	}

	return s.issueToken(user.Email, member.Role, userID, orgID)
}

// GetUserByID returns an admin user by ID.
func (s *AdminService) GetUserByID(id string) (*db.AdminUser, error) {
	return s.db.GetAdminUserByID(id)
}

// ListOrgMemberships returns all organizations a user belongs to.
func (s *AdminService) ListOrgMemberships(userID string) ([]db.OrgMember, error) {
	return s.db.ListOrgMembersByUser(userID)
}
