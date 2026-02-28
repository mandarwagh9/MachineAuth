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

type AuthHandler struct {
	agentService *services.AgentService
	tokenService *services.TokenService
}

func NewAuthHandler(agentService *services.AgentService, tokenService *services.TokenService) *AuthHandler {
	return &AuthHandler{
		agentService: agentService,
		tokenService: tokenService,
	}
}

func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/x-www-form-urlencoded") && !strings.Contains(contentType, "application/json") {
		h.writeError(w, "invalid_request", "content-type must be application/x-www-form-urlencoded or application/json")
		return
	}

	var tokenReq models.TokenRequest
	var err error

	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.writeError(w, "invalid_request", "failed to read request body")
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &tokenReq); err != nil {
			h.writeError(w, "invalid_request", "invalid JSON")
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			h.writeError(w, "invalid_request", "failed to parse form")
			return
		}

		tokenReq.GrantType = r.Form.Get("grant_type")
		tokenReq.ClientID = r.Form.Get("client_id")
		tokenReq.ClientSecret = r.Form.Get("client_secret")
		tokenReq.Scope = r.Form.Get("scope")
	}

	if tokenReq.GrantType != "client_credentials" {
		h.writeError(w, "unsupported_grant_type", "grant_type must be client_credentials")
		return
	}

	if tokenReq.ClientID == "" || tokenReq.ClientSecret == "" {
		h.writeError(w, "invalid_request", "client_id and client_secret are required")
		return
	}

	agent, err := h.agentService.ValidateCredentials(tokenReq.ClientID, tokenReq.ClientSecret)
	if err != nil {
		log.Printf("failed to validate credentials for client_id=%s: %v", tokenReq.ClientID, err)
		h.writeError(w, "invalid_client", "invalid client credentials")
		return
	}

	tokenResp, err := h.tokenService.GenerateToken(agent, tokenReq.Scope)
	if err != nil {
		log.Printf("failed to generate token for client_id=%s: %v", tokenReq.ClientID, err)
		h.writeError(w, "server_error", "failed to generate token")
		return
	}

	h.agentService.RecordTokenIssuance(agent.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResp)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, error, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:            error,
		ErrorDescription: description,
	})
}
