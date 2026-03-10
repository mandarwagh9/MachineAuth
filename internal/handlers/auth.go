package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"machineauth/internal/config"
	"machineauth/internal/middleware"
	"machineauth/internal/models"
	"machineauth/internal/services"
)

type AuthHandler struct {
	agentService *services.AgentService
	tokenService *services.TokenService
	adminService *services.AdminService
	oauthService *services.OAuthService
	cfg          *config.Config

	// Brute-force tracking: map[key] -> {failures, lockedUntil}
	bfMu    sync.Mutex
	bfState map[string]*bruteForceEntry
}

type bruteForceEntry struct {
	failures    int
	lockedUntil time.Time
}

const (
	maxFailedAttempts = 5
	lockoutDuration   = 5 * time.Minute
)

func NewAuthHandler(agentService *services.AgentService, tokenService *services.TokenService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		agentService: agentService,
		tokenService: tokenService,
		cfg:          cfg,
		bfState:      make(map[string]*bruteForceEntry),
	}
}

// SetAdminService sets the admin service (called after construction to
// break the initialization ordering dependency).
func (h *AuthHandler) SetAdminService(adminService *services.AdminService) {
	h.adminService = adminService
}

// SetOAuthService sets the OAuth service (called after construction to
// break the initialization ordering dependency).
func (h *AuthHandler) SetOAuthService(oauthService *services.OAuthService) {
	h.oauthService = oauthService
}

// checkBruteForce returns true if the key is currently locked out.
func (h *AuthHandler) checkBruteForce(key string) bool {
	h.bfMu.Lock()
	defer h.bfMu.Unlock()
	entry, ok := h.bfState[key]
	if !ok {
		return false
	}
	if time.Now().Before(entry.lockedUntil) {
		return true
	}
	if time.Now().After(entry.lockedUntil) && entry.failures >= maxFailedAttempts {
		// Lockout expired — reset.
		delete(h.bfState, key)
	}
	return false
}

// recordFailure increments the failure counter for the given key.
func (h *AuthHandler) recordFailure(key string) {
	h.bfMu.Lock()
	defer h.bfMu.Unlock()
	entry, ok := h.bfState[key]
	if !ok {
		entry = &bruteForceEntry{}
		h.bfState[key] = entry
	}
	entry.failures++
	if entry.failures >= maxFailedAttempts {
		entry.lockedUntil = time.Now().Add(lockoutDuration)
		log.Printf("brute-force lockout: key=%s locked for %v", key, lockoutDuration)
	}
}

// recordSuccess clears the brute-force counter.
func (h *AuthHandler) recordSuccess(key string) {
	h.bfMu.Lock()
	defer h.bfMu.Unlock()
	delete(h.bfState, key)
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
		tokenReq.RefreshToken = r.Form.Get("refresh_token")
		tokenReq.Code = r.Form.Get("code")
		tokenReq.RedirectURI = r.Form.Get("redirect_uri")
		tokenReq.CodeVerifier = r.Form.Get("code_verifier")
	}

	if tokenReq.GrantType != "client_credentials" && tokenReq.GrantType != "refresh_token" && tokenReq.GrantType != "authorization_code" {
		h.writeError(w, "unsupported_grant_type", "grant_type must be client_credentials, refresh_token, or authorization_code")
		return
	}

	var agent *models.Agent

	if tokenReq.GrantType == "client_credentials" {
		if tokenReq.ClientID == "" || tokenReq.ClientSecret == "" {
			h.writeError(w, "invalid_request", "client_id and client_secret are required")
			return
		}

		// Brute-force protection.
		if h.checkBruteForce("client:" + tokenReq.ClientID) {
			w.Header().Set("Retry-After", "300")
			writeJSONError(w, http.StatusTooManyRequests, "too_many_requests", "too many failed attempts, try again later")
			return
		}

		agent, err = h.agentService.ValidateCredentials(tokenReq.ClientID, tokenReq.ClientSecret)
		if err != nil {
			h.recordFailure("client:" + tokenReq.ClientID)
			log.Printf("failed to validate credentials for client_id=%s: %v", tokenReq.ClientID, err)
			h.writeError(w, "invalid_client", "invalid client credentials")
			return
		}
		h.recordSuccess("client:" + tokenReq.ClientID)
	} else if tokenReq.GrantType == "refresh_token" {
		agent, err = h.tokenService.ValidateRefreshToken(tokenReq.RefreshToken)
		if err != nil {
			log.Printf("failed to validate refresh token: %v", err)
			h.writeError(w, "invalid_grant", "invalid or expired refresh token")
			return
		}

		h.agentService.RecordTokenRefresh(agent.ID)
	} else if tokenReq.GrantType == "authorization_code" {
		if h.oauthService == nil {
			h.writeError(w, "server_error", "OAuth service not configured")
			return
		}

		if tokenReq.Code == "" {
			h.writeError(w, "invalid_request", "code is required")
			return
		}

		if tokenReq.RedirectURI == "" {
			h.writeError(w, "invalid_request", "redirect_uri is required")
			return
		}

		code, err := h.oauthService.ValidateAuthorizationCode(tokenReq.Code, tokenReq.RedirectURI)
		if err != nil {
			log.Printf("failed to validate authorization code: %v", err)
			h.writeError(w, "invalid_grant", "invalid or expired authorization code")
			return
		}

		if code.CodeChallenge != "" {
			if err := h.oauthService.ValidateCodeVerifier(
				tokenReq.CodeVerifier,
				code.CodeChallenge,
				code.CodeChallengeMethod,
			); err != nil {
				log.Printf("failed to validate code verifier: %v", err)
				h.writeError(w, "invalid_grant", "invalid code_verifier")
				return
			}
		}

		if err := h.oauthService.ConsumeAuthorizationCode(code.ID); err != nil {
			log.Printf("failed to consume authorization code: %v", err)
			h.writeError(w, "server_error", "failed to consume authorization code")
			return
		}

		agent, err = h.agentService.GetByClientID(code.ClientID)
		if err != nil {
			log.Printf("failed to get agent: %v", err)
			h.writeError(w, "invalid_grant", "client not found")
			return
		}

		if !agent.IsActive {
			h.writeError(w, "invalid_grant", "client is inactive")
			return
		}

		user, err := h.adminService.GetUserByID(code.UserID)
		if err != nil {
			log.Printf("failed to get user: %v", err)
			h.writeError(w, "invalid_grant", "user not found")
			return
		}

		_ = user
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
	} else if tokenReq.GrantType == "refresh_token" {
		// Refresh token rotation: issue a new refresh token on each refresh.
		refreshToken, err := h.tokenService.GenerateRefreshToken(agent.ID.String())
		if err != nil {
			log.Printf("failed to rotate refresh token: %v", err)
		} else {
			tokenResp.RefreshToken = refreshToken
		}
	} else if tokenReq.GrantType == "authorization_code" {
		// Generate ID token for authorization_code grant
		if h.oauthService != nil {
			code, _ := h.oauthService.ValidateAuthorizationCode(tokenReq.Code, tokenReq.RedirectURI)
			if code != nil {
				user, _ := h.adminService.GetUserByID(code.UserID)
				if user != nil {
					idToken, err := h.oauthService.GenerateIDToken(user.ID, user.Email, code.OrganizationID, code.ClientID)
					if err == nil {
						tokenResp.IDToken = idToken
					}
				}
			}
		}
	}

	writeJSON(w, tokenResp)
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
		writeJSON(w, models.IntrospectResponse{Active: false})
		return
	}

	writeJSON(w, resp)
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

	writeJSON(w, map[string]string{"status": "revoked"})
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

	// Refresh token rotation: issue a new refresh token.
	newRefresh, err := h.tokenService.GenerateRefreshToken(agent.ID.String())
	if err != nil {
		log.Printf("failed to rotate refresh token: %v", err)
	} else {
		tokenResp.RefreshToken = newRefresh
	}

	writeJSON(w, tokenResp)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, errCode, description string) {
	writeJSONError(w, http.StatusBadRequest, errCode, description)
}

