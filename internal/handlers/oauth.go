package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"machineauth/internal/middleware"
	"machineauth/internal/services"
)

type OAuthHandler struct {
	oauthService *services.OAuthService
	agentService *services.AgentService
	adminService *services.AdminService
}

func NewOAuthHandler(oauthService *services.OAuthService, agentService *services.AgentService, adminService *services.AdminService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		agentService: agentService,
		adminService: adminService,
	}
}

func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")

	if responseType == "" {
		h.writeAuthorizeError(w, r, "invalid_request", "response_type is required", state)
		return
	}

	if responseType != "code" {
		h.writeAuthorizeError(w, r, "unsupported_response_type", "only response_type=code is supported", state)
		return
	}

	if clientID == "" {
		h.writeAuthorizeError(w, r, "invalid_request", "client_id is required", state)
		return
	}

	if redirectURI == "" {
		h.writeAuthorizeError(w, r, "invalid_request", "redirect_uri is required", state)
		return
	}

	if _, err := url.Parse(redirectURI); err != nil {
		h.writeAuthorizeError(w, r, "invalid_request", "invalid redirect_uri", state)
		return
	}

	agent, err := h.agentService.GetByClientID(clientID)
	if err != nil {
		h.writeAuthorizeError(w, r, "invalid_client", "unknown client_id", state)
		return
	}

	if !agent.IsActive {
		h.writeAuthorizeError(w, r, "invalid_client", "client is inactive", state)
		return
	}

	if agent.RedirectURIs == nil || len(agent.RedirectURIs) == 0 {
		h.writeAuthorizeError(w, r, "invalid_request", "client has no registered redirect URIs", state)
		return
	}

	validRedirect := false
	for _, uri := range agent.RedirectURIs {
		if uri == redirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		h.writeAuthorizeError(w, r, "invalid_request", "redirect_uri not registered", state)
		return
	}

	hasAuthCodeGrant := false
	if agent.GrantTypes != nil {
		for _, g := range agent.GrantTypes {
			if g == "authorization_code" {
				hasAuthCodeGrant = true
				break
			}
		}
	}
	if !hasAuthCodeGrant && len(agent.GrantTypes) > 0 {
		h.writeAuthorizeError(w, r, "unauthorized_client", "client not authorized for authorization_code grant", state)
		return
	}

	if codeChallenge != "" && codeChallengeMethod != "S256" {
		h.writeAuthorizeError(w, r, "invalid_request", "code_challenge_method must be S256", state)
		return
	}

	adminID, hasAdmin := middleware.GetAdminIDFromContext(r.Context())
	if !hasAdmin {
		redirectURL := "/api/auth/login?redirect_uri=" + url.QueryEscape(r.URL.String())
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	user, err := h.adminService.GetUserByID(adminID)
	if err != nil {
		h.writeAuthorizeError(w, r, "server_error", "failed to get user", state)
		return
	}

	authCode, err := h.oauthService.CreateAuthorizationCode(
		clientID,
		user.ID,
		agent.OrganizationID,
		redirectURI,
		scope,
		codeChallenge,
		codeChallengeMethod,
	)
	if err != nil {
		h.writeAuthorizeError(w, r, "server_error", "failed to create authorization code", state)
		return
	}

	resp := url.Values{}
	resp.Set("code", authCode.Code)
	if state != "" {
		resp.Set("state", state)
	}

	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		h.writeAuthorizeError(w, r, "server_error", "failed to parse redirect_uri", state)
		return
	}
	redirectURL.RawQuery = resp.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

func (h *OAuthHandler) writeAuthorizeError(w http.ResponseWriter, r *http.Request, errCode, description, state string) {
	resp := url.Values{}
	resp.Set("error", errCode)
	resp.Set("error_description", description)
	if state != "" {
		resp.Set("state", state)
	}

	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		w.Header().Set("Content-Type", "application/json")
		writeJSONError(w, http.StatusBadRequest, errCode, description)
		return
	}

	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		writeJSONError(w, http.StatusBadRequest, errCode, description)
		return
	}
	redirectURL.RawQuery = resp.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

func (h *OAuthHandler) Consent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clientID := r.URL.Query().Get("client_id")
	scope := r.URL.Query().Get("scope")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")

	if r.Method == http.MethodPost {
		consent := r.FormValue("consent")
		if consent != "granted" {
			errResp := url.Values{}
			errResp.Set("error", "access_denied")
			errResp.Set("error_description", "user denied access")
			if state != "" {
				errResp.Set("state", state)
			}
			redirectURL, _ := url.Parse(redirectURI)
			redirectURL.RawQuery = errResp.Encode()
			http.Redirect(w, r, redirectURL.String(), http.StatusFound)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Authorization Request</title></head>
<body>
<h2>Authorize Application</h2>
<p>The application is requesting access to your account.</p>
<form method="POST">
<input type="hidden" name="client_id" value="` + htmlEscape(clientID) + `">
<input type="hidden" name="scope" value="` + htmlEscape(scope) + `">
<input type="hidden" name="redirect_uri" value="` + htmlEscape(redirectURI) + `">
<input type="hidden" name="state" value="` + htmlEscape(state) + `">
<button type="submit" name="consent" value="granted">Allow</button>
<button type="submit" name="consent" value="denied">Deny</button>
</form>
</body>
</html>`))
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
