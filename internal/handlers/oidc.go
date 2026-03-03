package handlers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"machineauth/internal/config"
	"machineauth/internal/middleware"
	"machineauth/internal/models"
	"machineauth/internal/services"
)

// OIDCHandler serves OpenID Connect discovery, JWKS, and UserInfo endpoints.
type OIDCHandler struct {
	cfg        *config.Config
	orgService *services.OrganizationService
	agentService *services.AgentService
	tokenService *services.TokenService
}

// NewOIDCHandler creates a new OIDC handler.
func NewOIDCHandler(cfg *config.Config, orgService *services.OrganizationService, agentService *services.AgentService, tokenService *services.TokenService) *OIDCHandler {
	return &OIDCHandler{
		cfg:        cfg,
		orgService: orgService,
		agentService: agentService,
		tokenService: tokenService,
	}
}

// Discovery handles GET /.well-known/openid-configuration — global OIDC discovery.
func (h *OIDCHandler) Discovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// OIDC discovery and JWKS are public endpoints — allow any origin and
	// don't restrict by User-Agent so SDK clients (e.g. PyJWKClient) can
	// fetch keys without being blocked by intermediary firewalls.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	baseURL := strings.TrimRight(h.cfg.BaseURL, "/")

	doc := models.OIDCDiscovery{
		Issuer:                            h.cfg.JWTIssuer,
		TokenEndpoint:                     baseURL + "/oauth/token",
		UserInfoEndpoint:                  baseURL + "/oauth/userinfo",
		JWKSUri:                           baseURL + "/.well-known/jwks.json",
		IntrospectionEndpoint:             baseURL + "/oauth/introspect",
		RevocationEndpoint:                baseURL + "/oauth/revoke",
		GrantTypesSupported:               []string{"client_credentials", "refresh_token"},
		ResponseTypesSupported:            []string{"token"},
		ScopesSupported:                   []string{"openid", "read", "write", "admin", "agent:read", "agent:write", "agent:delete", "token:issue", "token:revoke", "webhook:read", "webhook:write"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "client_secret_basic"},
		ClaimsSupported:                   []string{"sub", "iss", "aud", "exp", "iat", "jti", "scope", "agent_id", "org_id", "team_id", "name", "auth_time", "at_hash"},
		ServiceDocumentation:              baseURL + "/docs",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(doc)
}

