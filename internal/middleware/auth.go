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

const AgentIDKey contextKey = "agent_id"

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

func GetAgentIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	agentID, ok := ctx.Value(AgentIDKey).(uuid.UUID)
	return agentID, ok
}