func (h *AuthHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	// Support both "email" and legacy "username" field.
	email := req.Email
	if email == "" {
		email = req.Username
	}

	if email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "email and password are required")
		return
	}

	// Brute-force protection.
	if h.checkBruteForce("admin:" + email) {
		w.Header().Set("Retry-After", "300")
		writeJSONError(w, http.StatusTooManyRequests, "too_many_requests", "too many failed attempts, try again later")
		return
	}

	// JWT-based admin auth.
	if h.adminService != nil {
		resp, err := h.adminService.Authenticate(email, req.Password)
		if err != nil {
			h.recordFailure("admin:" + email)
			writeJSONStatus(w, http.StatusUnauthorized, models.AdminTokenResponse{
				Success: false,
			})
			return
		}
		h.recordSuccess("admin:" + email)
		writeJSON(w, resp)
		return
	}

	// Legacy fallback (no admin service configured).
	if email == h.cfg.AdminEmail && req.Password == h.cfg.AdminPassword {
		h.recordSuccess("admin:" + email)
		writeJSON(w, models.AdminTokenResponse{
			Success: true,
		})
		return
	}

	h.recordFailure("admin:" + email)
	writeJSONStatus(w, http.StatusUnauthorized, models.AdminTokenResponse{
		Success: false,
	})
}

// Signup handles POST /api/auth/signup — self-service registration.
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.OrgName == "" || req.OrgSlug == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "email, password, org_name, and org_slug are required")
		return
	}

	if h.adminService == nil {
		writeJSONError(w, http.StatusInternalServerError, "server_error", "signup not available")
		return
	}

	resp, err := h.adminService.Signup(req)
	if err != nil {
		log.Printf("signup failed: %v", err)
		writeJSONError(w, http.StatusBadRequest, "signup_failed", err.Error())
		return
	}

	writeJSONStatus(w, http.StatusCreated, resp)
}

// SwitchOrg handles POST /api/auth/switch-org — switch organization context.
func (h *AuthHandler) SwitchOrg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adminID, ok := middleware.GetAdminIDFromContext(r.Context())
	if !ok || adminID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req struct {
		OrgID string `json:"org_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrgID == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "org_id is required")
		return
	}

	resp, err := h.adminService.IssueTokenForOrg(adminID, req.OrgID)
	if err != nil {
		log.Printf("switch org failed for user %s: %v", adminID, err)
		writeJSONError(w, http.StatusForbidden, "forbidden", err.Error())
		return
	}

	writeJSON(w, resp)
}

// GetMe returns the current admin user info and org memberships.
func (h *AuthHandler) AdminGetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adminID, ok := middleware.GetAdminIDFromContext(r.Context())
	if !ok || adminID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	user, err := h.adminService.GetUserByID(adminID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	memberships, _ := h.adminService.ListOrgMemberships(adminID)

	writeJSON(w, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
		"memberships": memberships,
	})
}