// UserInfo handles GET /oauth/userinfo — returns agent profile for the bearer token.
func (h *OIDCHandler) UserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentIDFromContext(r.Context())
	if !ok {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="missing or invalid access token"`)
		http.Error(w, `{"error":"invalid_token"}`, http.StatusUnauthorized)
		return
	}

	agent, err := h.agentService.GetByID(agentID)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="agent not found"`)
		http.Error(w, `{"error":"invalid_token"}`, http.StatusUnauthorized)
		return
	}

	var teamIDStr string
	if agent.TeamID != nil {
		teamIDStr = agent.TeamID.String()
	}

	resp := models.UserInfoResponse{
		Sub:       agent.ClientID,
		Name:      agent.Name,
		AgentID:   agent.ID.String(),
		OrgID:     agent.OrganizationID,
		TeamID:    teamIDStr,
		Scopes:    agent.Scopes,
		Status:    string(agent.Status),
		Metadata:  agent.Metadata,
		CreatedAt: agent.CreatedAt.Unix(),
		UpdatedAt: agent.UpdatedAt.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// OrgDiscovery handles GET /.well-known/openid-configuration for a specific org.
// The org slug is extracted from the Host header (subdomain-based routing).
func (h *OIDCHandler) OrgDiscovery(w http.ResponseWriter, r *http.Request, orgSlug string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	org, err := h.orgService.GetBySlug(orgSlug)
	if err != nil {
		http.Error(w, `{"error":"organization not found"}`, http.StatusNotFound)
		return
	}

	// Construct per-org base URL from the subdomain pattern.
	orgBaseURL := h.orgBaseURL(orgSlug)

	// Use the org's configured JWT issuer, or fall back to the org base URL.
	issuer := org.JWTIssuer
	if issuer == "" || issuer == "https://auth.agentauth.io" {
		issuer = orgBaseURL
	}

	doc := models.OIDCDiscovery{
		Issuer:                            issuer,
		TokenEndpoint:                     orgBaseURL + "/oauth/token",
		UserInfoEndpoint:                  orgBaseURL + "/oauth/userinfo",
		JWKSUri:                           orgBaseURL + "/.well-known/jwks.json",
		IntrospectionEndpoint:             orgBaseURL + "/oauth/introspect",
		RevocationEndpoint:                orgBaseURL + "/oauth/revoke",
		GrantTypesSupported:               []string{"client_credentials", "refresh_token"},
		ResponseTypesSupported:            []string{"token"},
		ScopesSupported:                   []string{"openid", "read", "write", "admin", "agent:read", "agent:write", "agent:delete", "token:issue", "token:revoke", "webhook:read", "webhook:write"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "client_secret_basic"},
		ClaimsSupported:                   []string{"sub", "iss", "aud", "exp", "iat", "jti", "scope", "agent_id", "org_id", "team_id", "name", "auth_time", "at_hash"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(doc)
}

// OrgJWKS handles GET /.well-known/jwks.json for a specific org.
// Returns only that org's signing keys (not the global key).
func (h *OIDCHandler) OrgJWKS(w http.ResponseWriter, r *http.Request, orgSlug string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	org, err := h.orgService.GetBySlug(orgSlug)
	if err != nil {
		http.Error(w, `{"error":"organization not found"}`, http.StatusNotFound)
		return
	}

	keys, err := h.tokenService.GetOrgJWKS(org.ID)
	if err != nil {
		http.Error(w, `{"error":"failed to retrieve keys"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(keys)
}

// HandleWellKnown is the catch-all handler for /.well-known/ paths.
// It dispatches to either global or per-org endpoints based on the Host header.
func (h *OIDCHandler) HandleWellKnown(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Check if this is a per-org subdomain request.
	orgSlug := h.extractOrgSlug(r)

	if orgSlug != "" {
		// Per-org OIDC via subdomain.
		switch {
		case path == "/.well-known/openid-configuration":
			h.OrgDiscovery(w, r, orgSlug)
		case path == "/.well-known/jwks.json":
			h.OrgJWKS(w, r, orgSlug)
		default:
			http.NotFound(w, r)
		}
		return
	}

	// Global endpoints.
	switch {
	case path == "/.well-known/openid-configuration":
		h.Discovery(w, r)
	case path == "/.well-known/jwks.json":
		h.tokenService.JWKS(w, r)
	default:
		http.NotFound(w, r)
	}
}

// extractOrgSlug extracts the org slug from a subdomain-based Host header.
// e.g., "acme.auth.writesomething.fun" → "acme"
// Returns "" if the Host is the base domain (no org subdomain).
func (h *OIDCHandler) extractOrgSlug(r *http.Request) string {
	host := r.Host
	// Strip port if present.
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Extract the base domain from BASE_URL.
	baseDomain := h.baseDomain()
	if baseDomain == "" {
		return ""
	}

	// Check if host is a subdomain of the base domain.
	// e.g., host = "acme.auth.writesomething.fun", baseDomain = "auth.writesomething.fun"
	if !strings.HasSuffix(host, "."+baseDomain) {
		return ""
	}

	// Extract the subdomain prefix.
	slug := strings.TrimSuffix(host, "."+baseDomain)

	// Ensure it's a single-level subdomain (no dots).
	if strings.Contains(slug, ".") {
		return ""
	}

	return slug
}

// baseDomain extracts the domain from BASE_URL for subdomain detection.
func (h *OIDCHandler) baseDomain() string {
	baseURL := h.cfg.BaseURL
	// Strip protocol.
	baseURL = strings.TrimPrefix(baseURL, "https://")
	baseURL = strings.TrimPrefix(baseURL, "http://")
	// Strip path.
	if idx := strings.Index(baseURL, "/"); idx != -1 {
		baseURL = baseURL[:idx]
	}
	// Strip port.
	if idx := strings.LastIndex(baseURL, ":"); idx != -1 {
		baseURL = baseURL[:idx]
	}
	return baseURL
}

// orgBaseURL constructs the per-org base URL from the slug.
func (h *OIDCHandler) orgBaseURL(slug string) string {
	domain := h.baseDomain()
	scheme := "https"
	if strings.HasPrefix(h.cfg.BaseURL, "http://") {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s.%s", scheme, slug, domain)
}

// OrgJWKSFromPEM builds a JWKS response from PEM-encoded public keys.
func OrgJWKSFromPEM(keys []models.OrgSigningKey) models.JWKS {
	jwks := models.JWKS{Keys: make([]models.JWK, 0, len(keys))}

	for _, k := range keys {
		if !k.IsActive {
			continue
		}

		block, _ := pem.Decode([]byte(k.PublicKeyPEM))
		if block == nil {
			continue
		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			continue
		}

		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			continue
		}

		jwks.Keys = append(jwks.Keys, models.JWK{
			Kty: "RSA",
			Kid: k.KeyID,
			Use: "sig",
			Alg: k.Algorithm,
			N:   base64.RawURLEncoding.EncodeToString(rsaPub.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsaPub.E)).Bytes()),
		})
	}

	return jwks
}
