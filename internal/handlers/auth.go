package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"machineauth/internal/config"
	"machineauth/internal/models"
	"machineauth/internal/services"
)

type AuthHandler struct {
	agentService *services.AgentService
	tokenService *services.TokenService
	cfg          *config.Config
}

func NewAuthHandler(agentService *services.AgentService, tokenService *services.TokenService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		agentService: agentService,
		tokenService: tokenService,
		cfg:          cfg,
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

	if tokenReq.GrantType != "client_credentials" && tokenReq.GrantType != "refresh_token" {
		h.writeError(w, "unsupported_grant_type", "grant_type must be client_credentials or refresh_token")
		return
	}

	var agent *models.Agent

	if tokenReq.GrantType == "client_credentials" {
		if tokenReq.ClientID == "" || tokenReq.ClientSecret == "" {
			h.writeError(w, "invalid_request", "client_id and client_secret are required")
			return
		}

		agent, err = h.agentService.ValidateCredentials(tokenReq.ClientID, tokenReq.ClientSecret)
		if err != nil {
			log.Printf("failed to validate credentials for client_id=%s: %v", tokenReq.ClientID, err)
			h.writeError(w, "invalid_client", "invalid client credentials")
			return
		}
	} else if tokenReq.GrantType == "refresh_token" {
		agent, err = h.tokenService.ValidateRefreshToken(tokenReq.RefreshToken)
		if err != nil {
			log.Printf("failed to validate refresh token: %v", err)
			h.writeError(w, "invalid_grant", "invalid or expired refresh token")
			return
		}

		h.agentService.RecordTokenRefresh(agent.ID)
	}

	tokenResp, err := h.tokenService.GenerateToken(agent, tokenReq.Scope)
	if err != nil {
		log.Printf("failed to generate token for client_id=%s: %v", agent.ClientID, err)
		h.writeError(w, "server_error", "failed to generate token")
		return
	}

	h.agentService.RecordTokenIssuance(agent.ID)

	if tokenReq.GrantType == "client_credentials" {
		refreshToken, err := h.tokenService.GenerateRefreshToken(agent.ID.String())
		if err != nil {
			log.Printf("failed to generate refresh token: %v", err)
		} else {
			tokenResp.RefreshToken = refreshToken
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResp)
}

func (h *AuthHandler) Introspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var token string
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.writeError(w, "invalid_request", "failed to read request body")
			return
		}
		defer r.Body.Close()

		var req struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			h.writeError(w, "invalid_request", "invalid JSON")
			return
		}
		token = req.Token
	} else {
		if err := r.ParseForm(); err != nil {
			h.writeError(w, "invalid_request", "failed to parse form")
			return
		}
		token = r.Form.Get("token")
	}

	if token == "" {
		h.writeError(w, "invalid_request", "token is required")
		return
	}

	resp, err := h.tokenService.IntrospectToken(token)
	if err != nil {
		log.Printf("failed to introspect token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.IntrospectResponse{Active: false})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var token, tokenTypeHint string
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.writeError(w, "invalid_request", "failed to read request body")
			return
		}
		defer r.Body.Close()

		var req models.RevokeRequest
		if err := json.Unmarshal(body, &req); err != nil {
			h.writeError(w, "invalid_request", "invalid JSON")
			return
		}
		token = req.Token
		tokenTypeHint = req.TokenTypeHint
	} else {
		if err := r.ParseForm(); err != nil {
			h.writeError(w, "invalid_request", "failed to parse form")
			return
		}
		token = r.Form.Get("token")
		tokenTypeHint = r.Form.Get("token_type_hint")
	}

	if token == "" {
		h.writeError(w, "invalid_request", "token is required")
		return
	}

	if tokenTypeHint == "refresh_token" || tokenTypeHint == "" {
		err := h.tokenService.RevokeRefreshToken(token)
		if err != nil {
			log.Printf("failed to revoke refresh token: %v", err)
		}
	}

	if tokenTypeHint == "access_token" || tokenTypeHint == "" {
		parsedToken, err := h.tokenService.ValidateToken(token)
		if err == nil {
			if jti, ok := parsedToken["jti"].(string); ok {
				h.tokenService.RevokeAccessToken(jti)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "revoked"})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var refreshToken, clientID, clientSecret string
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.writeError(w, "invalid_request", "failed to read request body")
			return
		}
		defer r.Body.Close()

		var req models.RefreshRequest
		if err := json.Unmarshal(body, &req); err != nil {
			h.writeError(w, "invalid_request", "invalid JSON")
			return
		}
		refreshToken = req.RefreshToken
		clientID = req.ClientID
		clientSecret = req.ClientSecret
	} else {
		if err := r.ParseForm(); err != nil {
			h.writeError(w, "invalid_request", "failed to parse form")
			return
		}
		refreshToken = r.Form.Get("refresh_token")
		clientID = r.Form.Get("client_id")
		clientSecret = r.Form.Get("client_secret")
	}

	if refreshToken == "" {
		h.writeError(w, "invalid_request", "refresh_token is required")
		return
	}

	if clientID != "" && clientSecret != "" {
		_, err := h.agentService.ValidateCredentials(clientID, clientSecret)
		if err != nil {
			h.writeError(w, "invalid_client", "invalid client credentials")
			return
		}
	}

	agent, err := h.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Printf("failed to validate refresh token: %v", err)
		h.writeError(w, "invalid_grant", "invalid or expired refresh token")
		return
	}

	h.agentService.RecordTokenRefresh(agent.ID)
	h.tokenService.RecordTokenRefresh()

	tokenResp, err := h.tokenService.GenerateToken(agent, "")
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		h.writeError(w, "server_error", "failed to generate token")
		return
	}

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

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func (h *AuthHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeAdminError(w, "invalid_request", "invalid request body")
		return
	}

	if req.Username == h.cfg.AdminEmail && req.Password == h.cfg.AdminPassword {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AdminLoginResponse{
			Success: true,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(AdminLoginResponse{
		Success: false,
		Message: "invalid credentials",
	})
}

func (h *AuthHandler) writeAdminError(w http.ResponseWriter, error, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   error,
		"message": description,
	})
}
