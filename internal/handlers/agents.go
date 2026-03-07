package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"machineauth/internal/middleware"
	"machineauth/internal/models"
	"machineauth/internal/services"
)

type AgentsHandler struct {
	agentService *services.AgentService
	auditService *services.AuditService
}

func NewAgentsHandler(agentService *services.AgentService, auditService *services.AuditService) *AgentsHandler {
	return &AgentsHandler{
		agentService: agentService,
		auditService: auditService,
	}
}

func (h *AgentsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents, err := h.agentService.List()
	if err != nil {
		log.Printf("failed to list agents: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if agents == nil {
		agents = []models.Agent{}
	}

	writeJSON(w, models.AgentsListResponse{Agents: agents})
}

// ListPaginated returns agents with pagination, search, and filtering.
func (h *AgentsHandler) ListPaginated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	params := models.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: q.Get("q"),
		Status: q.Get("status"),
		OrgID:  q.Get("org_id"),
		Sort:   q.Get("sort"),
	}

	// If org_id is in context (from auth middleware), use it to scope results.
	if ctxOrgID, ok := middleware.GetOrgIDFromContext(r.Context()); ok && ctxOrgID != "" {
		params.OrgID = ctxOrgID
	}

	agents, total, err := h.agentService.ListPaginated(params)
	if err != nil {
		log.Printf("failed to list agents: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	writeJSON(w, models.AgentsListResponse{
		Agents: agents,
		Pagination: &models.Pagination{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

func (h *AgentsHandler) CreateInOrganization(w http.ResponseWriter, r *http.Request, orgID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req models.CreateAgentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	req.OrganizationID = orgID

	agent, err := h.agentService.Create(req)
	if err != nil {
		log.Printf("failed to create agent: %v", err)
		http.Error(w, "failed to create agent", http.StatusInternalServerError)
		return
	}

	writeJSONStatus(w, http.StatusCreated, agent)
}

func (h *AgentsHandler) HandleAgent(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/agents/") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	idStr := strings.TrimPrefix(path, "/api/agents/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid agent ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getAgent(w, r, id)
	case http.MethodPut, http.MethodPatch:
		h.updateAgent(w, r, id)
	case http.MethodDelete:
		h.deleteAgent(w, r, id)
	case http.MethodPost:
		h.rotateAgent(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AgentsHandler) getAgent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	agent, err := h.agentService.GetByID(id)
	if err != nil {
		log.Printf("failed to get agent %s: %v", id, err)
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	writeJSON(w, models.AgentResponse{Agent: *agent})
}

func (h *AgentsHandler) updateAgent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req models.UpdateAgentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	agent, err := h.agentService.Update(id, req)
	if err != nil {
		log.Printf("failed to update agent %s: %v", id, err)
		http.Error(w, "failed to update agent", http.StatusInternalServerError)
		return
	}

	h.auditService.LogAgentUpdated(id, r.RemoteAddr, r.UserAgent())

	writeJSON(w, models.AgentResponse{Agent: *agent})
}

func (h *AgentsHandler) deleteAgent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.agentService.Delete(id); err != nil {
		log.Printf("failed to delete agent %s: %v", id, err)
		http.Error(w, "failed to delete agent", http.StatusInternalServerError)
		return
	}

	h.auditService.LogAgentDeleted(id, r.RemoteAddr, r.UserAgent())

	w.WriteHeader(http.StatusNoContent)
}

func (h *AgentsHandler) rotateAgent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) > 0 {
		var req struct {
			Action string `json:"action"`
		}
		if err := json.Unmarshal(body, &req); err != nil || req.Action != "rotate" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	}

	newSecret, err := h.agentService.RotateCredentials(id)
	if err != nil {
		log.Printf("failed to rotate credentials for agent %s: %v", id, err)
		http.Error(w, "failed to rotate credentials", http.StatusInternalServerError)
		return
	}

	h.auditService.LogCredentialsRotated(id, r.RemoteAddr, r.UserAgent())

	writeJSON(w, map[string]string{
		"client_secret": newSecret,
	})
}

func (h *AgentsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req models.CreateAgentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Inject org_id from context if not explicitly set in the request.
	if req.OrganizationID == "" {
		if ctxOrgID, ok := middleware.GetOrgIDFromContext(r.Context()); ok && ctxOrgID != "" {
			req.OrganizationID = ctxOrgID
		}
	}

	resp, err := h.agentService.Create(req)
	if err != nil {
		log.Printf("failed to create agent: %v", err)
		http.Error(w, "failed to create agent", http.StatusInternalServerError)
		return
	}

	h.auditService.LogAgentCreated(&resp.Agent, r.RemoteAddr, r.UserAgent())

	writeJSONStatus(w, http.StatusCreated, resp)
}
