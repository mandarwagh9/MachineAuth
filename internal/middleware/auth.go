package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"machineauth/internal/services"
)

type contextKey string

const (
	AgentIDKey  contextKey = "agent_id"
	AdminIDKey  contextKey = "admin_id"
	AdminRoleKey contextKey = "admin_role"
	AdminEmailKey contextKey = "admin_email"
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
// It extracts admin_id, role, and email into request context.
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

			ctx := r.Context()
			ctx = context.WithValue(ctx, AdminIDKey, adminID)
			ctx = context.WithValue(ctx, AdminRoleKey, role)
			ctx = context.WithValue(ctx, AdminEmailKey, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
