package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"machineauth/internal/models"
	"machineauth/internal/services"
)

type OrganizationHandler struct {
	orgService  *services.OrganizationService
	teamService *services.TeamService
}

func NewOrganizationHandler(orgService *services.OrganizationService, teamService *services.TeamService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService:  orgService,
		teamService: teamService,
	}
}

func (h *OrganizationHandler) writeError(w http.ResponseWriter, errCode, description string) {
	writeJSONError(w, http.StatusBadRequest, errCode, description)
}

func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.orgService.List()
	if err != nil {
		log.Printf("failed to list organizations: %v", err)
		h.writeError(w, "server_error", "failed to list organizations")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.OrganizationsResponse{Organizations: orgs})
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, "invalid_request", "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req models.CreateOrganizationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, "invalid_request", "invalid JSON")
		return
	}

	if req.Name == "" || req.Slug == "" {
		h.writeError(w, "invalid_request", "name and slug are required")
		return
	}

	org, err := h.orgService.Create(req)
	if err != nil {
		log.Printf("failed to create organization: %v", err)
		h.writeError(w, "server_error", "failed to create organization")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id := extractID(r)
	if id == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	org, err := h.orgService.GetByID(id)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractID(r)
	if id == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, "invalid_request", "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req models.UpdateOrganizationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, "invalid_request", "invalid JSON")
		return
	}

	org, err := h.orgService.Update(id, req)
	if err != nil {
		log.Printf("failed to update organization: %v", err)
		h.writeError(w, "server_error", "failed to update organization")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractID(r)
	if id == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	if err := h.orgService.Delete(id); err != nil {
		log.Printf("failed to delete organization: %v", err)
		h.writeError(w, "server_error", "failed to delete organization")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *OrganizationHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	orgID := extractID(r)
	if orgID == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	teams, err := h.teamService.ListByOrganization(orgID)
	if err != nil {
		log.Printf("failed to list teams: %v", err)
		h.writeError(w, "server_error", "failed to list teams")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TeamsResponse{Teams: teams})
}

func (h *OrganizationHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := extractID(r)
	if orgID == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, "invalid_request", "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req models.CreateTeamRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, "invalid_request", "invalid JSON")
		return
	}

	if req.Name == "" {
		h.writeError(w, "invalid_request", "name is required")
		return
	}

	team, err := h.teamService.Create(orgID, req)
	if err != nil {
		log.Printf("failed to create team: %v", err)
		h.writeError(w, "server_error", "failed to create team")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func extractID(r *http.Request) string {
	path := r.URL.Path
	if strings.HasSuffix(path, "/teams") {
		orgID := strings.TrimPrefix(path, "/api/organizations/")
		orgID = strings.TrimSuffix(orgID, "/teams")
		return orgID
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return ""
}
