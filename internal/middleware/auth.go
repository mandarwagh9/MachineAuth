package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"machineauth/internal/db"
	"machineauth/internal/services"
)

type contextKey string

const (
	AgentIDKey    contextKey = "agent_id"
	AdminIDKey    contextKey = "admin_id"
	AdminRoleKey  contextKey = "admin_role"
	AdminEmailKey contextKey = "admin_email"
	OrgIDKey      contextKey = "org_id"
	TeamIDKey     contextKey = "team_id"
	AuthMethodKey contextKey = "auth_method" // "admin_jwt" | "api_key"
)

func JWTAuth(tokenService *services.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			claims, err := tokenService.ValidateToken(tokenString)
			if err != nil {
				log.Printf("invalid token: %v", err)
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			agentIDStr, ok := tokenService.GetAgentIDFromClaims(claims)
			if !ok || agentIDStr == "" {
				http.Error(w, "invalid token claims: missing agent_id", http.StatusUnauthorized)
				return
			}

			agentID, err := uuid.Parse(agentIDStr)
			if err != nil {
				log.Printf("invalid agent ID in token: %v", err)
				http.Error(w, "invalid agent ID in token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AgentIDKey, agentID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminAuth is a middleware that validates admin JWT tokens issued by AdminService.
// It extracts admin_id, role, email, and org_id into request context.
func AdminAuth(adminService *services.AdminService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeUnauthorized(w, "invalid authorization header format")
				return
			}

			claims, err := adminService.ValidateAdminToken(parts[1])
			if err != nil {
				log.Printf("invalid admin token: %v", err)
				writeUnauthorized(w, "invalid or expired admin token")
				return
			}

			adminID, _ := claims["admin_id"].(string)
			role, _ := claims["role"].(string)
			email, _ := claims["sub"].(string)
			orgID, _ := claims["org_id"].(string)

			ctx := r.Context()
			ctx = context.WithValue(ctx, AdminIDKey, adminID)
			ctx = context.WithValue(ctx, AdminRoleKey, role)
			ctx = context.WithValue(ctx, AdminEmailKey, email)
			ctx = context.WithValue(ctx, OrgIDKey, orgID)
			ctx = context.WithValue(ctx, AuthMethodKey, "admin_jwt")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APIKeyAuth validates API keys (sk_... prefix) from Authorization header or
// X-API-Key header. On success it injects org_id, team_id, and auth_method
// into the request context. This is how external customers' backends authenticate.
func APIKeyAuth(database db.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractAPIKey(r)
			if key == "" {
				writeUnauthorized(w, "missing API key")
				return
			}

			// Look up by iterating all active keys and comparing with bcrypt.
			// For production at scale this should be optimised with a prefix lookup.
			apiKey, err := validateAPIKeyFromDB(database, key)
			if err != nil {
				log.Printf("invalid API key (prefix %s...): %v", safePrefix(key), err)
				writeUnauthorized(w, "invalid or expired API key")
				return
			}

			// Update last_used_at (fire-and-forget).
			go func() {
				now := time.Now()
				_ = database.UpdateAPIKey(apiKey.ID, func(k *db.APIKey) error {
					k.LastUsedAt = &now
					return nil
				})
			}()

			ctx := r.Context()
			ctx = context.WithValue(ctx, OrgIDKey, apiKey.OrganizationID)
			if apiKey.TeamID != nil {
				ctx = context.WithValue(ctx, TeamIDKey, *apiKey.TeamID)
			}
			ctx = context.WithValue(ctx, AuthMethodKey, "api_key")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOrAPIKeyAuth accepts either an admin JWT or an API key.
// Admin JWTs take precedence when both are present.
func AdminOrAPIKeyAuth(adminService *services.AdminService, database db.Database) func(http.Handler) http.Handler {
	adminMW := AdminAuth(adminService)
	apiKeyMW := APIKeyAuth(database)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractAPIKey(r)
			authHeader := r.Header.Get("Authorization")

			// If there's a Bearer token that's NOT an sk_ key, try admin JWT first.
			if authHeader != "" && !strings.HasPrefix(key, "sk_") {
				adminMW(next).ServeHTTP(w, r)
				return
			}

			// Otherwise try API key.
			if key != "" {
				apiKeyMW(next).ServeHTTP(w, r)
				return
			}

			writeUnauthorized(w, "missing authorization header or API key")
		})
	}
}

// extractAPIKey gets the API key from Authorization: Bearer sk_... or X-API-Key header.
func extractAPIKey(r *http.Request) string {
	// Check X-API-Key header first.
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	// Check Authorization: Bearer sk_...
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" && strings.HasPrefix(parts[1], "sk_") {
		return parts[1]
	}
	return ""
}

// validateAPIKeyFromDB looks up the key by checking the SHA-256 hash.
func validateAPIKeyFromDB(database db.Database, key string) (*db.APIKey, error) {
	keyHash := services.HashKey(key)
	apiKey, err := database.GetAPIKeyByKeyHash(keyHash)
	if err != nil {
		return nil, err
	}
	if !apiKey.IsActive {
		return nil, fmt.Errorf("API key is inactive")
	}
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}
	return apiKey, nil
}

func safePrefix(key string) string {
	if len(key) > 12 {
		return key[:12]
	}
	return key
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized","error_description":"` + msg + `"}`))
}

func GetAgentIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	agentID, ok := ctx.Value(AgentIDKey).(uuid.UUID)
	return agentID, ok
}

func GetAdminIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(AdminIDKey).(string)
	return id, ok
}

func GetAdminRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(AdminRoleKey).(string)
	return role, ok
}

func GetAdminEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(AdminEmailKey).(string)
	return email, ok
}

// GetOrgIDFromContext returns the organization ID from context.
// This is set by AdminAuth (from JWT claim) or APIKeyAuth (from key lookup).
func GetOrgIDFromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(OrgIDKey).(string)
	return orgID, ok
}

// GetTeamIDFromContext returns the team ID from context (set by API key auth).
func GetTeamIDFromContext(ctx context.Context) (string, bool) {
	teamID, ok := ctx.Value(TeamIDKey).(string)
	return teamID, ok
}

// GetAuthMethodFromContext returns the authentication method used ("admin_jwt" or "api_key").
func GetAuthMethodFromContext(ctx context.Context) (string, bool) {
	method, ok := ctx.Value(AuthMethodKey).(string)
	return method, ok
}

// RequireRole returns a middleware that checks the admin role is one of the allowed roles.
// Returns 403 Forbidden if the role doesn't match.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := GetAdminRoleFromContext(r.Context())
			if !ok {
				writeForbidden(w, "role not found in context")
				return
			}
			if !roleSet[role] {
				writeForbidden(w, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireOrgScope is middleware that ensures `org_id` is present in context.
// For super-admins (role=owner with no org_id), it allows global access.
func RequireOrgScope() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, hasOrg := GetOrgIDFromContext(r.Context())
			role, _ := GetAdminRoleFromContext(r.Context())

			// Super-admin (platform owner) can access everything.
			if role == "owner" && !hasOrg {
				next.ServeHTTP(w, r)
				return
			}
			if role == "owner" && orgID == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !hasOrg || orgID == "" {
				writeForbidden(w, "organization context required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeForbidden(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"forbidden","error_description":"` + msg + `"}`))
}
