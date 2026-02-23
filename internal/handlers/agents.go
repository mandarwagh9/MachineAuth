package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AgentsListResponse{Agents: agents})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AgentResponse{Agent: *agent})
}

func (h *AgentsHandler) deleteAgent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.agentService.Delete(id); err != nil {
		log.Printf("failed to delete agent %s: %v", id, err)
		http.Error(w, "failed to delete agent", http.StatusInternalServerError)
		return
	}

	h.auditService.LogAgentDeleted(id, r.RemoteAddr, r.UserAgent())

	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
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

	resp, err := h.agentService.Create(req)
	if err != nil {
		log.Printf("failed to create agent: %v", err)
		http.Error(w, "failed to create agent", http.StatusInternalServerError)
		return
	}

	h.auditService.LogAgentCreated(&resp.Agent, r.RemoteAddr, r.UserAgent())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
