package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"machineauth/internal/db"
	"machineauth/internal/models"
)

type OrganizationService struct {
	db *db.DB
}

func NewOrganizationService(database *db.DB) *OrganizationService {
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

	return toModelOrganization(org), nil
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
	db *db.DB
}

func NewTeamService(database *db.DB) *TeamService {
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
