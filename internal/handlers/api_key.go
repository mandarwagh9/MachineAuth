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

type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID := extractOrgID(r)
	if orgID == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	keys, err := h.apiKeyService.ListByOrganization(orgID)
	if err != nil {
		log.Printf("failed to list API keys: %v", err)
		http.Error(w, "failed to list API keys", http.StatusInternalServerError)
		return
	}

	if keys == nil {
		keys = []models.APIKey{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIKeysResponse{APIKeys: keys})
}

func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID := extractOrgID(r)
	if orgID == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req models.CreateAPIKeyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	key, err := h.apiKeyService.Create(orgID, req)
	if err != nil {
		log.Printf("failed to create API key: %v", err)
		http.Error(w, "failed to create API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(key)
}

func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/organizations/")
	id = strings.TrimSuffix(id, "/api-keys")
	id = strings.TrimPrefix(id, "/")

	if id == "" {
		http.Error(w, "API key ID required", http.StatusBadRequest)
		return
	}

	if err := h.apiKeyService.Revoke(id); err != nil {
		log.Printf("failed to revoke API key: %v", err)
		http.Error(w, "failed to revoke API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractOrgID(r *http.Request) string {
	path := r.URL.Path
	if strings.Contains(path, "/api-keys") {
		parts := strings.Split(path, "/api-keys")
		if len(parts) > 0 {
			orgPart := strings.TrimSuffix(parts[0], "/")
			orgID := strings.TrimPrefix(orgPart, "/api/organizations/")
			return orgID
		}
	}
	return ""
}
