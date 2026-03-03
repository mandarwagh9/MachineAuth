package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"machineauth/internal/config"
	"machineauth/internal/db"
	"machineauth/internal/handlers"
	"machineauth/internal/middleware"
	"machineauth/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	tokenService, err := services.NewTokenService(cfg, database)
	if err != nil {
		log.Fatalf("failed to create token service: %v", err)
	}

	agentService := services.NewAgentService(database)
	auditService := services.NewAuditService(database)
	orgService := services.NewOrganizationService(database)
	teamService := services.NewTeamService(database)
	apiKeyService := services.NewAPIKeyService(database)
	webhookService := services.NewWebhookService(database)

	// Admin service — uses the same RSA key for signing admin JWTs.
	adminService := services.NewAdminService(cfg, database, tokenService.PrivateKey(), tokenService.KeyID())
	if err := adminService.EnsureDefaultAdmin(); err != nil {
		log.Printf("warning: failed to ensure default admin: %v", err)
	}

	// Wire up webhook triggering in audit service
	auditService.SetWebhookService(webhookService)

	// Start webhook delivery worker
	webhookWorker := services.NewDeliveryWorker(webhookService, cfg.WebhookWorkerCount)
	webhookWorker.Start()
	defer webhookWorker.Stop()

	authHandler := handlers.NewAuthHandler(agentService, tokenService, cfg)
	authHandler.SetAdminService(adminService)
	agentsHandler := handlers.NewAgentsHandler(agentService, auditService)
	agentSelfHandler := handlers.NewAgentSelfHandler(agentService)
	orgHandler := handlers.NewOrganizationHandler(orgService, teamService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	webhookHandler := handlers.NewWebhookHandler(webhookService, auditService)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Rate limiters.
	oauthLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{Limit: 30, Window: 60 * time.Second})
	adminLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{Limit: 10, Window: 60 * time.Second})
	apiLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{Limit: 120, Window: 60 * time.Second})

	// Auth middleware.
	jwtAuth := middleware.JWTAuth(tokenService)
	adminAuth := middleware.AdminAuth(adminService)

	// Helper to apply admin auth + rate limit to a handler.
	adminProtected := func(h http.HandlerFunc) http.Handler {
		return middleware.RateLimit(apiLimiter)(adminAuth(http.HandlerFunc(h)))
	}

	mux := http.NewServeMux()

	// ── Public infrastructure endpoints ────────────────────────────────

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"endpoints":"/oauth/token, /oauth/introspect, /oauth/revoke, /oauth/refresh, /api/agents, /.well-known/jwks.json","service":"MachineAuth","status":"running","version":"2.12.61"}`))
	})

	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		agentCount, _ := agentService.Count()
		w.Write([]byte(fmt.Sprintf(`{"status":"ok","timestamp":"%s","agents_count":%d}`, time.Now().UTC().Format(time.RFC3339), agentCount)))
	})

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/.well-known/jwks.json", tokenService.JWKS)

	// ── OAuth endpoints (public, rate-limited) ─────────────────────────

	oauthRL := middleware.RateLimit(oauthLimiter)

	mux.Handle("/oauth/token", oauthRL(http.HandlerFunc(authHandler.Token)))
	mux.Handle("/oauth/introspect", oauthRL(http.HandlerFunc(authHandler.Introspect)))
	mux.Handle("/oauth/revoke", oauthRL(http.HandlerFunc(authHandler.Revoke)))
	mux.Handle("/oauth/refresh", oauthRL(http.HandlerFunc(authHandler.Refresh)))

	// ── Admin auth endpoint (public, aggressively rate-limited) ────────

	mux.Handle("/api/auth/login", middleware.RateLimit(adminLimiter)(http.HandlerFunc(authHandler.AdminLogin)))

	// ── Admin-protected CRUD endpoints ─────────────────────────────────

	mux.Handle("/api/agents", adminProtected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			agentsHandler.ListPaginated(w, r)
		case http.MethodPost:
			agentsHandler.Create(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/agents/", adminProtected(agentsHandler.HandleAgent))

	// Audit logs (admin-protected).
	mux.Handle("/api/audit-logs", adminProtected(auditHandler.ListAuditLogs))

	// Webhook routes (admin-protected).
	mux.Handle("/api/webhooks", adminProtected(webhookHandler.ListAndCreate))
	mux.Handle("/api/webhooks/", adminProtected(webhookHandler.HandleWebhook))
	mux.Handle("/api/webhook-events", adminProtected(webhookHandler.HandleEvents))

	// Organization routes (admin-protected).
	mux.Handle("/api/organizations", adminProtected(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			orgHandler.ListOrganizations(w, r)
		case http.MethodPost:
			orgHandler.CreateOrganization(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.Handle("/api/organizations/", adminProtected(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/api-keys") || strings.HasSuffix(path, "/api-keys/") {
			orgID := strings.TrimPrefix(path, "/api/organizations/")
			orgID = strings.TrimSuffix(orgID, "/api-keys")
			orgID = strings.TrimSuffix(orgID, "/")
			if orgID == "" {
				http.Error(w, "organization ID required", http.StatusBadRequest)
				return
			}
			switch r.Method {
			case http.MethodGet:
				apiKeyHandler.ListAPIKeys(w, r)
			case http.MethodPost:
				apiKeyHandler.CreateAPIKey(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.Contains(path, "/api-keys/") {
			switch r.Method {
			case http.MethodDelete:
				apiKeyHandler.DeleteAPIKey(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/teams") || strings.HasSuffix(path, "/teams/") {
			orgID := strings.TrimPrefix(path, "/api/organizations/")
			orgID = strings.TrimSuffix(orgID, "/teams")
			orgID = strings.TrimSuffix(orgID, "/")
			if orgID == "" {
				http.Error(w, "organization ID required", http.StatusBadRequest)
				return
			}
			switch r.Method {
			case http.MethodGet:
				orgHandler.ListTeams(w, r)
			case http.MethodPost:
				orgHandler.CreateTeam(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/agents") || strings.HasSuffix(path, "/agents/") {
			orgID := strings.TrimPrefix(path, "/api/organizations/")
			orgID = strings.TrimSuffix(orgID, "/agents")
			orgID = strings.TrimSuffix(orgID, "/")
			if orgID == "" {
				http.Error(w, "organization ID required", http.StatusBadRequest)
				return
			}
			switch r.Method {
			case http.MethodGet:
				agents, err := agentService.ListByOrganization(orgID)
				if err != nil {
					http.Error(w, "failed to list agents", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"agents": agents})
			case http.MethodPost:
				agentsHandler.CreateInOrganization(w, r, orgID)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			switch r.Method {
			case http.MethodGet:
				orgHandler.GetOrganization(w, r)
			case http.MethodPut, http.MethodPatch:
				orgHandler.UpdateOrganization(w, r)
			case http.MethodDelete:
				orgHandler.DeleteOrganization(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	}))

	// ── Agent self-service endpoints (agent JWT-protected) ─────────────

	mux.Handle("/api/agents/me", jwtAuth(http.HandlerFunc(agentSelfHandler.GetMe)))
	mux.Handle("/api/agents/me/usage", jwtAuth(http.HandlerFunc(agentSelfHandler.GetUsage)))
	mux.Handle("/api/agents/me/rotate", jwtAuth(http.HandlerFunc(agentSelfHandler.RotateCredentials)))
	mux.Handle("/api/agents/me/deactivate", jwtAuth(http.HandlerFunc(agentSelfHandler.Deactivate)))
	mux.Handle("/api/agents/me/reactivate", jwtAuth(http.HandlerFunc(agentSelfHandler.Reactivate)))
	mux.Handle("/api/agents/me/delete", jwtAuth(http.HandlerFunc(agentSelfHandler.Delete)))

	mux.Handle("/api/verify", jwtAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := middleware.GetAgentIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		agent, err := agentService.GetByID(agentID)
		if err != nil {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":       true,
			"agent_id":    agent.ID,
			"client_id":   agent.ClientID,
			"name":        agent.Name,
			"scopes":      agent.Scopes,
			"is_active":   agent.IsActive,
			"token_count": agent.TokenCount,
		})
	})))

	// ── Middleware chain ────────────────────────────────────────────────

	handler := middleware.SecurityHeaders(
		middleware.BodyLimit(1 << 20)(
			middleware.Logging(
				middleware.CORS(cfg.AllowedOrigins, mux),
			),
		),
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
