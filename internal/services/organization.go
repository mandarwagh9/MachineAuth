package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"machineauth/internal/db"
	"machineauth/internal/models"
)

type OrganizationService struct {
	db db.Database
}

func NewOrganizationService(database db.Database) *OrganizationService {
	return &OrganizationService{db: database}
}

func (s *OrganizationService) Create(req models.CreateOrganizationRequest) (*models.Organization, error) {
	id := uuid.New()
	now := time.Now()

	org := db.Organization{
		ID:             id.String(),
		Name:           req.Name,
		Slug:           req.Slug,
		OwnerEmail:     req.OwnerEmail,
		JWTIssuer:      "https://auth.agentauth.io",
		JWTExpirySecs:  3600,
		AllowedOrigins: "*",
		Plan:           "free",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.CreateOrganization(org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Auto-generate a per-org RSA signing key for OIDC.
	if err := s.generateOrgSigningKey(org.ID, org.Slug); err != nil {
		log.Printf("warning: failed to generate signing key for org %s: %v", org.ID, err)
	}

	return toModelOrganization(org), nil
}

// generateOrgSigningKey creates a 2048-bit RSA key pair for the given org
// and stores it in the database. This enables per-org OIDC with isolated signing keys.
func (s *OrganizationService) generateOrgSigningKey(orgID, orgSlug string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key to PEM.
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privPEM := string(pem.EncodeToMemory(privBlock))

	// Encode public key to PEM.
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}
	pubBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	}
	pubPEM := string(pem.EncodeToMemory(pubBlock))

	now := time.Now()
	key := db.OrgSigningKey{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		KeyID:          orgSlug + "-key-1",
		PublicKeyPEM:   pubPEM,
		PrivateKeyPEM:  privPEM,
		Algorithm:      "RS256",
		IsActive:       true,
		CreatedAt:      now,
	}

	return s.db.CreateOrgSigningKey(key)
}

func (s *OrganizationService) GetByID(id string) (*models.Organization, error) {
	org, err := s.db.GetOrganization(id)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return toModelOrganization(*org), nil
}

func (s *OrganizationService) GetBySlug(slug string) (*models.Organization, error) {
	org, err := s.db.GetOrganizationBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return toModelOrganization(*org), nil
}

func (s *OrganizationService) List() ([]models.Organization, error) {
	orgs, err := s.db.ListOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	result := make([]models.Organization, len(orgs))
	for i, o := range orgs {
		result[i] = *toModelOrganization(o)
	}
	return result, nil
}

func (s *OrganizationService) Update(id string, req models.UpdateOrganizationRequest) (*models.Organization, error) {
	err := s.db.UpdateOrganization(id, func(org *db.Organization) error {
		if req.Name != "" {
			org.Name = req.Name
		}
		if req.JWTIssuer != "" {
			org.JWTIssuer = req.JWTIssuer
		}
		if req.JWTExpirySecs > 0 {
			org.JWTExpirySecs = req.JWTExpirySecs
		}
		if req.AllowedOrigins != "" {
			org.AllowedOrigins = req.AllowedOrigins
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return s.GetByID(id)
}

func (s *OrganizationService) Delete(id string) error {
	return s.db.DeleteOrganization(id)
}

func toModelOrganization(o db.Organization) *models.Organization {
	return &models.Organization{
		ID:             o.ID,
		Name:           o.Name,
		Slug:           o.Slug,
		OwnerEmail:     o.OwnerEmail,
		JWTIssuer:      o.JWTIssuer,
		JWTExpirySecs:  o.JWTExpirySecs,
		AllowedOrigins: o.AllowedOrigins,
		Plan:           o.Plan,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

type TeamService struct {
	db db.Database
}

func NewTeamService(database db.Database) *TeamService {
	return &TeamService{db: database}
}

func (s *TeamService) Create(orgID string, req models.CreateTeamRequest) (*models.Team, error) {
	id := uuid.New()
	now := time.Now()

	team := db.Team{
		ID:             id.String(),
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.CreateTeam(team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return toModelTeam(team), nil
}

func (s *TeamService) GetByID(id string) (*models.Team, error) {
	team, err := s.db.GetTeam(id)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}
	return toModelTeam(*team), nil
}

func (s *TeamService) ListByOrganization(orgID string) ([]models.Team, error) {
	teams, err := s.db.ListTeamsByOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	result := make([]models.Team, len(teams))
	for i, t := range teams {
		result[i] = *toModelTeam(t)
	}
	return result, nil
}

func (s *TeamService) Update(id string, req models.UpdateTeamRequest) (*models.Team, error) {
	err := s.db.UpdateTeam(id, func(team *db.Team) error {
		if req.Name != "" {
			team.Name = req.Name
		}
		if req.Description != "" {
			team.Description = req.Description
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return s.GetByID(id)
}

func (s *TeamService) Delete(id string) error {
	return s.db.DeleteTeam(id)
}

func toModelTeam(t db.Team) *models.Team {
	return &models.Team{
		ID:             t.ID,
		OrganizationID: t.OrganizationID,
		Name:           t.Name,
		Description:    t.Description,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
