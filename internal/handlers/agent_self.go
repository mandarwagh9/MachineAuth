package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"machineauth/internal/middleware"
	"machineauth/internal/models"
	"machineauth/internal/services"
)

type AgentSelfHandler struct {
	agentService *services.AgentService
}

func NewAgentSelfHandler(agentService *services.AgentService) *AgentSelfHandler {
	return &AgentSelfHandler{agentService: agentService}
}

func (h *AgentSelfHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	agentID, ok := middleware.GetAgentIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	agent, err := h.agentService.GetByID(agentID)
	if err != nil {
		log.Printf("failed to get agent %s: %v", agentID, err)
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AgentResponse{Agent: *agent})
}

func (h *AgentSelfHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	agentID, ok := middleware.GetAgentIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	usage, err := h.agentService.GetUsage(agentID)
	if err != nil {
		log.Printf("failed to get usage for agent %s: %v", agentID, err)
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

func (h *AgentSelfHandler) RotateCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	newSecret, err := h.agentService.RotateCredentials(agentID)
	if err != nil {
		log.Printf("failed to rotate credentials for agent %s: %v", agentID, err)
		http.Error(w, "failed to rotate credentials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"client_secret": newSecret,
	})
}

func (h *AgentSelfHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.agentService.Deactivate(agentID)
	if err != nil {
		log.Printf("failed to deactivate agent %s: %v", agentID, err)
		http.Error(w, "failed to deactivate agent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "agent deactivated successfully",
	})
}

func (h *AgentSelfHandler) Reactivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.agentService.Reactivate(agentID)
	if err != nil {
		log.Printf("failed to reactivate agent %s: %v", agentID, err)
		http.Error(w, "failed to reactivate agent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "agent reactivated successfully",
	})
}

func (h *AgentSelfHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.agentService.Delete(agentID)
	if err != nil {
		log.Printf("failed to delete agent %s: %v", agentID, err)
		http.Error(w, "failed to delete agent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
